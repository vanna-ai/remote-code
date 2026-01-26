package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"remote-code/db"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Session duration
const sessionDuration = 24 * time.Hour

// In-memory storage for WebAuthn challenges (short-lived)
var (
	challengeStore     = make(map[string]*webauthn.SessionData)
	challengeStoreMu   sync.RWMutex
	challengeStoreTTL  = 5 * time.Minute
)

// WebAuthnUser implements webauthn.User interface for a single-user system
type WebAuthnUser struct {
	credentials []webauthn.Credential
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	return []byte("single-user")
}

func (u *WebAuthnUser) WebAuthnName() string {
	return "admin"
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return "Administrator"
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

func (u *WebAuthnUser) WebAuthnIcon() string {
	return ""
}

// getWebAuthn creates a WebAuthn configuration with dynamic origin handling
func getWebAuthn(r *http.Request) (*webauthn.WebAuthn, error) {
	origin := getOriginFromRequest(r)
	rpID := getRPID(r)

	cfg := &webauthn.Config{
		RPDisplayName: "Web Terminal",
		RPID:          rpID,
		RPOrigins:     []string{origin},
	}

	return webauthn.New(cfg)
}

// getOriginFromRequest extracts the origin from the request
func getOriginFromRequest(r *http.Request) string {
	// Check X-Forwarded-Proto and X-Forwarded-Host headers first (for proxies)
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}

	return fmt.Sprintf("%s://%s", proto, host)
}

// getRPID determines the Relying Party ID based on the host
func getRPID(r *http.Request) string {
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}

	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Note: trycloudflare.com is on the Public Suffix List, so we cannot use
	// it as an RP ID. Each subdomain must use its full hostname.
	// This means credentials are specific to each tunnel subdomain.

	return host
}

// loadUserCredentials loads WebAuthn credentials from the database for a specific RP ID
func loadUserCredentials(ctx context.Context, rpID string) (*WebAuthnUser, error) {
	creds, err := queries.ListWebAuthnCredentialsByRpID(ctx, rpID)
	if err != nil {
		return nil, err
	}

	user := &WebAuthnUser{
		credentials: make([]webauthn.Credential, len(creds)),
	}

	for i, c := range creds {
		var transports []protocol.AuthenticatorTransport
		if c.Transport.Valid && c.Transport.String != "" {
			for _, t := range strings.Split(c.Transport.String, ",") {
				transports = append(transports, protocol.AuthenticatorTransport(t))
			}
		}

		user.credentials[i] = webauthn.Credential{
			ID:              []byte(c.ID),
			PublicKey:       c.PublicKey,
			AttestationType: c.AttestationType,
			Transport:       transports,
			Authenticator: webauthn.Authenticator{
				AAGUID:    c.Aaguid,
				SignCount: uint32(c.SignCount),
			},
		}
	}

	return user, nil
}

// generateSessionToken creates a cryptographically secure session token
func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// validateSession checks if a session token is valid
func validateSession(ctx context.Context, token string) bool {
	_, err := queries.GetSession(ctx, token)
	return err == nil
}

// handleAuthAPI routes auth-related requests
func handleAuthAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	if len(pathParts) == 0 {
		http.Error(w, "Invalid auth path", http.StatusBadRequest)
		return
	}

	switch pathParts[0] {
	case "status":
		handleAuthStatus(w, r, ctx)
	case "register":
		if len(pathParts) < 2 {
			http.Error(w, "Invalid register path", http.StatusBadRequest)
			return
		}
		switch pathParts[1] {
		case "begin":
			handleRegisterBegin(w, r, ctx)
		case "finish":
			handleRegisterFinish(w, r, ctx)
		default:
			http.Error(w, "Unknown register endpoint", http.StatusNotFound)
		}
	case "login":
		if len(pathParts) < 2 {
			http.Error(w, "Invalid login path", http.StatusBadRequest)
			return
		}
		switch pathParts[1] {
		case "begin":
			handleLoginBegin(w, r, ctx)
		case "finish":
			handleLoginFinish(w, r, ctx)
		default:
			http.Error(w, "Unknown login endpoint", http.StatusNotFound)
		}
	case "logout":
		handleLogout(w, r, ctx)
	default:
		http.Error(w, "Unknown auth endpoint", http.StatusNotFound)
	}
}

// AuthStatus represents the authentication status response
type AuthStatus struct {
	Authenticated  bool `json:"authenticated"`
	HasCredentials bool `json:"hasCredentials"`
}

// handleAuthStatus returns the current authentication status
func handleAuthStatus(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := AuthStatus{}
	rpID := getRPID(r)

	// Check if user has credentials for this RP ID (domain)
	count, err := queries.CountWebAuthnCredentialsByRpID(ctx, rpID)
	if err != nil {
		log.Printf("Error counting credentials: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	status.HasCredentials = count > 0

	// Check if user is authenticated
	cookie, err := r.Cookie("session")
	if err == nil && cookie.Value != "" {
		status.Authenticated = validateSession(ctx, cookie.Value)
	}

	json.NewEncoder(w).Encode(status)
}

// handleRegisterBegin starts the WebAuthn registration process
func handleRegisterBegin(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	webAuthn, err := getWebAuthn(r)
	if err != nil {
		log.Printf("Error creating WebAuthn config: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Load existing credentials to exclude them
	rpID := getRPID(r)
	user, err := loadUserCredentials(ctx, rpID)
	if err != nil {
		log.Printf("Error loading credentials: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create registration options
	options, session, err := webAuthn.BeginRegistration(user)
	if err != nil {
		log.Printf("Error beginning registration: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store the session data for verification
	challengeID := base64.RawURLEncoding.EncodeToString(options.Response.Challenge)
	challengeStoreMu.Lock()
	challengeStore[challengeID] = session
	challengeStoreMu.Unlock()

	// Clean up old challenges
	go cleanupOldChallenges()

	json.NewEncoder(w).Encode(options)
}

// handleRegisterFinish completes the WebAuthn registration process
func handleRegisterFinish(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	webAuthn, err := getWebAuthn(r)
	if err != nil {
		log.Printf("Error creating WebAuthn config: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Parse the credential creation response
	response, err := protocol.ParseCredentialCreationResponseBody(r.Body)
	if err != nil {
		log.Printf("Error parsing credential creation response: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find the session data
	challengeID := response.Response.CollectedClientData.Challenge
	challengeStoreMu.RLock()
	session, ok := challengeStore[challengeID]
	challengeStoreMu.RUnlock()

	if !ok {
		http.Error(w, "Challenge not found or expired", http.StatusBadRequest)
		return
	}

	// Load user to verify against
	rpID := getRPID(r)
	user, err := loadUserCredentials(ctx, rpID)
	if err != nil {
		log.Printf("Error loading credentials: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Verify the registration
	credential, err := webAuthn.CreateCredential(user, *session, response)
	if err != nil {
		log.Printf("Error creating credential: %v", err)
		http.Error(w, "Registration verification failed", http.StatusBadRequest)
		return
	}

	// Store the credential in the database
	var transports []string
	for _, t := range credential.Transport {
		transports = append(transports, string(t))
	}

	_, err = queries.CreateWebAuthnCredential(ctx, db.CreateWebAuthnCredentialParams{
		ID:              string(credential.ID),
		RpID:            rpID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Transport: sql.NullString{
			String: strings.Join(transports, ","),
			Valid:  len(transports) > 0,
		},
		Aaguid:    credential.Authenticator.AAGUID,
		SignCount: int64(credential.Authenticator.SignCount),
	})

	if err != nil {
		log.Printf("Error storing credential: %v", err)
		http.Error(w, "Failed to store credential", http.StatusInternalServerError)
		return
	}

	// Clean up the challenge
	challengeStoreMu.Lock()
	delete(challengeStore, challengeID)
	challengeStoreMu.Unlock()

	// Create a session for the newly registered user
	token, err := generateSessionToken()
	if err != nil {
		log.Printf("Error generating session token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = queries.CreateSession(ctx, db.CreateSessionParams{
		Token:     token,
		ExpiresAt: time.Now().Add(sessionDuration),
	})
	if err != nil {
		log.Printf("Error creating session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(sessionDuration.Seconds()),
	})

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleLoginBegin starts the WebAuthn authentication process
func handleLoginBegin(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	webAuthn, err := getWebAuthn(r)
	if err != nil {
		log.Printf("Error creating WebAuthn config: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Load user credentials
	rpID := getRPID(r)
	user, err := loadUserCredentials(ctx, rpID)
	if err != nil {
		log.Printf("Error loading credentials: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(user.credentials) == 0 {
		http.Error(w, "No credentials registered", http.StatusBadRequest)
		return
	}

	// Create login options
	options, session, err := webAuthn.BeginLogin(user)
	if err != nil {
		log.Printf("Error beginning login: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store the session data for verification
	challengeID := base64.RawURLEncoding.EncodeToString(options.Response.Challenge)
	challengeStoreMu.Lock()
	challengeStore[challengeID] = session
	challengeStoreMu.Unlock()

	// Clean up old challenges
	go cleanupOldChallenges()

	json.NewEncoder(w).Encode(options)
}

// handleLoginFinish completes the WebAuthn authentication process
func handleLoginFinish(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	webAuthn, err := getWebAuthn(r)
	if err != nil {
		log.Printf("Error creating WebAuthn config: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Parse the credential assertion response
	response, err := protocol.ParseCredentialRequestResponseBody(r.Body)
	if err != nil {
		log.Printf("Error parsing credential request response: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find the session data
	challengeID := response.Response.CollectedClientData.Challenge
	challengeStoreMu.RLock()
	session, ok := challengeStore[challengeID]
	challengeStoreMu.RUnlock()

	if !ok {
		http.Error(w, "Challenge not found or expired", http.StatusBadRequest)
		return
	}

	// Load user to verify against
	rpID := getRPID(r)
	user, err := loadUserCredentials(ctx, rpID)
	if err != nil {
		log.Printf("Error loading credentials: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Verify the assertion
	credential, err := webAuthn.ValidateLogin(user, *session, response)
	if err != nil {
		log.Printf("Error validating login: %v", err)
		http.Error(w, "Login verification failed", http.StatusUnauthorized)
		return
	}

	// Update the sign count in the database
	err = queries.UpdateWebAuthnCredentialSignCount(ctx, db.UpdateWebAuthnCredentialSignCountParams{
		SignCount: int64(credential.Authenticator.SignCount),
		ID:        string(credential.ID),
	})
	if err != nil {
		log.Printf("Error updating sign count: %v", err)
		// Don't fail the login for this
	}

	// Clean up the challenge
	challengeStoreMu.Lock()
	delete(challengeStore, challengeID)
	challengeStoreMu.Unlock()

	// Create a new session
	token, err := generateSessionToken()
	if err != nil {
		log.Printf("Error generating session token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = queries.CreateSession(ctx, db.CreateSessionParams{
		Token:     token,
		ExpiresAt: time.Now().Add(sessionDuration),
	})
	if err != nil {
		log.Printf("Error creating session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(sessionDuration.Seconds()),
	})

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleLogout invalidates the current session
func handleLogout(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session")
	if err == nil && cookie.Value != "" {
		queries.DeleteSession(ctx, cookie.Value)
	}

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// cleanupOldChallenges removes expired challenges from memory
func cleanupOldChallenges() {
	// This is a simple cleanup - in production, use proper expiry tracking
	challengeStoreMu.Lock()
	defer challengeStoreMu.Unlock()

	// Simple approach: if store gets too large, clear it
	// Challenges are short-lived anyway (5 minutes)
	if len(challengeStore) > 100 {
		challengeStore = make(map[string]*webauthn.SessionData)
	}
}

// authMiddleware checks for valid session before allowing access
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		rpID := getRPID(r)

		// Check if any credentials are registered for this RP ID (domain)
		// If not, skip auth (allow first-time setup for this domain)
		count, err := queries.CountWebAuthnCredentialsByRpID(ctx, rpID)
		if err != nil {
			log.Printf("Error counting credentials in middleware: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// If no credentials registered for this domain, allow access (setup mode)
		if count == 0 {
			next(w, r)
			return
		}

		// Check for valid session
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !validateSession(ctx, cookie.Value) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// getCORSOrigin returns the appropriate CORS origin for the request
func getCORSOrigin(r *http.Request) string {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return "*"
	}

	// Parse the origin to check if it's a trycloudflare.com subdomain
	if strings.Contains(origin, ".trycloudflare.com") ||
	   strings.Contains(origin, "localhost") ||
	   strings.Contains(origin, "127.0.0.1") {
		return origin
	}

	return "*"
}

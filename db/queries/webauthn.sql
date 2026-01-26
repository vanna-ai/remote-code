-- name: CreateWebAuthnCredential :one
INSERT INTO webauthn_credentials (id, rp_id, public_key, attestation_type, transport, aaguid, sign_count)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetWebAuthnCredential :one
SELECT * FROM webauthn_credentials
WHERE id = ?;

-- name: ListWebAuthnCredentials :many
SELECT * FROM webauthn_credentials
ORDER BY created_at;

-- name: ListWebAuthnCredentialsByRpID :many
SELECT * FROM webauthn_credentials
WHERE rp_id = ?
ORDER BY created_at;

-- name: UpdateWebAuthnCredentialSignCount :exec
UPDATE webauthn_credentials
SET sign_count = ?
WHERE id = ?;

-- name: DeleteWebAuthnCredential :exec
DELETE FROM webauthn_credentials WHERE id = ?;

-- name: CountWebAuthnCredentials :one
SELECT COUNT(*) FROM webauthn_credentials;

-- name: CountWebAuthnCredentialsByRpID :one
SELECT COUNT(*) FROM webauthn_credentials
WHERE rp_id = ?;

-- name: CreateSession :one
INSERT INTO sessions (token, expires_at)
VALUES (?, ?)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE token = ? AND expires_at > CURRENT_TIMESTAMP;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at <= CURRENT_TIMESTAMP;

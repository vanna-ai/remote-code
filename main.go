package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"remote-code/db"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var database *sql.DB
var queries *db.Queries

func main() {
	// Initialize database
	database, queries = initDatabase()
	defer database.Close()

	// Setup HTTP routes
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/", handleAPI)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	// API routes should be handled by the API handler
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	// Check if the requested file exists in static directory
	filePath := "static" + r.URL.Path
	if r.URL.Path == "/" {
		filePath = "static/index.html"
	}

	// Check if file exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, serve it
		http.ServeFile(w, r, filePath)
	} else {
		// File doesn't exist, serve index.html for SPA routing
		http.ServeFile(w, r, "static/index.html")
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Check if a specific session is requested
	sessionName := r.URL.Query().Get("session")
	
	var cmd *exec.Cmd
	if sessionName != "" {
		// Attach to specific tmux session
		log.Printf("Attaching to tmux session: %s", sessionName)
		cmd = exec.Command("tmux", "attach-session", "-t", sessionName)
	} else {
		// Create or attach to global session for general terminal use
		log.Printf("Creating/attaching to global terminal session")
		cmd = exec.Command("tmux", "new-session", "-A", "-s", "remote-code")
	}

	// Ensure proper terminal environment for UTF-8 and colors
	cmd.Env = append(os.Environ(),
		"LANG=C.UTF-8",
		"LC_ALL=C.UTF-8",
		"TERM=xterm-256color",
	)
 
 	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Printf("Failed to start pty: %v", err)
		return
	}
	defer ptmx.Close()

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				break
			}

			// Try to parse control messages (e.g., resize)
			type resizeMsg struct {
				Type string `json:"type"`
				Cols int    `json:"cols"`
				Rows int    `json:"rows"`
			}
			var rm resizeMsg
			if err := json.Unmarshal(message, &rm); err == nil && rm.Type == "resize" {
				if rm.Cols > 0 && rm.Rows > 0 {
					_ = pty.Setsize(ptmx, &pty.Winsize{Cols: uint16(rm.Cols), Rows: uint16(rm.Rows)})
				}
				continue
			}

			ptmx.Write(message)
		}
	}()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("PTY read error: %v", err)
				}
				break
			}
			// Send raw bytes as a binary WebSocket frame; the browser will decode UTF-8 progressively
			if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				log.Printf("WebSocket write error: %v", err)
				break
			}
		}
	}()

	cmd.Wait()
}
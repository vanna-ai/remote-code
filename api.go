package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"remote-code/db"
)

func handleAPI(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse the path
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	pathParts := strings.Split(path, "/")

	if len(pathParts) == 0 {
		http.Error(w, "Invalid API path", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	switch pathParts[0] {
	case "roots":
		handleRootsAPI(w, r, ctx, pathParts[1:])
	case "projects":
		handleProjectsAPI(w, r, ctx, pathParts[1:])
	case "agents":
		handleAgentsAPI(w, r, ctx, pathParts[1:])
	case "base-directories":
		handleBaseDirectoriesAPI(w, r, ctx, pathParts[1:])
	case "worktrees":
		handleWorktreesAPI(w, r, ctx, pathParts[1:])
	case "tasks":
		handleTasksAPI(w, r, ctx, pathParts[1:])
	default:
		http.Error(w, "Unknown API endpoint", http.StatusNotFound)
	}
}

func handleRootsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	switch r.Method {
	case "GET":
		if len(pathParts) == 0 {
			// List all roots - for now just return empty array
			json.NewEncoder(w).Encode([]Root{})
		} else {
			// Get specific root
			id, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid root ID", http.StatusBadRequest)
				return
			}

			root, err := queries.GetRoot(ctx, id)
			if err != nil {
				http.Error(w, "Root not found", http.StatusNotFound)
				return
			}

			// Get associated agents
			agents, err := queries.GetAgentsByRootID(ctx, id)
			if err != nil {
				http.Error(w, "Failed to get agents", http.StatusInternalServerError)
				return
			}

			// Get associated projects (simplified for now)
			projects := []Project{}

			result := dbRootToRoot(root, agents, projects)
			json.NewEncoder(w).Encode(result)
		}

	case "POST":
		var createReq struct {
			LocalPort   string  `json:"local_port"`
			ExternalUrl *string `json:"external_url"`
		}

		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		root, err := queries.CreateRoot(ctx, db.CreateRootParams{
			LocalPort:   createReq.LocalPort,
			ExternalUrl: stringToNullString(createReq.ExternalUrl),
		})
		if err != nil {
			http.Error(w, "Failed to create root", http.StatusInternalServerError)
			return
		}

		result := dbRootToRoot(root, []db.Agent{}, []Project{})
		json.NewEncoder(w).Encode(result)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleProjectsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	switch r.Method {
	case "GET":
		if len(pathParts) == 0 {
			// List all projects - for now return empty
			json.NewEncoder(w).Encode([]Project{})
		} else {
			// Get specific project
			id, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid project ID", http.StatusBadRequest)
				return
			}

			project, err := queries.GetProject(ctx, id)
			if err != nil {
				http.Error(w, "Project not found", http.StatusNotFound)
				return
			}

			// Get base directories
			dbBaseDirs, err := queries.GetBaseDirectoriesByProjectID(ctx, id)
			if err != nil {
				http.Error(w, "Failed to get base directories", http.StatusInternalServerError)
				return
			}

			var baseDirs []BaseDirectory
			for _, dbBaseDir := range dbBaseDirs {
				baseDirs = append(baseDirs, dbBaseDirectoryToBaseDirectory(dbBaseDir))
			}

			// Get tasks (simplified for now)
			tasks := []Task{}

			result := Project{
				Name:            project.Name,
				BaseDirectories: baseDirs,
				Tasks:           tasks,
			}
			json.NewEncoder(w).Encode(result)
		}

	case "POST":
		var createReq struct {
			RootId int64  `json:"root_id"`
			Name   string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		project, err := queries.CreateProject(ctx, db.CreateProjectParams{
			RootID: createReq.RootId,
			Name:   createReq.Name,
		})
		if err != nil {
			http.Error(w, "Failed to create project", http.StatusInternalServerError)
			return
		}

		result := Project{
			Name:            project.Name,
			BaseDirectories: []BaseDirectory{},
			Tasks:           []Task{},
		}
		json.NewEncoder(w).Encode(result)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Placeholder handlers for other endpoints
func handleAgentsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Agents API not implemented yet"})
}

func handleBaseDirectoriesAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Base directories API not implemented yet"})
}

func handleWorktreesAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Worktrees API not implemented yet"})
}

func handleTasksAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Tasks API not implemented yet"})
}
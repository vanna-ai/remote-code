package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	// Handle tasks sub-resource: /api/projects/{id}/tasks
	if len(pathParts) >= 2 && pathParts[1] == "tasks" {
		handleProjectTasksAPI(w, r, ctx, pathParts)
		return
	}
	
	// Handle base-directories sub-resource: /api/projects/{id}/base-directories
	if len(pathParts) >= 2 && pathParts[1] == "base-directories" {
		handleProjectBaseDirectoriesAPI(w, r, ctx, pathParts)
		return
	}
	
	switch r.Method {
	case "GET":
		if len(pathParts) == 0 {
			// List all projects
			dbProjects, err := queries.GetProjectsByRootID(ctx, 1) // Use default root_id for now
			if err != nil {
				// If root doesn't exist, return empty array instead of error
				log.Printf("No projects found for root_id 1: %v", err)
				json.NewEncoder(w).Encode([]Project{})
				return
			}

			projects := make([]Project, 0)
			for _, dbProject := range dbProjects {
				// Get base directories for this project
				dbBaseDirs, err := queries.GetBaseDirectoriesByProjectID(ctx, dbProject.ID)
				if err != nil {
					// Log error but continue
					log.Printf("Failed to get base directories for project %d: %v", dbProject.ID, err)
					dbBaseDirs = []db.BaseDirectory{}
				}

				var baseDirs []BaseDirectory
				for _, dbBaseDir := range dbBaseDirs {
					baseDirs = append(baseDirs, dbBaseDirectoryToBaseDirectory(dbBaseDir))
				}

				// Get tasks for this project (simplified for now)
				tasks := []Task{}

				// Ensure arrays are never null
				if baseDirs == nil {
					baseDirs = []BaseDirectory{}
				}
				if tasks == nil {
					tasks = []Task{}
				}
				
				project := Project{
					ID:              dbProject.ID,
					Name:            dbProject.Name,
					BaseDirectories: baseDirs,
					Tasks:           tasks,
				}
				projects = append(projects, project)
			}

			json.NewEncoder(w).Encode(projects)
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

			// Ensure arrays are never null
			if baseDirs == nil {
				baseDirs = []BaseDirectory{}
			}
			if tasks == nil {
				tasks = []Task{}
			}

			result := Project{
				ID:              project.ID,
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
			ID:              project.ID,
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

func handleProjectTasksAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	if len(pathParts) < 2 {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}
	
	_, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}
	
	switch r.Method {
	case "GET":
		// Get tasks for project - for now return empty array as tasks aren't fully implemented
		json.NewEncoder(w).Encode([]Task{})
		
	case "POST":
		// Create a new task for this project
		var createReq struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Status      string `json:"status"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		// For now, just return a mock task since the task table isn't fully implemented
		mockTask := Task{
			Title:       createReq.Title,
			Description: createReq.Description,
			Status:      createReq.Status,
			Worktree:    Worktree{}, // Empty worktree for now
		}
		
		json.NewEncoder(w).Encode(mockTask)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleProjectBaseDirectoriesAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	if len(pathParts) < 2 {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}
	
	projectID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}
	
	switch r.Method {
	case "GET":
		// Get base directories for project
		dbBaseDirs, err := queries.GetBaseDirectoriesByProjectID(ctx, projectID)
		if err != nil {
			log.Printf("Failed to get base directories for project %d: %v", projectID, err)
			json.NewEncoder(w).Encode([]BaseDirectory{})
			return
		}
		
		baseDirs := make([]BaseDirectory, 0)
		for _, dbBaseDir := range dbBaseDirs {
			baseDirs = append(baseDirs, dbBaseDirectoryToBaseDirectory(dbBaseDir))
		}
		
		json.NewEncoder(w).Encode(baseDirs)
		
	case "POST":
		// Create a new base directory for this project
		var createReq struct {
			Path                      string `json:"path"`
			GitInitialized            bool   `json:"gitInitialized"`
			WorktreeSetupCommands     string `json:"worktreeSetupCommands"`
			WorktreeTeardownCommands  string `json:"worktreeTeardownCommands"`
			DevServerSetupCommands    string `json:"devServerSetupCommands"`
			DevServerTeardownCommands string `json:"devServerTeardownCommands"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		// Generate a unique base directory ID
		baseDirectoryID := fmt.Sprintf("bd_%d_%d", projectID, time.Now().Unix())
		
		dbBaseDir, err := queries.CreateBaseDirectory(ctx, db.CreateBaseDirectoryParams{
			ProjectID:                 projectID,
			BaseDirectoryID:           baseDirectoryID,
			Path:                      createReq.Path,
			GitInitialized:            createReq.GitInitialized,
			WorktreeSetupCommands:     createReq.WorktreeSetupCommands,
			WorktreeTeardownCommands:  createReq.WorktreeTeardownCommands,
			DevServerSetupCommands:    createReq.DevServerSetupCommands,
			DevServerTeardownCommands: createReq.DevServerTeardownCommands,
		})
		if err != nil {
			http.Error(w, "Failed to create base directory", http.StatusInternalServerError)
			return
		}
		
		result := dbBaseDirectoryToBaseDirectory(dbBaseDir)
		json.NewEncoder(w).Encode(result)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
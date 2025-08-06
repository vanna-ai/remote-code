package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
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

				// Get tasks for this project
				tasks, err := loadTasksForProject(ctx, dbProject.ID, baseDirs)
				if err != nil {
					log.Printf("Failed to get tasks for project %d: %v", dbProject.ID, err)
					tasks = []Task{}
				}

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

			// Get tasks for this project
			tasks, err := loadTasksForProject(ctx, id, baseDirs)
			if err != nil {
				log.Printf("Failed to get tasks for project %d: %v", id, err)
				tasks = []Task{}
			}

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
	// Handle specific agent operations: /api/agents/{id}
	if len(pathParts) > 0 && pathParts[0] != "detect" {
		handleSingleAgentAPI(w, r, ctx, pathParts)
		return
	}
	
	switch r.Method {
	case "GET":
		if len(pathParts) > 0 && pathParts[0] == "detect" {
			// Detect available agents on the system
			handleAgentDetection(w, r, ctx)
			return
		}
		
		// List all agents for default root (assuming root_id = 1 for now)
		dbAgents, err := queries.GetAgentsByRootID(ctx, 1)
		if err != nil {
			log.Printf("Failed to get agents: %v", err)
			json.NewEncoder(w).Encode([]Agent{})
			return
		}
		
		agents := make([]Agent, 0)
		for _, dbAgent := range dbAgents {
			agents = append(agents, Agent{
				ID:      dbAgent.ID,
				Name:    dbAgent.Name,
				Command: dbAgent.Command,
				Params:  dbAgent.Params,
			})
		}
		
		json.NewEncoder(w).Encode(agents)
		
	case "POST":
		// Create a new agent
		var createReq struct {
			RootId  int64  `json:"root_id"`
			Name    string `json:"name"`
			Command string `json:"command"`
			Params  string `json:"params"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		agent, err := queries.CreateAgent(ctx, db.CreateAgentParams{
			RootID:  createReq.RootId,
			Name:    createReq.Name,
			Command: createReq.Command,
			Params:  createReq.Params,
		})
		if err != nil {
			log.Printf("Failed to create agent: %v", err)
			http.Error(w, "Failed to create agent", http.StatusInternalServerError)
			return
		}
		
		result := Agent{
			ID:      agent.ID,
			Name:    agent.Name,
			Command: agent.Command,
			Params:  agent.Params,
		}
		json.NewEncoder(w).Encode(result)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAgentDetection(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	// Define the agents to detect
	agentsToDetect := []string{"claude", "amp", "codex", "gemini", "codabuff", "aider", "opencode", "friday", "grok"}
	
	var availableAgents []map[string]interface{}
	
	for _, agentName := range agentsToDetect {
		cmd := exec.Command("which", agentName)
		output, err := cmd.Output()
		
		if err == nil && len(output) > 0 {
			// Agent is available
			path := strings.TrimSpace(string(output))
			availableAgents = append(availableAgents, map[string]interface{}{
				"name":      agentName,
				"command":   agentName,
				"path":      path,
				"available": true,
			})
		} else {
			// Agent not found
			availableAgents = append(availableAgents, map[string]interface{}{
				"name":      agentName,
				"command":   agentName,
				"path":      "",
				"available": false,
			})
		}
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents": availableAgents,
	})
}

func handleSingleAgentAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	if len(pathParts) == 0 {
		http.Error(w, "Agent ID required", http.StatusBadRequest)
		return
	}
	
	agentID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}
	
	switch r.Method {
	case "PUT":
		// Update agent
		var updateReq struct {
			Name    string `json:"name"`
			Command string `json:"command"`
			Params  string `json:"params"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		updatedAgent, err := queries.UpdateAgent(ctx, db.UpdateAgentParams{
			ID:      agentID,
			Name:    updateReq.Name,
			Command: updateReq.Command,
			Params:  updateReq.Params,
		})
		if err != nil {
			log.Printf("Failed to update agent: %v", err)
			http.Error(w, "Failed to update agent", http.StatusInternalServerError)
			return
		}
		
		result := Agent{
			ID:      updatedAgent.ID,
			Name:    updatedAgent.Name,
			Command: updatedAgent.Command,
			Params:  updatedAgent.Params,
		}
		json.NewEncoder(w).Encode(result)
		
	case "DELETE":
		// Delete agent
		err := queries.DeleteAgent(ctx, agentID)
		if err != nil {
			log.Printf("Failed to delete agent: %v", err)
			http.Error(w, "Failed to delete agent", http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusNoContent)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
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
	
	projectID, err := strconv.ParseInt(pathParts[0], 10, 64)
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
			Title           string `json:"title"`
			Description     string `json:"description"`
			Status          string `json:"status"`
			BaseDirectoryId string `json:"baseDirectoryId"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		if createReq.BaseDirectoryId == "" {
			http.Error(w, "Base directory ID is required", http.StatusBadRequest)
			return
		}
		
		// Create the task (no worktree creation needed)
		dbTask, err := queries.CreateTask(ctx, db.CreateTaskParams{
			ProjectID:       projectID,
			BaseDirectoryID: createReq.BaseDirectoryId,
			Title:           createReq.Title,
			Description:     createReq.Description,
			Status:          createReq.Status,
		})
		if err != nil {
			log.Printf("Failed to create task: %v", err)
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			return
		}
		
		// Get the base directory info to include in response
		dbBaseDirs, err := queries.GetBaseDirectoriesByProjectID(ctx, projectID)
		if err != nil {
			http.Error(w, "Failed to get base directories", http.StatusInternalServerError)
			return
		}
		
		var baseDirectory BaseDirectory
		for _, dir := range dbBaseDirs {
			if dir.BaseDirectoryID == createReq.BaseDirectoryId {
				baseDirectory = dbBaseDirectoryToBaseDirectory(dir)
				break
			}
		}
		
		// Convert to API model
		task := dbTaskToTask(dbTask, baseDirectory)
		json.NewEncoder(w).Encode(task)
		
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
	
	// Check if we have a specific directory ID in the path (for DELETE operations)
	if len(pathParts) >= 3 {
		directoryIDParam := pathParts[2]
		handleSingleBaseDirectoryAPI(w, r, ctx, projectID, directoryIDParam)
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

func loadTasksForProject(ctx context.Context, projectID int64, baseDirs []BaseDirectory) ([]Task, error) {
	dbTasks, err := queries.GetTasksByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	
	// Create a map for fast base directory lookup
	baseDirMap := make(map[string]BaseDirectory)
	for _, dir := range baseDirs {
		baseDirMap[dir.BaseDirectoryId] = dir
	}
	
	var tasks []Task
	for _, dbTask := range dbTasks {
		baseDir, exists := baseDirMap[dbTask.BaseDirectoryID]
		if !exists {
			// Skip tasks with missing base directories
			continue
		}
		tasks = append(tasks, dbTaskToTask(dbTask, baseDir))
	}
	
	return tasks, nil
}

func handleSingleBaseDirectoryAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, projectID int64, directoryIDParam string) {
	switch r.Method {
	case "DELETE":
		// Find the directory by base_directory_id and project_id
		dbBaseDirs, err := queries.GetBaseDirectoriesByProjectID(ctx, projectID)
		if err != nil {
			http.Error(w, "Failed to get base directories", http.StatusInternalServerError)
			return
		}
		
		var directoryToDelete *db.BaseDirectory
		for _, dir := range dbBaseDirs {
			if dir.BaseDirectoryID == directoryIDParam {
				directoryToDelete = &dir
				break
			}
		}
		
		if directoryToDelete == nil {
			http.Error(w, "Base directory not found", http.StatusNotFound)
			return
		}
		
		// Delete the directory
		err = queries.DeleteBaseDirectory(ctx, directoryToDelete.ID)
		if err != nil {
			log.Printf("Failed to delete base directory %d: %v", directoryToDelete.ID, err)
			http.Error(w, "Failed to delete base directory", http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusNoContent)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
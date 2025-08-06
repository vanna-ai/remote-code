package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
	case "dashboard":
		handleDashboardAPI(w, r, ctx, pathParts[1:])
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
	case "task-executions":
		handleTaskExecutionsAPI(w, r, ctx, pathParts[1:])
	case "tmux-sessions":
		handleTmuxSessionsAPI(w, r, ctx, pathParts[1:])
	default:
		http.Error(w, "Unknown API endpoint", http.StatusNotFound)
	}
}

type DashboardStats struct {
	ActiveSessions   int `json:"active_sessions"`
	Projects         int `json:"projects"`
	TaskExecutions   int `json:"task_executions"`
	Agents          int `json:"agents"`
}

func handleDashboardAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	switch r.Method {
	case "GET":
		if len(pathParts) > 0 && pathParts[0] == "stats" {
			// Get dashboard statistics
			stats := DashboardStats{}
			
			// Count active tmux sessions
			cmd := exec.Command("tmux", "list-sessions")
			output, err := cmd.Output()
			if err == nil {
				sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
				if len(sessions) > 0 && sessions[0] != "" {
					stats.ActiveSessions = len(sessions)
				}
			}
			
			// Count projects
			projects, err := queries.ListProjects(ctx)
			if err == nil {
				stats.Projects = len(projects)
			}
			
			// Count task executions
			executions, err := queries.ListTaskExecutions(ctx)
			if err == nil {
				stats.TaskExecutions = len(executions)
			}
			
			// Count agents
			agents, err := queries.ListAgents(ctx)
			if err == nil {
				stats.Agents = len(agents)
			}
			
			json.NewEncoder(w).Encode(stats)
		} else {
			http.Error(w, "Unknown dashboard endpoint", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

func handleTmuxSessionsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	switch r.Method {
	case "GET":
		// Get all tmux sessions
		sessions, err := getTmuxSessions()
		if err != nil {
			log.Printf("Failed to get tmux sessions: %v", err)
			http.Error(w, "Failed to get tmux sessions", http.StatusInternalServerError)
			return
		}
		
		json.NewEncoder(w).Encode(sessions)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type TmuxSession struct {
	Name     string `json:"name"`
	Created  string `json:"created"`
	Preview  string `json:"preview"`
	TaskID   *int64 `json:"task_id,omitempty"`
	AgentID  *int64 `json:"agent_id,omitempty"`
	IsTask   bool   `json:"is_task"`
}

func getTmuxSessions() ([]TmuxSession, error) {
	// Get list of tmux sessions
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}|#{session_created}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var sessions []TmuxSession
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue
		}
		
		sessionName := parts[0]
		created := parts[1]
		
		// Get preview of the session
		preview := ""
		previewCmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-p")
		if previewOutput, err := previewCmd.Output(); err == nil {
			// Get last few lines
			previewLines := strings.Split(strings.TrimSpace(string(previewOutput)), "\n")
			if len(previewLines) > 10 {
				previewLines = previewLines[len(previewLines)-10:]
			}
			preview = strings.Join(previewLines, "\n")
		}
		
		session := TmuxSession{
			Name:    sessionName,
			Created: created,
			Preview: preview,
		}
		
		// Check if this is a task session
		if strings.HasPrefix(sessionName, "task_") {
			session.IsTask = true
			// Parse task and agent IDs from session name (format: task_{taskId}_agent_{agentId})
			parts := strings.Split(sessionName, "_")
			if len(parts) >= 4 {
				if taskID, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					session.TaskID = &taskID
				}
				if agentID, err := strconv.ParseInt(parts[3], 10, 64); err == nil {
					session.AgentID = &agentID
				}
			}
		}
		
		sessions = append(sessions, session)
	}
	
	return sessions, nil
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
	agentsToDetect := []string{"claude", "amp", "codex", "gemini", "codebuff", "aider", "opencode", "friday", "grok"}
	
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

func handleTaskExecutionsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	switch r.Method {
	case "GET":
		// Handle individual task execution by ID
		if len(pathParts) > 0 {
			executionID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid execution ID", http.StatusBadRequest)
				return
			}
			
			queries := db.New(database)
			execution, err := queries.GetTaskExecutionWithDetails(ctx, executionID)
			if err != nil {
				log.Printf("Failed to get task execution: %v", err)
				http.Error(w, "Task execution not found", http.StatusNotFound)
				return
			}
			
			json.NewEncoder(w).Encode(execution)
			return
		}
		
		// Get task executions, optionally filtered by task_id
		taskIDStr := r.URL.Query().Get("task_id")
		if taskIDStr != "" {
			taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
			if err != nil {
				http.Error(w, "Invalid task ID", http.StatusBadRequest)
				return
			}
			
			queries := db.New(database)
			executions, err := queries.GetTaskExecutionsByTaskID(ctx, taskID)
			if err != nil {
			log.Printf("Failed to get task executions: %v", err)
			http.Error(w, "Failed to get task executions", http.StatusInternalServerError)
			return
			}
			
			// Ensure we return empty array instead of null
		if executions == nil {
			executions = []db.TaskExecution{}
		}
		
		json.NewEncoder(w).Encode(executions)
			return
		}
		
		// Get all task executions
		queries := db.New(database)
		executions, err := queries.ListTaskExecutions(ctx)
		if err != nil {
			log.Printf("Failed to list task executions: %v", err)
			http.Error(w, "Failed to list task executions", http.StatusInternalServerError)
			return
		}
		
		// Ensure we return empty array instead of null
		if executions == nil {
			executions = []db.TaskExecution{}
		}
		
		json.NewEncoder(w).Encode(executions)
		
	case "POST":
		// Create and start a new task execution
		var createReq struct {
			TaskId  int64 `json:"task_id"`
			AgentId int64 `json:"agent_id"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		// Get the task details
		dbTask, err := queries.GetTask(ctx, createReq.TaskId)
		if err != nil {
			log.Printf("Failed to get task: %v", err)
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		
		// Get the agent details
		dbAgent, err := queries.GetAgent(ctx, createReq.AgentId)
		if err != nil {
			log.Printf("Failed to get agent: %v", err)
			http.Error(w, "Agent not found", http.StatusNotFound)
			return
		}
		
		// Create a unique worktree path for this execution
		worktreePath := fmt.Sprintf("/tmp/task_%d_agent_%d_%d", createReq.TaskId, createReq.AgentId, time.Now().Unix())
		
		// Create the worktree
		dbWorktree, err := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
			BaseDirectoryID: dbTask.BaseDirectoryID,
			Path:            worktreePath,
			AgentTmuxID:     sql.NullString{Valid: false},
			DevServerTmuxID: sql.NullString{Valid: false},
			ExternalUrl:     sql.NullString{Valid: false},
		})
		if err != nil {
			log.Printf("Failed to create worktree: %v", err)
			http.Error(w, "Failed to create worktree", http.StatusInternalServerError)
			return
		}
		
		// Create the task execution record
		dbTaskExecution, err := queries.CreateTaskExecution(ctx, db.CreateTaskExecutionParams{
			TaskID:     createReq.TaskId,
			AgentID:    createReq.AgentId,
			WorktreeID: dbWorktree.ID,
			Status:     "starting",
		})
		if err != nil {
			log.Printf("Failed to create task execution: %v", err)
			http.Error(w, "Failed to create task execution", http.StatusInternalServerError)
			return
		}
		
		// Update task status to "in_progress" if it's not already
		if dbTask.Status != "in_progress" {
			_, err = queries.UpdateTask(ctx, db.UpdateTaskParams{
				ID:          dbTask.ID,
				Title:       dbTask.Title,
				Description: dbTask.Description,
				Status:      "in_progress",
			})
			if err != nil {
				log.Printf("Failed to update task status: %v", err)
			}
		}
		
		// Start the execution in the background
		go startTaskExecutionProcess(dbTaskExecution.ID, dbTask, dbAgent, dbWorktree)
		
		// Return the task execution details
		result := map[string]interface{}{
			"id":          dbTaskExecution.ID,
			"task_id":     dbTaskExecution.TaskID,
			"agent_id":    dbTaskExecution.AgentID,
			"worktree_id": dbTaskExecution.WorktreeID,
			"status":      dbTaskExecution.Status,
		}
		json.NewEncoder(w).Encode(result)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func startTaskExecutionProcess(executionID int64, task db.Task, agent db.Agent, worktree db.Worktree) {
	ctx := context.Background()
	
	log.Printf("Starting task execution %d: Task '%s' with agent '%s'", executionID, task.Title, agent.Name)
	
	// Get the base directory for setup commands
	baseDir, err := queries.GetBaseDirectoryByProjectAndID(ctx, db.GetBaseDirectoryByProjectAndIDParams{
		ProjectID:       task.ProjectID,
		BaseDirectoryID: worktree.BaseDirectoryID,
	})
	if err != nil {
		log.Printf("Failed to get base directory: %v", err)
		updateTaskExecutionStatus(ctx, executionID, "failed")
		return
	}
	
	// Create the worktree directory
	err = os.MkdirAll(worktree.Path, 0755)
	if err != nil {
		log.Printf("Failed to create worktree directory %s: %v", worktree.Path, err)
		updateTaskExecutionStatus(ctx, executionID, "failed")
		return
	}
	
	// Generate a unique tmux session name
	sessionName := fmt.Sprintf("task_%d_agent_%d", task.ID, agent.ID)
	
	// Start tmux session in the worktree directory
	tmuxCmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", worktree.Path)
	err = tmuxCmd.Run()
	if err != nil {
		log.Printf("Failed to start tmux session: %v", err)
		updateTaskExecutionStatus(ctx, executionID, "failed")
		return
	}
	
	// Update worktree with tmux session info
	_, err = queries.UpdateWorktree(ctx, db.UpdateWorktreeParams{
		ID:              worktree.ID,
		Path:            worktree.Path,
		AgentTmuxID:     sql.NullString{String: sessionName, Valid: true},
		DevServerTmuxID: worktree.DevServerTmuxID,
		ExternalUrl:     worktree.ExternalUrl,
	})
	if err != nil {
		log.Printf("Failed to update worktree with tmux session: %v", err)
	}
	
	// Function to send command and wait
	sendCommandAndWait := func(command string, description string) error {
		log.Printf("Executing %s: %s", description, command)
		cmd := exec.Command("tmux", "send-keys", "-t", sessionName, command, "Enter")
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to send %s command: %v", description, err)
		}
		
		// Wait a moment for command to execute
		time.Sleep(2 * time.Second)
		return nil
	}
	
	// Run git worktree setup if this is a git-initialized base directory
	if baseDir.GitInitialized {
		// Set up git worktree
		worktreeCommand := fmt.Sprintf("git worktree add %s", worktree.Path)
		// Change to base directory first, then set up worktree
		if err := sendCommandAndWait(fmt.Sprintf("cd %s", baseDir.Path), "change to base directory"); err != nil {
			log.Printf("Warning: %v", err)
		}
		if err := sendCommandAndWait(worktreeCommand, "git worktree setup"); err != nil {
			log.Printf("Warning: %v", err)
		}
		// Change back to worktree directory
		if err := sendCommandAndWait(fmt.Sprintf("cd %s", worktree.Path), "change to worktree directory"); err != nil {
			log.Printf("Warning: %v", err)
		}
	}
	
	// Run custom worktree setup commands if provided
	if baseDir.WorktreeSetupCommands != "" {
		if err := sendCommandAndWait(baseDir.WorktreeSetupCommands, "worktree setup commands"); err != nil {
			log.Printf("Warning: %v", err)
		}
	}
	
	// Create a task file with the task details
	taskContent := fmt.Sprintf(`Task: %s

Description: %s

Working directory: %s

Instructions:
- Use this directory as your working directory
- Complete the task described above
- The codebase should be available in this directory
`, task.Title, task.Description, worktree.Path)
	
	taskFilePath := filepath.Join(worktree.Path, "TASK.md")
	err = os.WriteFile(taskFilePath, []byte(taskContent), 0644)
	if err != nil {
		log.Printf("Failed to create task file: %v", err)
	}
	
	// Show the task file in the terminal
	if err := sendCommandAndWait(fmt.Sprintf("cat %s", taskFilePath), "show task file"); err != nil {
		log.Printf("Warning: %v", err)
	}
	
	// Start the agent command
	agentCommand := fmt.Sprintf("%s %s", agent.Command, agent.Params)
	if err := sendCommandAndWait(agentCommand, "agent command"); err != nil {
		log.Printf("Failed to start agent: %v", err)
		updateTaskExecutionStatus(ctx, executionID, "failed")
		return
	}
	
	// Update status to running
	updateTaskExecutionStatus(ctx, executionID, "running")
	
	log.Printf("Task execution %d started successfully in tmux session %s", executionID, sessionName)
}

func updateTaskExecutionStatus(ctx context.Context, executionID int64, status string) {
	_, err := queries.UpdateTaskExecutionStatus(ctx, db.UpdateTaskExecutionStatusParams{
		ID:     executionID,
		Status: status,
	})
	if err != nil {
		log.Printf("Failed to update task execution status: %v", err)
	}
}

func handleBaseDirectoriesAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Base directories API not implemented yet"})
}

func handleWorktreesAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	queries := db.New(database)
	
	switch r.Method {
	case "GET":
		if len(pathParts) == 0 {
			// List all worktrees - for now return empty array since no ListWorktrees method exists
			json.NewEncoder(w).Encode([]db.Worktree{})
		} else {
			// Get specific worktree
			worktreeID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid worktree ID", http.StatusBadRequest)
				return
			}
			
			worktree, err := queries.GetWorktree(ctx, worktreeID)
			if err != nil {
				log.Printf("Failed to get worktree: %v", err)
				http.Error(w, "Worktree not found", http.StatusNotFound)
				return
			}
			
			json.NewEncoder(w).Encode(worktree)
		}
		
	case "POST":
		// Handle sub-endpoints like /api/worktrees/{id}/dev-server
		if len(pathParts) >= 2 && pathParts[1] == "dev-server" {
			worktreeID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid worktree ID", http.StatusBadRequest)
				return
			}
			
			// Start dev server for worktree
			err = startDevServer(ctx, queries, worktreeID)
			if err != nil {
				log.Printf("Failed to start dev server: %v", err)
				http.Error(w, "Failed to start dev server", http.StatusInternalServerError)
				return
			}
			
			json.NewEncoder(w).Encode(map[string]string{"status": "dev server started"})
		} else {
			http.Error(w, "Invalid endpoint", http.StatusBadRequest)
		}
		
	case "DELETE":
		// Handle sub-endpoints like /api/worktrees/{id}/dev-server
		if len(pathParts) >= 2 && pathParts[1] == "dev-server" {
			worktreeID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid worktree ID", http.StatusBadRequest)
				return
			}
			
			// Stop dev server for worktree
			err = stopDevServer(ctx, queries, worktreeID)
			if err != nil {
				log.Printf("Failed to stop dev server: %v", err)
				http.Error(w, "Failed to stop dev server", http.StatusInternalServerError)
				return
			}
			
			json.NewEncoder(w).Encode(map[string]string{"status": "dev server stopped"})
		} else {
			http.Error(w, "Invalid endpoint", http.StatusBadRequest)
		}
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func startDevServer(ctx context.Context, queries *db.Queries, worktreeID int64) error {
	worktree, err := queries.GetWorktree(ctx, worktreeID)
	if err != nil {
		return err
	}
	
	// We need to get the task to find dev server setup commands
	// First, let's get the task execution to find the task_id
	executions, err := queries.ListTaskExecutions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list task executions: %v", err)
	}
	
	var taskID int64
	for _, exec := range executions {
		if exec.WorktreeID == worktreeID {
			taskID = exec.TaskID
			break
		}
	}
	
	if taskID == 0 {
		return fmt.Errorf("no task found for worktree %d", worktreeID)
	}
	
	// Get task with base directory info to access dev server setup commands
	taskWithBaseDir, err := queries.GetTaskWithBaseDirectory(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task with base directory: %v", err)
	}
	
	// Create dev server session name
	devSessionName := fmt.Sprintf("dev_%d", worktreeID)
	
	// Start tmux session for dev server in the worktree directory
	tmuxCmd := exec.Command("tmux", "new-session", "-d", "-s", devSessionName, "-c", worktree.Path)
	err = tmuxCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to create dev server tmux session: %v", err)
	}
	
	// If there are dev server setup commands, execute them in the session
	if taskWithBaseDir.DevServerSetupCommands != "" {
		// Execute the dev server setup commands
		setupCmd := exec.Command("tmux", "send-keys", "-t", devSessionName, taskWithBaseDir.DevServerSetupCommands, "Enter")
		err = setupCmd.Run()
		if err != nil {
			log.Printf("Failed to send dev server setup commands: %v", err)
			// Don't return error here, session is still created
		}
		
		// Add a separator and info message
		infoCmd := exec.Command("tmux", "send-keys", "-t", devSessionName, "", "Enter")
		infoCmd.Run()
		
		echoCmd := exec.Command("tmux", "send-keys", "-t", devSessionName, "echo 'Dev server started. Session: "+devSessionName+"'", "Enter")
		echoCmd.Run()
	} else {
		// No setup commands, just show info
		echoCmd := exec.Command("tmux", "send-keys", "-t", devSessionName, "echo 'Dev server session created. No setup commands configured.'", "Enter")
		echoCmd.Run()
		
		bashCmd := exec.Command("tmux", "send-keys", "-t", devSessionName, "bash", "Enter")
		bashCmd.Run()
	}
	
	// Update worktree with dev server session info
	_, err = queries.UpdateWorktree(ctx, db.UpdateWorktreeParams{
		ID:              worktree.ID,
		Path:            worktree.Path,
		AgentTmuxID:     worktree.AgentTmuxID,
		DevServerTmuxID: sql.NullString{String: devSessionName, Valid: true},
		ExternalUrl:     worktree.ExternalUrl,
	})
	
	return err
}

func stopDevServer(ctx context.Context, queries *db.Queries, worktreeID int64) error {
	worktree, err := queries.GetWorktree(ctx, worktreeID)
	if err != nil {
		return err
	}
	
	if worktree.DevServerTmuxID.Valid {
		// Kill the tmux session
		tmuxCmd := exec.Command("tmux", "kill-session", "-t", worktree.DevServerTmuxID.String)
		_ = tmuxCmd.Run() // Ignore error if session doesn't exist
	}
	
	// Update worktree to remove dev server session info
	_, err = queries.UpdateWorktree(ctx, db.UpdateWorktreeParams{
		ID:              worktree.ID,
		Path:            worktree.Path,
		AgentTmuxID:     worktree.AgentTmuxID,
		DevServerTmuxID: sql.NullString{Valid: false},
		ExternalUrl:     worktree.ExternalUrl,
	})
	
	return err
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

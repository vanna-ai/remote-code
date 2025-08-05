package main

import (
	"database/sql"
	"remote-code/db"
)

// Root represents the main configuration structure
type Root struct {
	Projects        []Project `yaml:"projects" json:"projects"`
	AvailableAgents []Agent   `yaml:"available_agents" json:"available_agents"`
	LocalPort       string    `yaml:"local_port" json:"local_port"`
	ExternalUrl     *string   `yaml:"-" json:"external_url,omitempty"`
}

// Project represents a project configuration
type Project struct {
	ID              int64           `json:"id"`
	Name            string          `yaml:"name" json:"name"`
	BaseDirectories []BaseDirectory `yaml:"base_directories" json:"baseDirectories"`
	Tasks           []Task          `yaml:"tasks" json:"tasks"`
}

// BaseDirectory represents a base directory configuration
type BaseDirectory struct {
	BaseDirectoryId           string `yaml:"base_directory_id" json:"base_directory_id"`
	Path                      string `yaml:"path" json:"path"`
	GitInitialized            bool   `yaml:"git_initialized" json:"git_initialized"`
	WorktreeSetupCommands     string `yaml:"worktree_setup_commands" json:"worktree_setup_commands"`
	WorktreeTeardownCommands  string `yaml:"worktree_teardown_commands" json:"worktree_teardown_commands"`
	DevServerSetupCommands    string `yaml:"dev_server_setup_commands" json:"dev_server_setup_commands"`
	DevServerTeardownCommands string `yaml:"dev_server_teardown_commands" json:"dev_server_teardown_commands"`
}

// Worktree represents a worktree instance
type Worktree struct {
	BaseDirectoryId string  `yaml:"base_directory_id" json:"base_directory_id"`
	Path            string  `yaml:"path" json:"path"`
	AgentTmuxId     *string `yaml:"-" json:"agent_tmux_id,omitempty"`
	DevServerTmuxId *string `yaml:"-" json:"dev_server_tmux_id,omitempty"`
	ExternalUrl     *string `yaml:"-" json:"external_url,omitempty"`
}

// Task represents a task configuration
type Task struct {
	ID              int64         `json:"id"`
	Title           string        `yaml:"title" json:"title"`
	Description     string        `yaml:"description" json:"description"`
	Status          string        `json:"status"` // For Kanban board (todo, in_progress, done)
	BaseDirectory   BaseDirectory `json:"baseDirectory"`
}

// TaskExecution represents a task being executed by an agent in a worktree
type TaskExecution struct {
	ID        int64    `json:"id"`
	TaskID    int64    `json:"taskId"`
	Status    string   `json:"status"`
	Agent     Agent    `json:"agent"`
	Worktree  Worktree `json:"worktree"`
}

// Agent represents an available agent
type Agent struct {
	Name    string `yaml:"name" json:"name"`
	Command string `yaml:"command" json:"command"`
	Params  string `yaml:"params" json:"params"`
}

// Conversion functions from database models to API models

func dbRootToRoot(dbRoot db.Root, agents []db.Agent, projects []Project) Root {
	var externalUrl *string
	if dbRoot.ExternalUrl.Valid {
		externalUrl = &dbRoot.ExternalUrl.String
	}

	var availableAgents []Agent
	for _, agent := range agents {
		availableAgents = append(availableAgents, Agent{
			Name:    agent.Name,
			Command: agent.Command,
			Params:  agent.Params,
		})
	}

	return Root{
		Projects:        projects,
		AvailableAgents: availableAgents,
		LocalPort:       dbRoot.LocalPort,
		ExternalUrl:     externalUrl,
	}
}

func dbBaseDirectoryToBaseDirectory(dbBaseDir db.BaseDirectory) BaseDirectory {
	return BaseDirectory{
		BaseDirectoryId:           dbBaseDir.BaseDirectoryID,
		Path:                      dbBaseDir.Path,
		GitInitialized:            dbBaseDir.GitInitialized,
		WorktreeSetupCommands:     dbBaseDir.WorktreeSetupCommands,
		WorktreeTeardownCommands:  dbBaseDir.WorktreeTeardownCommands,
		DevServerSetupCommands:    dbBaseDir.DevServerSetupCommands,
		DevServerTeardownCommands: dbBaseDir.DevServerTeardownCommands,
	}
}

func dbWorktreeToWorktree(dbWorktree db.Worktree) Worktree {
	var agentTmuxId, devServerTmuxId, externalUrl *string
	
	if dbWorktree.AgentTmuxID.Valid {
		agentTmuxId = &dbWorktree.AgentTmuxID.String
	}
	if dbWorktree.DevServerTmuxID.Valid {
		devServerTmuxId = &dbWorktree.DevServerTmuxID.String
	}
	if dbWorktree.ExternalUrl.Valid {
		externalUrl = &dbWorktree.ExternalUrl.String
	}

	return Worktree{
		BaseDirectoryId: dbWorktree.BaseDirectoryID,
		Path:            dbWorktree.Path,
		AgentTmuxId:     agentTmuxId,
		DevServerTmuxId: devServerTmuxId,
		ExternalUrl:     externalUrl,
	}
}

func dbTaskToTask(dbTask db.Task, baseDirectory BaseDirectory) Task {
	return Task{
		ID:            dbTask.ID,
		Title:         dbTask.Title,
		Description:   dbTask.Description,
		Status:        dbTask.Status,
		BaseDirectory: baseDirectory,
	}
}

// Helper functions for converting to sql.NullString
func stringToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}
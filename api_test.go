package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"remote-code/db"
)

var testDbPath string

func TestMain(m *testing.M) {
	// Setup test database
	database, queries, testDbPath = initTestDatabase()
	defer database.Close()
	
	// Clean up test database file when done
	defer func() {
		os.Remove(testDbPath)
		// Also clean up any other test database files just in case
		matches, _ := filepath.Glob("remote-code-test-*.db")
		for _, match := range matches {
			os.Remove(match)
		}
	}()
	
	// Run tests
	code := m.Run()
	os.Exit(code)
}

func setupTestDB(t *testing.T) {
	// Clean existing data for isolated tests
	ctx := context.Background()
	database.ExecContext(ctx, "DELETE FROM tasks")
	database.ExecContext(ctx, "DELETE FROM worktrees")
	database.ExecContext(ctx, "DELETE FROM base_directories")
	database.ExecContext(ctx, "DELETE FROM projects")
	database.ExecContext(ctx, "DELETE FROM agents")
	database.ExecContext(ctx, "DELETE FROM roots")
}

func TestProjectsAPI_GET_Empty(t *testing.T) {
	setupTestDB(t)
	
	req := httptest.NewRequest("GET", "/api/projects", nil)
	w := httptest.NewRecorder()
	
	handleAPI(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var projects []Project
	if err := json.Unmarshal(w.Body.Bytes(), &projects); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if len(projects) != 0 {
		t.Errorf("Expected empty projects list, got %d projects", len(projects))
	}
}

func TestProjectsAPI_POST_Create(t *testing.T) {
	setupTestDB(t)
	
	// First create a root to associate the project with
	ctx := context.Background()
	root, err := queries.CreateRoot(ctx, db.CreateRootParams{
		LocalPort: "8080",
	})
	if err != nil {
		t.Fatalf("Failed to create root: %v", err)
	}
	
	projectData := map[string]interface{}{
		"root_id": root.ID,
		"name":    "Test Project",
	}
	
	jsonData, _ := json.Marshal(projectData)
	req := httptest.NewRequest("POST", "/api/projects", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handleAPI(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}
	
	var project Project
	if err := json.Unmarshal(w.Body.Bytes(), &project); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if project.Name != "Test Project" {
		t.Errorf("Expected project name 'Test Project', got '%s'", project.Name)
	}
}

func TestProjectsAPI_GET_WithProjects(t *testing.T) {
	setupTestDB(t)
	
	// Create a project with root_id = 1 (which is what the API uses)
	ctx := context.Background()
	_, err := queries.CreateProject(ctx, db.CreateProjectParams{
		RootID: 1,
		Name:   "Another Test Project",
	})
	if err != nil {
		// If root_id = 1 doesn't exist, create it first
		_, err = queries.CreateRoot(ctx, db.CreateRootParams{
			LocalPort: "8080",
		})
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}
		
		// Try creating the project again
		_, err = queries.CreateProject(ctx, db.CreateProjectParams{
			RootID: 1,
			Name:   "Another Test Project",
		})
		if err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}
	}
	
	req := httptest.NewRequest("GET", "/api/projects", nil)
	w := httptest.NewRecorder()
	
	handleAPI(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var projects []Project
	if err := json.Unmarshal(w.Body.Bytes(), &projects); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	// Should have at least the project we just created
	if len(projects) == 0 {
		t.Errorf("Expected at least 1 project, got %d", len(projects))
	}
	
	// Check if our project is in the list
	found := false
	for _, project := range projects {
		if project.Name == "Another Test Project" {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("Expected to find 'Another Test Project' in projects list")
	}
}

func TestProjectsAPI_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/projects", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handleAPI(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestProjectsAPI_CORS(t *testing.T) {
	req := httptest.NewRequest("OPTIONS", "/api/projects", nil)
	w := httptest.NewRecorder()
	
	handleAPI(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS header to be '*', got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestProjectsAPI_ArraysNotNull(t *testing.T) {
	setupTestDB(t)
	
	ctx := context.Background()
	_, err := queries.CreateProject(ctx, db.CreateProjectParams{
		RootID: 1,
		Name:   "Test Project",
	})
	if err != nil {
		// Create root first if it doesn't exist
		_, err = queries.CreateRoot(ctx, db.CreateRootParams{
			LocalPort: "8080",
		})
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}
		
		_, err = queries.CreateProject(ctx, db.CreateProjectParams{
			RootID: 1,
			Name:   "Test Project",
		})
		if err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}
	}
	
	req := httptest.NewRequest("GET", "/api/projects", nil)
	w := httptest.NewRecorder()
	
	handleAPI(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Parse response to verify array structure
	var projects []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &projects); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if len(projects) == 0 {
		t.Errorf("Expected at least one project")
		return
	}
	
	project := projects[0]
	
	// Check that base_directories is an array, not null
	baseDirs, exists := project["base_directories"]
	if !exists {
		t.Errorf("base_directories field missing from response")
	} else if baseDirs == nil {
		t.Errorf("base_directories should not be null, should be an empty array")
	}
	
	// Check that tasks is an array, not null
	tasks, exists := project["tasks"]
	if !exists {
		t.Errorf("tasks field missing from response")
	} else if tasks == nil {
		t.Errorf("tasks should not be null, should be an empty array")
	}
}

func TestProjectTasksAPI_GET(t *testing.T) {
	setupTestDB(t)
	
	req := httptest.NewRequest("GET", "/api/projects/1/tasks", nil)
	w := httptest.NewRecorder()
	
	handleAPI(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var tasks []Task
	if err := json.Unmarshal(w.Body.Bytes(), &tasks); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	// Should return empty array for now
	if len(tasks) != 0 {
		t.Errorf("Expected empty tasks array, got %d tasks", len(tasks))
	}
}

func TestProjectTasksAPI_POST(t *testing.T) {
	setupTestDB(t)
	
	taskData := map[string]interface{}{
		"title":       "Test Task",
		"description": "Test task description",
		"status":      "todo",
	}
	
	jsonData, _ := json.Marshal(taskData)
	req := httptest.NewRequest("POST", "/api/projects/1/tasks", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handleAPI(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}
	
	var task Task
	if err := json.Unmarshal(w.Body.Bytes(), &task); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if task.Title != "Test Task" {
		t.Errorf("Expected task title 'Test Task', got '%s'", task.Title)
	}
	
	if task.Status != "todo" {
		t.Errorf("Expected task status 'todo', got '%s'", task.Status)
	}
}

func TestProjectBaseDirectoriesAPI_POST(t *testing.T) {
	setupTestDB(t)
	
	// Create a project first
	ctx := context.Background()
	project, err := queries.CreateProject(ctx, db.CreateProjectParams{
		RootID: 1,
		Name:   "Test Project",
	})
	if err != nil {
		// Create root first if it doesn't exist
		_, err = queries.CreateRoot(ctx, db.CreateRootParams{
			LocalPort: "8080",
		})
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}
		
		project, err = queries.CreateProject(ctx, db.CreateProjectParams{
			RootID: 1,
			Name:   "Test Project",
		})
		if err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}
	}
	
	directoryData := map[string]interface{}{
		"path":                        "/test/directory",
		"gitInitialized":             true,
		"worktreeSetupCommands":      "npm install",
		"worktreeTeardownCommands":   "",
		"devServerSetupCommands":     "npm run dev",
		"devServerTeardownCommands":  "",
	}
	
	jsonData, _ := json.Marshal(directoryData)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/projects/%d/base-directories", project.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handleAPI(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}
	
	var directory BaseDirectory
	if err := json.Unmarshal(w.Body.Bytes(), &directory); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if directory.Path != "/test/directory" {
		t.Errorf("Expected directory path '/test/directory', got '%s'", directory.Path)
	}
	
	if !directory.GitInitialized {
		t.Errorf("Expected GitInitialized to be true")
	}
}
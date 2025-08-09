package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"remote-code/db"
)

func createTestCompetitionData(t *testing.T, ctx context.Context) (db.Agent, db.Agent, db.Task, db.TaskExecution, db.TaskExecution, db.AgentCompetition) {
	// Create test agents
	root, err := queries.CreateRoot(ctx, db.CreateRootParams{
		LocalPort:   "8080",
		ExternalUrl: sql.NullString{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create test root: %v", err)
	}

	agent1, err := queries.CreateAgent(ctx, db.CreateAgentParams{
		RootID:  root.ID,
		Name:    "APITestAgent1",
		Command: "test-agent-1",
		Params:  "",
	})
	if err != nil {
		t.Fatalf("Failed to create test agent1: %v", err)
	}

	agent2, err := queries.CreateAgent(ctx, db.CreateAgentParams{
		RootID:  root.ID,
		Name:    "APITestAgent2",
		Command: "test-agent-2",
		Params:  "",
	})
	if err != nil {
		t.Fatalf("Failed to create test agent2: %v", err)
	}

	// Create project and task
	project, err := queries.CreateProject(ctx, db.CreateProjectParams{
		RootID: root.ID,
		Name:   "API Test Project",
	})
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	baseDir, err := queries.CreateBaseDirectory(ctx, db.CreateBaseDirectoryParams{
		ProjectID:                 project.ID,
		BaseDirectoryID:          "api-test-dir",
		Path:                      "/tmp/apitest",
		GitInitialized:           false,
		WorktreeSetupCommands:     "",
		WorktreeTeardownCommands:  "",
		DevServerSetupCommands:    "",
		DevServerTeardownCommands: "",
	})
	if err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}

	task, err := queries.CreateTask(ctx, db.CreateTaskParams{
		ProjectID:       project.ID,
		BaseDirectoryID: baseDir.BaseDirectoryID,
		Title:           "API Test Task",
		Description:     "Test Description",
		Status:          "pending",
	})
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Create worktrees and executions
	worktree1, err := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
		BaseDirectoryID: baseDir.BaseDirectoryID,
		Path:            "/tmp/apitest1",
		AgentTmuxID:     sql.NullString{Valid: false},
		DevServerTmuxID: sql.NullString{Valid: false},
		ExternalUrl:     sql.NullString{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create worktree1: %v", err)
	}

	worktree2, err := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
		BaseDirectoryID: baseDir.BaseDirectoryID,
		Path:            "/tmp/apitest2",
		AgentTmuxID:     sql.NullString{Valid: false},
		DevServerTmuxID: sql.NullString{Valid: false},
		ExternalUrl:     sql.NullString{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create worktree2: %v", err)
	}

	execution1, err := queries.CreateTaskExecution(ctx, db.CreateTaskExecutionParams{
		TaskID:     task.ID,
		AgentID:    agent1.ID,
		WorktreeID: worktree1.ID,
		Status:     "completed",
	})
	if err != nil {
		t.Fatalf("Failed to create execution1: %v", err)
	}

	execution2, err := queries.CreateTaskExecution(ctx, db.CreateTaskExecutionParams{
		TaskID:     task.ID,
		AgentID:    agent2.ID,
		WorktreeID: worktree2.ID,
		Status:     "failed",
	})
	if err != nil {
		t.Fatalf("Failed to create execution2: %v", err)
	}

	// Create competition
	competition, err := queries.CreateCompetition(ctx, db.CreateCompetitionParams{
		TaskID:            task.ID,
		Agent1ID:          agent1.ID,
		Agent2ID:          agent2.ID,
		Agent1ExecutionID: execution1.ID,
		Agent2ExecutionID: execution2.ID,
		WinnerAgentID:     sql.NullInt64{Valid: true, Int64: agent1.ID},
		Agent1EloBefore:   1500.0,
		Agent2EloBefore:   1500.0,
		Agent1EloAfter:    1516.0,
		Agent2EloAfter:    1484.0,
		KFactor:           sql.NullFloat64{Valid: true, Float64: 32.0},
		CompetitionType:   sql.NullString{Valid: true, String: "head_to_head"},
		Notes:             sql.NullString{Valid: true, String: "API test competition"},
	})
	if err != nil {
		t.Fatalf("Failed to create competition: %v", err)
	}

	// Update agent ELO ratings
	_, err = queries.UpdateAgentELO(ctx, db.UpdateAgentELOParams{
		EloRating: sql.NullFloat64{Valid: true, Float64: 1516.0},
		Wins:      sql.NullInt64{Valid: true, Int64: 1},
		Losses:    sql.NullInt64{Valid: true, Int64: 0},
		Draws:     sql.NullInt64{Valid: true, Int64: 0},
		ID:        agent1.ID,
	})
	if err != nil {
		t.Fatalf("Failed to update agent1 ELO: %v", err)
	}

	_, err = queries.UpdateAgentELO(ctx, db.UpdateAgentELOParams{
		EloRating: sql.NullFloat64{Valid: true, Float64: 1484.0},
		Wins:      sql.NullInt64{Valid: true, Int64: 0},
		Losses:    sql.NullInt64{Valid: true, Int64: 1},
		Draws:     sql.NullInt64{Valid: true, Int64: 0},
		ID:        agent2.ID,
	})
	if err != nil {
		t.Fatalf("Failed to update agent2 ELO: %v", err)
	}

	return agent1, agent2, task, execution1, execution2, competition
}

func TestELOAPI_GetLeaderboard(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()

	agent1, agent2, _, _, _, _ := createTestCompetitionData(t, ctx)

	req := httptest.NewRequest("GET", "/api/elo/leaderboard", nil)
	w := httptest.NewRecorder()

	handleAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var leaderboard []db.AgentLeaderboard
	err := json.Unmarshal(w.Body.Bytes(), &leaderboard)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(leaderboard) != 2 {
		t.Errorf("Expected 2 agents in leaderboard, got %d", len(leaderboard))
	}

	// First agent should be the winner (higher ELO)
	if leaderboard[0].ID != agent1.ID {
		t.Errorf("Expected agent1 to be first in leaderboard, got agent %d", leaderboard[0].ID)
	}

	if !leaderboard[0].EloRating.Valid || leaderboard[0].EloRating.Float64 != 1516.0 {
		t.Errorf("Expected agent1 ELO to be 1516, got %v", leaderboard[0].EloRating)
	}

	// Second agent should be the loser (lower ELO)
	if leaderboard[1].ID != agent2.ID {
		t.Errorf("Expected agent2 to be second in leaderboard, got agent %d", leaderboard[1].ID)
	}

	if !leaderboard[1].EloRating.Valid || leaderboard[1].EloRating.Float64 != 1484.0 {
		t.Errorf("Expected agent2 ELO to be 1484, got %v", leaderboard[1].EloRating)
	}
}

func TestELOAPI_GetAgentHistory(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()

	agent1, _, _, _, _, _ := createTestCompetitionData(t, ctx)

	url := "/api/elo/agent/" + strconv.FormatInt(agent1.ID, 10) + "/history"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	handleAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var history []db.GetAgentELOHistoryRow
	err := json.Unmarshal(w.Body.Bytes(), &history)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(history))
	}

	if history[0].EloRating != 1516.0 {
		t.Errorf("Expected ELO rating 1516, got %v", history[0].EloRating)
	}
}

func TestELOAPI_GetHeadToHeadRecord(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()

	agent1, agent2, _, _, _, _ := createTestCompetitionData(t, ctx)

	url := "/api/elo/head-to-head/" + strconv.FormatInt(agent1.ID, 10) + "/" + strconv.FormatInt(agent2.ID, 10)
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	handleAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var record db.GetHeadToHeadRecordRow
	err := json.Unmarshal(w.Body.Bytes(), &record)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if record.TotalGames != 1 {
		t.Errorf("Expected 1 total game, got %d", record.TotalGames)
	}

	if !record.Agent1Wins.Valid || record.Agent1Wins.Float64 != 1.0 {
		t.Errorf("Expected agent1 to have 1 win, got %v", record.Agent1Wins)
	}

	if !record.Agent2Wins.Valid || record.Agent2Wins.Float64 != 0.0 {
		t.Errorf("Expected agent2 to have 0 wins, got %v", record.Agent2Wins)
	}
}

func TestCompetitionsAPI_ListCompetitions(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()

	_, _, _, _, _, competition := createTestCompetitionData(t, ctx)

	req := httptest.NewRequest("GET", "/api/competitions", nil)
	w := httptest.NewRecorder()

	handleAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var competitions []db.AgentCompetition
	err := json.Unmarshal(w.Body.Bytes(), &competitions)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(competitions) != 1 {
		t.Errorf("Expected 1 competition, got %d", len(competitions))
	}

	if competitions[0].ID != competition.ID {
		t.Errorf("Expected competition ID %d, got %d", competition.ID, competitions[0].ID)
	}
}

func TestCompetitionsAPI_GetCompetition(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()

	_, _, _, _, _, competition := createTestCompetitionData(t, ctx)

	url := "/api/competitions/" + strconv.FormatInt(competition.ID, 10)
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	handleAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var retrievedCompetition db.AgentCompetition
	err := json.Unmarshal(w.Body.Bytes(), &retrievedCompetition)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if retrievedCompetition.ID != competition.ID {
		t.Errorf("Expected competition ID %d, got %d", competition.ID, retrievedCompetition.ID)
	}

	if retrievedCompetition.Agent1EloBefore != 1500.0 {
		t.Errorf("Expected agent1 ELO before 1500, got %v", retrievedCompetition.Agent1EloBefore)
	}

	if retrievedCompetition.Agent1EloAfter != 1516.0 {
		t.Errorf("Expected agent1 ELO after 1516, got %v", retrievedCompetition.Agent1EloAfter)
	}
}

func TestCompetitionsAPI_CreateCompetition(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()

	agent1, agent2, task, execution1, execution2, _ := createTestCompetitionData(t, ctx)

	// Reset agent ELO ratings for clean test
	_, err := queries.UpdateAgentELO(ctx, db.UpdateAgentELOParams{
		EloRating: sql.NullFloat64{Valid: true, Float64: 1500.0},
		Wins:      sql.NullInt64{Valid: true, Int64: -1},
		Losses:    sql.NullInt64{Valid: true, Int64: -1},
		Draws:     sql.NullInt64{Valid: true, Int64: 0},
		ID:        agent1.ID,
	})
	if err != nil {
		t.Fatalf("Failed to reset agent1: %v", err)
	}

	_, err = queries.UpdateAgentELO(ctx, db.UpdateAgentELOParams{
		EloRating: sql.NullFloat64{Valid: true, Float64: 1500.0},
		Wins:      sql.NullInt64{Valid: true, Int64: 0},
		Losses:    sql.NullInt64{Valid: true, Int64: -1},
		Draws:     sql.NullInt64{Valid: true, Int64: 0},
		ID:        agent2.ID,
	})
	if err != nil {
		t.Fatalf("Failed to reset agent2: %v", err)
	}

	// Create competition via API
	competitionData := map[string]interface{}{
		"task_id":             task.ID,
		"agent1_id":           agent1.ID,
		"agent2_id":           agent2.ID,
		"agent1_execution_id": execution1.ID,
		"agent2_execution_id": execution2.ID,
		"result":              "agent1_wins",
		"notes":               "API created competition",
	}

	jsonData, err := json.Marshal(competitionData)
	if err != nil {
		t.Fatalf("Failed to marshal competition data: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/competitions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	var createdCompetition db.AgentCompetition
	err = json.Unmarshal(w.Body.Bytes(), &createdCompetition)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if createdCompetition.TaskID != task.ID {
		t.Errorf("Expected task ID %d, got %d", task.ID, createdCompetition.TaskID)
	}

	if !createdCompetition.WinnerAgentID.Valid || createdCompetition.WinnerAgentID.Int64 != agent1.ID {
		t.Errorf("Expected winner to be agent1 (%d), got %v", agent1.ID, createdCompetition.WinnerAgentID)
	}

	// Verify ELO ratings were updated
	updatedAgent1, err := queries.GetAgent(ctx, agent1.ID)
	if err != nil {
		t.Fatalf("Failed to get updated agent1: %v", err)
	}

	if !updatedAgent1.EloRating.Valid || updatedAgent1.EloRating.Float64 <= 1500.0 {
		t.Errorf("Agent1 ELO should be higher than 1500, got %v", updatedAgent1.EloRating)
	}
}

func TestCompetitionsAPI_ProcessTask(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()

	agent1, agent2, task, _, _, _ := createTestCompetitionData(t, ctx)

	// Reset agent stats for clean test
	database.ExecContext(ctx, "DELETE FROM agent_competitions")
	_, err := queries.UpdateAgentELO(ctx, db.UpdateAgentELOParams{
		EloRating: sql.NullFloat64{Valid: true, Float64: 1500.0},
		Wins:      sql.NullInt64{Valid: true, Int64: -1},
		Losses:    sql.NullInt64{Valid: true, Int64: -1},
		Draws:     sql.NullInt64{Valid: true, Int64: 0},
		ID:        agent1.ID,
	})
	if err != nil {
		t.Fatalf("Failed to reset agent1: %v", err)
	}

	_, err = queries.UpdateAgentELO(ctx, db.UpdateAgentELOParams{
		EloRating: sql.NullFloat64{Valid: true, Float64: 1500.0},
		Wins:      sql.NullInt64{Valid: true, Int64: 0},
		Losses:    sql.NullInt64{Valid: true, Int64: -1},
		Draws:     sql.NullInt64{Valid: true, Int64: 0},
		ID:        agent2.ID,
	})
	if err != nil {
		t.Fatalf("Failed to reset agent2: %v", err)
	}

	url := "/api/competitions/process-task/" + strconv.FormatInt(task.ID, 10)
	req := httptest.NewRequest("POST", url, nil)
	w := httptest.NewRecorder()

	handleAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	var result TaskCompetitionResult
	err = json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.TaskID != task.ID {
		t.Errorf("Expected task ID %d, got %d", task.ID, result.TaskID)
	}

	if result.TotalCompetitions != 1 {
		t.Errorf("Expected 1 competition to be processed, got %d", result.TotalCompetitions)
	}

	// Verify competition was created
	competitions, err := queries.ListCompetitionsByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to list competitions: %v", err)
	}

	if len(competitions) != 1 {
		t.Errorf("Expected 1 competition in database, got %d", len(competitions))
	}
}

func TestCompetitionsAPI_GetCompetitionHistory(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()

	_, _, _, _, _, _ = createTestCompetitionData(t, ctx)

	req := httptest.NewRequest("GET", "/api/competitions/history", nil)
	w := httptest.NewRecorder()

	handleAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var history []db.CompetitionHistory
	err := json.Unmarshal(w.Body.Bytes(), &history)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(history))
	}

	if history[0].Agent1Name != "APITestAgent1" {
		t.Errorf("Expected agent1 name 'APITestAgent1', got %v", history[0].Agent1Name)
	}

	if history[0].Agent2Name != "APITestAgent2" {
		t.Errorf("Expected agent2 name 'APITestAgent2', got %v", history[0].Agent2Name)
	}
}

func TestCompetitionsAPI_InvalidRequests(t *testing.T) {
	setupELOTestDB(t)

	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		expectedStatus int
	}{
		{
			name:           "Invalid competition ID",
			method:         "GET",
			url:            "/api/competitions/999999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid agent ID in ELO history",
			method:         "GET",
			url:            "/api/elo/agent/999999/history",
			expectedStatus: http.StatusOK, // Returns empty array
		},
		{
			name:           "Invalid JSON in create competition",
			method:         "POST",
			url:            "/api/competitions",
			body:           "{invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing required fields in create competition",
			method:         "POST",
			url:            "/api/competitions",
			body:           `{"result":"agent1_wins"}`,
			expectedStatus: http.StatusInternalServerError, // Agent not found
		},
		{
			name:           "Invalid result value",
			method:         "POST",
			url:            "/api/competitions",
			body:           `{"task_id":1,"agent1_id":1,"agent2_id":2,"agent1_execution_id":1,"agent2_execution_id":2,"result":"invalid"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.url, bytes.NewBuffer([]byte(tt.body)))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.url, nil)
			}

			w := httptest.NewRecorder()
			handleAPI(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}
package main

import (
	"context"
	"database/sql"
	"testing"
	"remote-code/db"
	"math"
)

func TestELOCalculator_CalculateELO(t *testing.T) {
	eloCalc := &ELOCalculator{}

	tests := []struct {
		name           string
		agent1Rating   float64
		agent2Rating   float64
		result         MatchResult
		kFactor        float64
		expectedAgent1 float64
		expectedAgent2 float64
	}{
		{
			name:           "Agent1 wins equal ratings",
			agent1Rating:   1500,
			agent2Rating:   1500,
			result:         Agent1Wins,
			kFactor:        32,
			expectedAgent1: 1516,
			expectedAgent2: 1484,
		},
		{
			name:           "Agent2 wins equal ratings",
			agent1Rating:   1500,
			agent2Rating:   1500,
			result:         Agent2Wins,
			kFactor:        32,
			expectedAgent1: 1484,
			expectedAgent2: 1516,
		},
		{
			name:           "Draw equal ratings",
			agent1Rating:   1500,
			agent2Rating:   1500,
			result:         Draw,
			kFactor:        32,
			expectedAgent1: 1500,
			expectedAgent2: 1500,
		},
		{
			name:           "Underdog wins",
			agent1Rating:   1400,
			agent2Rating:   1600,
			result:         Agent1Wins,
			kFactor:        32,
			expectedAgent1: 1424.64,
			expectedAgent2: 1575.36,
		},
		{
			name:           "Favorite wins",
			agent1Rating:   1600,
			agent2Rating:   1400,
			result:         Agent1Wins,
			kFactor:        32,
			expectedAgent1: 1607.36,
			expectedAgent2: 1392.64,
		},
		{
			name:           "High K-factor impact",
			agent1Rating:   1500,
			agent2Rating:   1500,
			result:         Agent1Wins,
			kFactor:        64,
			expectedAgent1: 1532,
			expectedAgent2: 1468,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eloCalc.CalculateELO(tt.agent1Rating, tt.agent2Rating, tt.result, tt.kFactor)
			
			if math.Abs(result.Agent1NewRating-tt.expectedAgent1) > 1.0 {
				t.Errorf("Agent1NewRating = %v, expected %v", result.Agent1NewRating, tt.expectedAgent1)
			}
			
			if math.Abs(result.Agent2NewRating-tt.expectedAgent2) > 1.0 {
				t.Errorf("Agent2NewRating = %v, expected %v", result.Agent2NewRating, tt.expectedAgent2)
			}
			
			// Verify rating changes sum to zero
			totalChange := result.Agent1Change + result.Agent2Change
			if math.Abs(totalChange) > 0.001 {
				t.Errorf("Total rating change should be zero, got %v", totalChange)
			}
		})
	}
}

func TestELOCalculator_DetermineKFactor(t *testing.T) {
	eloCalc := &ELOCalculator{}

	tests := []struct {
		name         string
		gamesPlayed  int64
		currentRating float64
		expected     float64
	}{
		{"New player", 5, 1500, 64.0},
		{"New player high rating", 25, 2000, 64.0},
		{"Experienced player normal rating", 50, 1800, 32.0},
		{"Experienced player high rating", 100, 2150, 16.0},
		{"Master level", 200, 2400, 16.0},
		{"Exactly 30 games", 30, 1500, 32.0},
		{"Exactly 2100 rating", 50, 2100, 16.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eloCalc.DetermineKFactor(tt.gamesPlayed, tt.currentRating)
			if result != tt.expected {
				t.Errorf("DetermineKFactor(%d, %v) = %v, expected %v", 
					tt.gamesPlayed, tt.currentRating, result, tt.expected)
			}
		})
	}
}

func TestELOCalculator_DetermineWinner(t *testing.T) {
	eloCalc := &ELOCalculator{}

	tests := []struct {
		name          string
		agent1Status  string
		agent2Status  string
		expectedResult MatchResult
		shouldError   bool
	}{
		{"Both completed - draw", "completed", "completed", Draw, false},
		{"Agent1 completed, Agent2 failed", "completed", "failed", Agent1Wins, false},
		{"Agent1 failed, Agent2 completed", "failed", "completed", Agent2Wins, false},
		{"Agent1 completed, Agent2 running", "completed", "running", Agent1Wins, false},
		{"Both failed - draw", "failed", "failed", Draw, true},
		{"Both running - error", "running", "running", Draw, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eloCalc.DetermineWinner(tt.agent1Status, tt.agent2Status)
			
			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if !tt.shouldError && result != tt.expectedResult {
				t.Errorf("DetermineWinner(%s, %s) = %v, expected %v", 
					tt.agent1Status, tt.agent2Status, result, tt.expectedResult)
			}
		})
	}
}

func setupELOTestDB(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()
	
	// Also clean ELO-specific tables
	database.ExecContext(ctx, "DELETE FROM agent_competitions")
}

func createTestAgentsForELO(t *testing.T, ctx context.Context) (db.Agent, db.Agent, db.Root) {
	// Create a test root
	root, err := queries.CreateRoot(ctx, db.CreateRootParams{
		LocalPort:   "8080",
		ExternalUrl: sql.NullString{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create test root: %v", err)
	}

	// Create two test agents
	agent1, err := queries.CreateAgent(ctx, db.CreateAgentParams{
		RootID:  root.ID,
		Name:    "TestAgent1",
		Command: "test-agent-1",
		Params:  "",
	})
	if err != nil {
		t.Fatalf("Failed to create test agent1: %v", err)
	}

	agent2, err := queries.CreateAgent(ctx, db.CreateAgentParams{
		RootID:  root.ID,
		Name:    "TestAgent2",
		Command: "test-agent-2",
		Params:  "",
	})
	if err != nil {
		t.Fatalf("Failed to create test agent2: %v", err)
	}

	return agent1, agent2, root
}

func TestELOCalculator_RecordCompetition(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()
	eloCalc := NewELOCalculator(queries)

	agent1, agent2, _ := createTestAgentsForELO(t, ctx)

	// Create a test project and task for the competition
	project, err := queries.CreateProject(ctx, db.CreateProjectParams{
		RootID: agent1.RootID,
		Name:   "Test Project",
	})
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	baseDir, err := queries.CreateBaseDirectory(ctx, db.CreateBaseDirectoryParams{
		ProjectID:                 project.ID,
		BaseDirectoryID:          "test-dir",
		Path:                      "/tmp/test",
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
		Title:           "Test Task",
		Description:     "Test Description",
		Status:          "pending",
	})
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Create task executions
	worktree1, err := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
		BaseDirectoryID: baseDir.BaseDirectoryID,
		Path:            "/tmp/test1",
		AgentTmuxID:     sql.NullString{Valid: false},
		DevServerTmuxID: sql.NullString{Valid: false},
		ExternalUrl:     sql.NullString{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create worktree1: %v", err)
	}

	worktree2, err := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
		BaseDirectoryID: baseDir.BaseDirectoryID,
		Path:            "/tmp/test2",
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

	// Record competition - Agent1 wins
	competition, err := eloCalc.RecordCompetition(ctx, CompetitionParams{
		TaskID:            task.ID,
		Agent1ID:          agent1.ID,
		Agent2ID:          agent2.ID,
		Agent1ExecutionID: execution1.ID,
		Agent2ExecutionID: execution2.ID,
		Result:            Agent1Wins,
		Notes:             "Test competition",
	})

	if err != nil {
		t.Fatalf("Failed to record competition: %v", err)
	}

	// Verify competition was created
	if competition.TaskID != task.ID {
		t.Errorf("Competition TaskID = %v, expected %v", competition.TaskID, task.ID)
	}

	if !competition.WinnerAgentID.Valid || competition.WinnerAgentID.Int64 != agent1.ID {
		t.Errorf("Competition WinnerAgentID = %v, expected %v", competition.WinnerAgentID, agent1.ID)
	}

	// Verify ELO ratings were updated
	updatedAgent1, err := queries.GetAgent(ctx, agent1.ID)
	if err != nil {
		t.Fatalf("Failed to get updated agent1: %v", err)
	}

	updatedAgent2, err := queries.GetAgent(ctx, agent2.ID)
	if err != nil {
		t.Fatalf("Failed to get updated agent2: %v", err)
	}

	// Agent1 should have higher rating (won)
	if !updatedAgent1.EloRating.Valid || updatedAgent1.EloRating.Float64 <= DefaultELORating {
		t.Errorf("Agent1 ELO should be higher than default, got %v", updatedAgent1.EloRating)
	}

	// Agent2 should have lower rating (lost)
	if !updatedAgent2.EloRating.Valid || updatedAgent2.EloRating.Float64 >= DefaultELORating {
		t.Errorf("Agent2 ELO should be lower than default, got %v", updatedAgent2.EloRating)
	}

	// Verify games played were incremented
	if !updatedAgent1.GamesPlayed.Valid || updatedAgent1.GamesPlayed.Int64 != 1 {
		t.Errorf("Agent1 games played should be 1, got %v", updatedAgent1.GamesPlayed)
	}

	if !updatedAgent2.GamesPlayed.Valid || updatedAgent2.GamesPlayed.Int64 != 1 {
		t.Errorf("Agent2 games played should be 1, got %v", updatedAgent2.GamesPlayed)
	}

	// Verify win/loss counts
	if !updatedAgent1.Wins.Valid || updatedAgent1.Wins.Int64 != 1 {
		t.Errorf("Agent1 wins should be 1, got %v", updatedAgent1.Wins)
	}

	if !updatedAgent2.Losses.Valid || updatedAgent2.Losses.Int64 != 1 {
		t.Errorf("Agent2 losses should be 1, got %v", updatedAgent2.Losses)
	}
}

func TestELOCalculator_ProcessTaskCompetitions(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()
	eloCalc := NewELOCalculator(queries)

	agent1, agent2, _ := createTestAgentsForELO(t, ctx)

	// Create a third agent
	agent3, err := queries.CreateAgent(ctx, db.CreateAgentParams{
		RootID:  agent1.RootID,
		Name:    "TestAgent3",
		Command: "test-agent-3",
		Params:  "",
	})
	if err != nil {
		t.Fatalf("Failed to create test agent3: %v", err)
	}

	// Create project and task
	project, err := queries.CreateProject(ctx, db.CreateProjectParams{
		RootID: agent1.RootID,
		Name:   "Test Project",
	})
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	baseDir, err := queries.CreateBaseDirectory(ctx, db.CreateBaseDirectoryParams{
		ProjectID:                 project.ID,
		BaseDirectoryID:          "test-dir",
		Path:                      "/tmp/test",
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
		Title:           "Test Task",
		Description:     "Test Description",
		Status:          "pending",
	})
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Create worktrees and task executions for all three agents
	agents := []db.Agent{agent1, agent2, agent3}
	statuses := []string{"completed", "failed", "completed"}

	for i, agent := range agents {
		worktree, err := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
			BaseDirectoryID: baseDir.BaseDirectoryID,
			Path:            "/tmp/test" + string(rune('1'+i)),
			AgentTmuxID:     sql.NullString{Valid: false},
			DevServerTmuxID: sql.NullString{Valid: false},
			ExternalUrl:     sql.NullString{Valid: false},
		})
		if err != nil {
			t.Fatalf("Failed to create worktree%d: %v", i+1, err)
		}

		_, err = queries.CreateTaskExecution(ctx, db.CreateTaskExecutionParams{
			TaskID:     task.ID,
			AgentID:    agent.ID,
			WorktreeID: worktree.ID,
			Status:     statuses[i],
		})
		if err != nil {
			t.Fatalf("Failed to create execution%d: %v", i+1, err)
		}
	}

	// Process competitions
	result, err := eloCalc.ProcessTaskCompetitions(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to process task competitions: %v", err)
	}

	// Should create 3 competitions: 1v2, 1v3, 2v3
	expectedCompetitions := 3
	if result.TotalCompetitions != expectedCompetitions {
		t.Errorf("Expected %d competitions, got %d", expectedCompetitions, result.TotalCompetitions)
	}

	// Verify competitions were created in database
	competitions, err := queries.ListCompetitionsByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to list competitions: %v", err)
	}

	if len(competitions) != expectedCompetitions {
		t.Errorf("Expected %d competitions in database, got %d", expectedCompetitions, len(competitions))
	}

	// Verify agent ELO ratings were updated
	for _, agent := range agents {
		updatedAgent, err := queries.GetAgent(ctx, agent.ID)
		if err != nil {
			t.Fatalf("Failed to get updated agent %d: %v", agent.ID, err)
		}

		if !updatedAgent.GamesPlayed.Valid || updatedAgent.GamesPlayed.Int64 != 2 {
			t.Errorf("Agent %d should have played 2 games, got %v", agent.ID, updatedAgent.GamesPlayed)
		}

		if !updatedAgent.EloRating.Valid {
			t.Errorf("Agent %d should have valid ELO rating", agent.ID)
		}
	}
}

func TestELOCalculator_PreventDuplicateCompetitions(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()
	eloCalc := NewELOCalculator(queries)

	agent1, agent2, _ := createTestAgentsForELO(t, ctx)

	// Create project and task setup (simplified)
	project, _ := queries.CreateProject(ctx, db.CreateProjectParams{
		RootID: agent1.RootID,
		Name:   "Test Project",
	})

	baseDir, _ := queries.CreateBaseDirectory(ctx, db.CreateBaseDirectoryParams{
		ProjectID:                 project.ID,
		BaseDirectoryID:          "test-dir",
		Path:                      "/tmp/test",
		GitInitialized:           false,
		WorktreeSetupCommands:     "",
		WorktreeTeardownCommands:  "",
		DevServerSetupCommands:    "",
		DevServerTeardownCommands: "",
	})

	task, _ := queries.CreateTask(ctx, db.CreateTaskParams{
		ProjectID:       project.ID,
		BaseDirectoryID: baseDir.BaseDirectoryID,
		Title:           "Test Task",
		Description:     "Test Description",
		Status:          "pending",
	})

	worktree1, _ := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
		BaseDirectoryID: baseDir.BaseDirectoryID,
		Path:            "/tmp/test1",
		AgentTmuxID:     sql.NullString{Valid: false},
		DevServerTmuxID: sql.NullString{Valid: false},
		ExternalUrl:     sql.NullString{Valid: false},
	})

	worktree2, _ := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
		BaseDirectoryID: baseDir.BaseDirectoryID,
		Path:            "/tmp/test2",
		AgentTmuxID:     sql.NullString{Valid: false},
		DevServerTmuxID: sql.NullString{Valid: false},
		ExternalUrl:     sql.NullString{Valid: false},
	})

	_, _ = queries.CreateTaskExecution(ctx, db.CreateTaskExecutionParams{
		TaskID:     task.ID,
		AgentID:    agent1.ID,
		WorktreeID: worktree1.ID,
		Status:     "completed",
	})

	_, _ = queries.CreateTaskExecution(ctx, db.CreateTaskExecutionParams{
		TaskID:     task.ID,
		AgentID:    agent2.ID,
		WorktreeID: worktree2.ID,
		Status:     "failed",
	})

	// Process competitions twice
	result1, err := eloCalc.ProcessTaskCompetitions(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed first competition processing: %v", err)
	}

	result2, err := eloCalc.ProcessTaskCompetitions(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed second competition processing: %v", err)
	}

	// First time should create 1 competition
	if result1.TotalCompetitions != 1 {
		t.Errorf("First processing should create 1 competition, got %d", result1.TotalCompetitions)
	}

	// Second time should create 0 competitions (duplicates prevented)
	if result2.TotalCompetitions != 0 {
		t.Errorf("Second processing should create 0 competitions, got %d", result2.TotalCompetitions)
	}

	// Verify only 1 competition exists in database
	competitions, err := queries.ListCompetitionsByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to list competitions: %v", err)
	}

	if len(competitions) != 1 {
		t.Errorf("Should have exactly 1 competition in database, got %d", len(competitions))
	}
}
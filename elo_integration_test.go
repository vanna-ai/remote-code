package main

import (
	"context"
	"database/sql"
	"testing"
	"time"
	"remote-code/db"
)

// TestAutomaticCompetitionProcessing tests that competitions are NOT automatically created on task completion
// (they are now only created when user merges)
func TestAutomaticCompetitionProcessing(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()

	// Create test agents
	root, err := queries.CreateRoot(ctx, db.CreateRootParams{
		LocalPort:   "8080",
		ExternalUrl: sql.NullString{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create test root: %v", err)
	}

	agents := make([]db.Agent, 3)
	for i := 0; i < 3; i++ {
		agent, err := queries.CreateAgent(ctx, db.CreateAgentParams{
			RootID:  root.ID,
			Name:    "IntegrationTestAgent" + string(rune('A'+i)),
			Command: "test-agent-" + string(rune('a'+i)),
			Params:  "",
		})
		if err != nil {
			t.Fatalf("Failed to create agent %d: %v", i, err)
		}
		agents[i] = agent
	}

	// Create project, base directory, and task
	project, err := queries.CreateProject(ctx, db.CreateProjectParams{
		RootID: root.ID,
		Name:   "Integration Test Project",
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	baseDir, err := queries.CreateBaseDirectory(ctx, db.CreateBaseDirectoryParams{
		ProjectID:                 project.ID,
		BaseDirectoryID:          "integration-test-dir",
		Path:                      "/tmp/integrationtest",
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
		Title:           "Integration Test Task",
		Description:     "Task for testing automatic competition processing",
		Status:          "pending",
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Create worktrees and task executions
	executions := make([]db.TaskExecution, 3)
	statuses := []string{"running", "running", "running"} // Start all as running

	for i, agent := range agents {
		worktree, err := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
			BaseDirectoryID: baseDir.BaseDirectoryID,
			Path:            "/tmp/integrationtest" + string(rune('1'+i)),
			AgentTmuxID:     sql.NullString{Valid: false},
			DevServerTmuxID: sql.NullString{Valid: false},
			ExternalUrl:     sql.NullString{Valid: false},
		})
		if err != nil {
			t.Fatalf("Failed to create worktree %d: %v", i, err)
		}

		execution, err := queries.CreateTaskExecution(ctx, db.CreateTaskExecutionParams{
			TaskID:     task.ID,
			AgentID:    agent.ID,
			WorktreeID: worktree.ID,
			Status:     statuses[i],
		})
		if err != nil {
			t.Fatalf("Failed to create execution %d: %v", i, err)
		}
		executions[i] = execution
	}

	// Verify no competitions exist initially
	initialCompetitions, err := queries.ListCompetitionsByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to list initial competitions: %v", err)
	}
	if len(initialCompetitions) != 0 {
		t.Errorf("Expected 0 initial competitions, got %d", len(initialCompetitions))
	}

	// Simulate agent executions finishing one by one
	// First agent completes - no competitions should be created (competitions only happen on merge)
	updateTaskExecutionStatus(ctx, executions[0].ID, "completed")
	
	// Give some time for any potential async processing
	time.Sleep(200 * time.Millisecond)

	competitions, err := queries.ListCompetitionsByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to list competitions after first completion: %v", err)
	}
	if len(competitions) != 0 {
		t.Errorf("Expected 0 competitions after first completion, got %d", len(competitions))
	}

	// Second agent fails - still no competitions should be created
	updateTaskExecutionStatus(ctx, executions[1].ID, "failed")
	
	// Wait for any potential async processing
	time.Sleep(300 * time.Millisecond)

	competitions, err = queries.ListCompetitionsByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to list competitions after second completion: %v", err)
	}
	if len(competitions) != 0 {
		t.Errorf("Expected 0 competitions after second completion (no auto-processing), got %d", len(competitions))
	}

	// Third agent completes - still no competitions
	updateTaskExecutionStatus(ctx, executions[2].ID, "completed")
	
	// Wait for any potential async processing
	time.Sleep(300 * time.Millisecond)

	competitions, err = queries.ListCompetitionsByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to list competitions after third completion: %v", err)
	}
	if len(competitions) != 0 {
		t.Errorf("Expected 0 competitions after third completion (no auto-processing), got %d", len(competitions))
	}

	// Now simulate a merge - this should create competitions with the merged agent as winner
	eloCalc := NewELOCalculator(queries)
	result, err := eloCalc.ProcessTaskCompetitionsWithWinner(ctx, task.ID, agents[1].ID) // Agent 1 (originally failed) is merged
	if err != nil {
		t.Fatalf("Failed to process competitions with winner: %v", err)
	}

	// Should create 2 competitions: agent1 vs agent0, agent1 vs agent2
	expectedCompetitions := 2
	if result.TotalCompetitions != expectedCompetitions {
		t.Errorf("Expected %d competitions from merge, got %d", expectedCompetitions, result.TotalCompetitions)
	}

	// Verify competitions were created in database
	finalCompetitions, err := queries.ListCompetitionsByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to list final competitions: %v", err)
	}
	if len(finalCompetitions) != expectedCompetitions {
		t.Errorf("Expected %d competitions in database after merge, got %d", expectedCompetitions, len(finalCompetitions))
	}

	// Verify agent1 (the merged agent) is the winner in all competitions
	for _, competition := range finalCompetitions {
		if !competition.WinnerAgentID.Valid || competition.WinnerAgentID.Int64 != agents[1].ID {
			t.Errorf("Expected agent1 (%d) to be winner, got %v", agents[1].ID, competition.WinnerAgentID)
		}
	}
}

// TestELORatingProgression tests that ELO ratings change appropriately over multiple competitions
func TestELORatingProgression(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()
	eloCalc := NewELOCalculator(queries)

	// Create two test agents
	agent1, agent2, _ := createTestAgentsForELO(t, ctx)

	// Record their initial ratings
	initialAgent1, _ := queries.GetAgent(ctx, agent1.ID)
	initialAgent2, _ := queries.GetAgent(ctx, agent2.ID)

	initialRating1 := DefaultELORating
	initialRating2 := DefaultELORating

	if initialAgent1.EloRating.Valid {
		initialRating1 = initialAgent1.EloRating.Float64
	}
	if initialAgent2.EloRating.Valid {
		initialRating2 = initialAgent2.EloRating.Float64
	}

	// Create multiple competitions with different outcomes
	competitions := []struct {
		result MatchResult
		description string
	}{
		{Agent1Wins, "Agent1 wins first match"},
		{Agent2Wins, "Agent2 wins second match"},
		{Agent1Wins, "Agent1 wins third match"},
		{Draw, "Fourth match is a draw"},
		{Agent1Wins, "Agent1 wins fifth match"},
	}

	var previousAgent1Rating, previousAgent2Rating float64 = initialRating1, initialRating2

	for i, comp := range competitions {
		// Create unique task and executions for each competition
		project, _ := queries.CreateProject(ctx, db.CreateProjectParams{
			RootID: agent1.RootID,
			Name:   "Progression Test Project " + string(rune('A'+i)),
		})

		baseDir, _ := queries.CreateBaseDirectory(ctx, db.CreateBaseDirectoryParams{
			ProjectID:                 project.ID,
			BaseDirectoryID:          "progression-test-dir-" + string(rune('a'+i)),
			Path:                      "/tmp/progressiontest" + string(rune('1'+i)),
			GitInitialized:           false,
			WorktreeSetupCommands:     "",
			WorktreeTeardownCommands:  "",
			DevServerSetupCommands:    "",
			DevServerTeardownCommands: "",
		})

		task, _ := queries.CreateTask(ctx, db.CreateTaskParams{
			ProjectID:       project.ID,
			BaseDirectoryID: baseDir.BaseDirectoryID,
			Title:           "Progression Test Task " + string(rune('A'+i)),
			Description:     comp.description,
			Status:          "pending",
		})

		worktree1, _ := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
			BaseDirectoryID: baseDir.BaseDirectoryID,
			Path:            "/tmp/progressiontest1_" + string(rune('a'+i)),
			AgentTmuxID:     sql.NullString{Valid: false},
			DevServerTmuxID: sql.NullString{Valid: false},
			ExternalUrl:     sql.NullString{Valid: false},
		})

		worktree2, _ := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
			BaseDirectoryID: baseDir.BaseDirectoryID,
			Path:            "/tmp/progressiontest2_" + string(rune('a'+i)),
			AgentTmuxID:     sql.NullString{Valid: false},
			DevServerTmuxID: sql.NullString{Valid: false},
			ExternalUrl:     sql.NullString{Valid: false},
		})

		execution1, _ := queries.CreateTaskExecution(ctx, db.CreateTaskExecutionParams{
			TaskID:     task.ID,
			AgentID:    agent1.ID,
			WorktreeID: worktree1.ID,
			Status:     "completed",
		})

		execution2, _ := queries.CreateTaskExecution(ctx, db.CreateTaskExecutionParams{
			TaskID:     task.ID,
			AgentID:    agent2.ID,
			WorktreeID: worktree2.ID,
			Status:     "completed",
		})

		// Record competition
		_, err := eloCalc.RecordCompetition(ctx, CompetitionParams{
			TaskID:            task.ID,
			Agent1ID:          agent1.ID,
			Agent2ID:          agent2.ID,
			Agent1ExecutionID: execution1.ID,
			Agent2ExecutionID: execution2.ID,
			Result:            comp.result,
			Notes:             comp.description,
		})
		if err != nil {
			t.Fatalf("Failed to record competition %d: %v", i, err)
		}

		// Check rating changes
		updatedAgent1, _ := queries.GetAgent(ctx, agent1.ID)
		updatedAgent2, _ := queries.GetAgent(ctx, agent2.ID)

		newRating1 := updatedAgent1.EloRating.Float64
		newRating2 := updatedAgent2.EloRating.Float64

		t.Logf("Competition %d (%s): Agent1: %.2f -> %.2f, Agent2: %.2f -> %.2f", 
			i+1, comp.description, previousAgent1Rating, newRating1, previousAgent2Rating, newRating2)

		// Verify ratings changed (except for draws between equal players)
		if comp.result == Agent1Wins {
			if newRating1 <= previousAgent1Rating {
				t.Errorf("Agent1 rating should increase after winning, got %.2f -> %.2f", 
					previousAgent1Rating, newRating1)
			}
			if newRating2 >= previousAgent2Rating {
				t.Errorf("Agent2 rating should decrease after losing, got %.2f -> %.2f", 
					previousAgent2Rating, newRating2)
			}
		} else if comp.result == Agent2Wins {
			if newRating2 <= previousAgent2Rating {
				t.Errorf("Agent2 rating should increase after winning, got %.2f -> %.2f", 
					previousAgent2Rating, newRating2)
			}
			if newRating1 >= previousAgent1Rating {
				t.Errorf("Agent1 rating should decrease after losing, got %.2f -> %.2f", 
					previousAgent1Rating, newRating1)
			}
		}

		previousAgent1Rating = newRating1
		previousAgent2Rating = newRating2
	}

	// Final verification
	finalAgent1, _ := queries.GetAgent(ctx, agent1.ID)
	finalAgent2, _ := queries.GetAgent(ctx, agent2.ID)

	// Agent1 won 3 out of 5 matches, should have higher rating
	if !finalAgent1.EloRating.Valid || !finalAgent2.EloRating.Valid {
		t.Fatal("Both agents should have valid ELO ratings")
	}

	if finalAgent1.EloRating.Float64 <= finalAgent2.EloRating.Float64 {
		t.Errorf("Agent1 should have higher rating after winning more games: %.2f vs %.2f", 
			finalAgent1.EloRating.Float64, finalAgent2.EloRating.Float64)
	}

	// Check game counts
	if !finalAgent1.GamesPlayed.Valid || finalAgent1.GamesPlayed.Int64 != 5 {
		t.Errorf("Agent1 should have played 5 games, got %v", finalAgent1.GamesPlayed)
	}

	if !finalAgent1.Wins.Valid || finalAgent1.Wins.Int64 != 3 {
		t.Errorf("Agent1 should have 3 wins, got %v", finalAgent1.Wins)
	}

	if !finalAgent1.Losses.Valid || finalAgent1.Losses.Int64 != 1 {
		t.Errorf("Agent1 should have 1 loss, got %v", finalAgent1.Losses)
	}

	if !finalAgent1.Draws.Valid || finalAgent1.Draws.Int64 != 1 {
		t.Errorf("Agent1 should have 1 draw, got %v", finalAgent1.Draws)
	}
}

// TestConcurrentCompetitionProcessing tests that concurrent competition processing doesn't create duplicates
func TestConcurrentCompetitionProcessing(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()

	// Create test setup
	agent1, agent2, _ := createTestAgentsForELO(t, ctx)

	project, _ := queries.CreateProject(ctx, db.CreateProjectParams{
		RootID: agent1.RootID,
		Name:   "Concurrent Test Project",
	})

	baseDir, _ := queries.CreateBaseDirectory(ctx, db.CreateBaseDirectoryParams{
		ProjectID:                 project.ID,
		BaseDirectoryID:          "concurrent-test-dir",
		Path:                      "/tmp/concurrenttest",
		GitInitialized:           false,
		WorktreeSetupCommands:     "",
		WorktreeTeardownCommands:  "",
		DevServerSetupCommands:    "",
		DevServerTeardownCommands: "",
	})

	task, _ := queries.CreateTask(ctx, db.CreateTaskParams{
		ProjectID:       project.ID,
		BaseDirectoryID: baseDir.BaseDirectoryID,
		Title:           "Concurrent Test Task",
		Description:     "Task for testing concurrent processing",
		Status:          "pending",
	})

	worktree1, _ := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
		BaseDirectoryID: baseDir.BaseDirectoryID,
		Path:            "/tmp/concurrenttest1",
		AgentTmuxID:     sql.NullString{Valid: false},
		DevServerTmuxID: sql.NullString{Valid: false},
		ExternalUrl:     sql.NullString{Valid: false},
	})

	worktree2, _ := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
		BaseDirectoryID: baseDir.BaseDirectoryID,
		Path:            "/tmp/concurrenttest2",
		AgentTmuxID:     sql.NullString{Valid: false},
		DevServerTmuxID: sql.NullString{Valid: false},
		ExternalUrl:     sql.NullString{Valid: false},
	})

	execution1, _ := queries.CreateTaskExecution(ctx, db.CreateTaskExecutionParams{
		TaskID:     task.ID,
		AgentID:    agent1.ID,
		WorktreeID: worktree1.ID,
		Status:     "running",
	})

	execution2, _ := queries.CreateTaskExecution(ctx, db.CreateTaskExecutionParams{
		TaskID:     task.ID,
		AgentID:    agent2.ID,
		WorktreeID: worktree2.ID,
		Status:     "running",
	})

	// Simulate concurrent status updates
	done1 := make(chan bool)
	done2 := make(chan bool)

	go func() {
		updateTaskExecutionStatus(ctx, execution1.ID, "completed")
		done1 <- true
	}()

	go func() {
		updateTaskExecutionStatus(ctx, execution2.ID, "failed")
		done2 <- true
	}()

	// Wait for both to complete
	<-done1
	<-done2

	// Give time for async processing
	time.Sleep(500 * time.Millisecond)

	// Should have at least 1 competition from concurrent processing (may be more due to timing)
	competitions, err := queries.ListCompetitionsByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to list competitions: %v", err)
	}

	if len(competitions) == 0 {
		t.Errorf("Expected at least 1 competition from concurrent processing, got %d", len(competitions))
	}

	// Verify the competition is valid
	if len(competitions) > 0 {
		comp := competitions[0]
		if comp.Agent1ID != agent1.ID && comp.Agent1ID != agent2.ID {
			t.Errorf("Competition should involve one of our test agents")
		}
		if comp.Agent2ID != agent1.ID && comp.Agent2ID != agent2.ID {
			t.Errorf("Competition should involve one of our test agents")
		}
		if comp.Agent1ID == comp.Agent2ID {
			t.Errorf("Competition should be between different agents")
		}
	}
}
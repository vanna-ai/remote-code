package main

import (
	"context"
	"database/sql"
	"testing"
	"remote-code/db"
)

func TestProcessTaskCompetitionsWithWinner(t *testing.T) {
	setupELOTestDB(t)
	ctx := context.Background()
	eloCalc := NewELOCalculator(queries)

	// Create test agents
	agent1, agent2, agent3, _ := createTestAgentsForMergeELO(t, ctx)

	// Create project, base directory, and task
	project, err := queries.CreateProject(ctx, db.CreateProjectParams{
		RootID: agent1.RootID,
		Name:   "Merge Test Project",
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	baseDir, err := queries.CreateBaseDirectory(ctx, db.CreateBaseDirectoryParams{
		ProjectID:                 project.ID,
		BaseDirectoryID:          "merge-test-dir",
		Path:                      "/tmp/mergetest",
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
		Title:           "Merge Test Task",
		Description:     "Task for testing merge-based winner determination",
		Status:          "pending",
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Create worktrees and task executions for all three agents
	agents := []db.Agent{agent1, agent2, agent3}
	statuses := []string{"completed", "failed", "completed"}

	for i, agent := range agents {
		worktree, err := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
			BaseDirectoryID: baseDir.BaseDirectoryID,
			Path:            "/tmp/mergetest" + string(rune('1'+i)),
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

	// Process competitions with agent2 (originally failed) as winner
	// This simulates the case where user merges agent2's execution despite it being marked as "failed"
	result, err := eloCalc.ProcessTaskCompetitionsWithWinner(ctx, task.ID, agent2.ID)
	if err != nil {
		t.Fatalf("Failed to process task competitions with winner: %v", err)
	}

	// Should create 2 competitions: agent2 vs agent1, agent2 vs agent3
	expectedCompetitions := 2
	if result.TotalCompetitions != expectedCompetitions {
		t.Errorf("Expected %d competitions, got %d", expectedCompetitions, result.TotalCompetitions)
	}

	if result.NewCompetitions != expectedCompetitions {
		t.Errorf("Expected %d new competitions, got %d", expectedCompetitions, result.NewCompetitions)
	}

	// Verify competitions were created in database
	competitions, err := queries.ListCompetitionsByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to list competitions: %v", err)
	}

	if len(competitions) != expectedCompetitions {
		t.Errorf("Expected %d competitions in database, got %d", expectedCompetitions, len(competitions))
	}

	// Verify agent2 is the winner in both competitions
	for _, competition := range competitions {
		if !competition.WinnerAgentID.Valid {
			t.Error("Competition should have a valid winner")
			continue
		}
		
		if competition.WinnerAgentID.Int64 != agent2.ID {
			t.Errorf("Expected agent2 (%d) to be winner, got agent %d", agent2.ID, competition.WinnerAgentID.Int64)
		}

		// Verify agent2 is always agent1 in the competition (since we pass it as agent1 to RecordCompetition)
		if competition.Agent1ID != agent2.ID {
			t.Errorf("Expected agent2 (%d) to be agent1 in competition, got agent %d", agent2.ID, competition.Agent1ID)
		}

		// Verify the other agent is either agent1 or agent3
		if competition.Agent2ID != agent1.ID && competition.Agent2ID != agent3.ID {
			t.Errorf("Expected agent2 in competition to be either agent1 (%d) or agent3 (%d), got %d", 
				agent1.ID, agent3.ID, competition.Agent2ID)
		}
	}

	// Verify agent ELO ratings were updated
	updatedAgent2, err := queries.GetAgent(ctx, agent2.ID)
	if err != nil {
		t.Fatalf("Failed to get updated agent2: %v", err)
	}

	// Agent2 should have higher rating (won 2 games)
	if !updatedAgent2.EloRating.Valid || updatedAgent2.EloRating.Float64 <= DefaultELORating {
		t.Errorf("Agent2 ELO should be higher than default (%v), got %v", DefaultELORating, updatedAgent2.EloRating)
	}

	// Agent2 should have played 2 games and won 2
	if !updatedAgent2.GamesPlayed.Valid || updatedAgent2.GamesPlayed.Int64 != 2 {
		t.Errorf("Agent2 should have played 2 games, got %v", updatedAgent2.GamesPlayed)
	}

	if !updatedAgent2.Wins.Valid || updatedAgent2.Wins.Int64 != 2 {
		t.Errorf("Agent2 should have 2 wins, got %v", updatedAgent2.Wins)
	}

	// Test duplicate prevention - processing again should create no new competitions
	result2, err := eloCalc.ProcessTaskCompetitionsWithWinner(ctx, task.ID, agent2.ID)
	if err != nil {
		t.Fatalf("Failed to process task competitions with winner (second time): %v", err)
	}

	if result2.TotalCompetitions != 0 {
		t.Errorf("Expected 0 competitions on second processing, got %d", result2.TotalCompetitions)
	}

	if result2.NewCompetitions != 0 {
		t.Errorf("Expected 0 new competitions on second processing, got %d", result2.NewCompetitions)
	}
}

func createTestAgentsForMergeELO(t *testing.T, ctx context.Context) (db.Agent, db.Agent, db.Agent, db.Root) {
	// Create a test root
	root, err := queries.CreateRoot(ctx, db.CreateRootParams{
		LocalPort:   "8080",
		ExternalUrl: sql.NullString{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create test root: %v", err)
	}

	// Create three test agents
	agent1, err := queries.CreateAgent(ctx, db.CreateAgentParams{
		RootID:  root.ID,
		Name:    "MergeTestAgent1",
		Command: "merge-test-agent-1",
		Params:  "",
	})
	if err != nil {
		t.Fatalf("Failed to create test agent1: %v", err)
	}

	agent2, err := queries.CreateAgent(ctx, db.CreateAgentParams{
		RootID:  root.ID,
		Name:    "MergeTestAgent2",
		Command: "merge-test-agent-2",
		Params:  "",
	})
	if err != nil {
		t.Fatalf("Failed to create test agent2: %v", err)
	}

	agent3, err := queries.CreateAgent(ctx, db.CreateAgentParams{
		RootID:  root.ID,
		Name:    "MergeTestAgent3",
		Command: "merge-test-agent-3",
		Params:  "",
	})
	if err != nil {
		t.Fatalf("Failed to create test agent3: %v", err)
	}

	return agent1, agent2, agent3, root
}
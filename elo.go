package main

import (
	"context"
	"database/sql"
	"fmt"
	"math"

	"remote-code/db"
)

const (
	DefaultELORating = 1500.0
	DefaultKFactor   = 32.0
	MinKFactor       = 16.0
	MaxKFactor       = 64.0
)

type ELOCalculator struct {
	queries *db.Queries
}

func NewELOCalculator(queries *db.Queries) *ELOCalculator {
	return &ELOCalculator{
		queries: queries,
	}
}

type MatchResult int

const (
	Agent1Wins MatchResult = iota
	Agent2Wins
	Draw
)

type ELOUpdateResult struct {
	Agent1NewRating float64
	Agent2NewRating float64
	Agent1Change    float64
	Agent2Change    float64
}

func (e *ELOCalculator) CalculateELO(agent1Rating, agent2Rating float64, result MatchResult, kFactor float64) ELOUpdateResult {
	expectedScore1 := 1.0 / (1.0 + math.Pow(10, (agent2Rating-agent1Rating)/400.0))
	expectedScore2 := 1.0 - expectedScore1

	var actualScore1, actualScore2 float64
	switch result {
	case Agent1Wins:
		actualScore1, actualScore2 = 1.0, 0.0
	case Agent2Wins:
		actualScore1, actualScore2 = 0.0, 1.0
	case Draw:
		actualScore1, actualScore2 = 0.5, 0.5
	}

	agent1Change := kFactor * (actualScore1 - expectedScore1)
	agent2Change := kFactor * (actualScore2 - expectedScore2)

	return ELOUpdateResult{
		Agent1NewRating: agent1Rating + agent1Change,
		Agent2NewRating: agent2Rating + agent2Change,
		Agent1Change:    agent1Change,
		Agent2Change:    agent2Change,
	}
}

func (e *ELOCalculator) DetermineKFactor(gamesPlayed int64, currentRating float64) float64 {
	if gamesPlayed < 30 {
		return MaxKFactor
	}
	if currentRating < 2100 {
		return DefaultKFactor
	}
	return MinKFactor
}

type CompetitionParams struct {
	TaskID            int64
	Agent1ID          int64
	Agent2ID          int64
	Agent1ExecutionID int64
	Agent2ExecutionID int64
	Result            MatchResult
	Notes             string
}

func (e *ELOCalculator) RecordCompetition(ctx context.Context, params CompetitionParams) (*db.AgentCompetition, error) {
	agent1, err := e.queries.GetAgent(ctx, params.Agent1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent1: %w", err)
	}

	agent2, err := e.queries.GetAgent(ctx, params.Agent2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent2: %w", err)
	}

	agent1Rating := DefaultELORating
	if agent1.EloRating.Valid {
		agent1Rating = agent1.EloRating.Float64
	}

	agent2Rating := DefaultELORating
	if agent2.EloRating.Valid {
		agent2Rating = agent2.EloRating.Float64
	}

	agent1Games := int64(0)
	if agent1.GamesPlayed.Valid {
		agent1Games = agent1.GamesPlayed.Int64
	}

	agent2Games := int64(0)
	if agent2.GamesPlayed.Valid {
		agent2Games = agent2.GamesPlayed.Int64
	}

	kFactor1 := e.DetermineKFactor(agent1Games, agent1Rating)
	kFactor2 := e.DetermineKFactor(agent2Games, agent2Rating)
	avgKFactor := (kFactor1 + kFactor2) / 2

	eloResult := e.CalculateELO(agent1Rating, agent2Rating, params.Result, avgKFactor)

	var winnerID sql.NullInt64
	switch params.Result {
	case Agent1Wins:
		winnerID = sql.NullInt64{Valid: true, Int64: params.Agent1ID}
	case Agent2Wins:
		winnerID = sql.NullInt64{Valid: true, Int64: params.Agent2ID}
	case Draw:
		winnerID = sql.NullInt64{Valid: false}
	}

	competition, err := e.queries.CreateCompetition(ctx, db.CreateCompetitionParams{
		TaskID:            params.TaskID,
		Agent1ID:          params.Agent1ID,
		Agent2ID:          params.Agent2ID,
		Agent1ExecutionID: params.Agent1ExecutionID,
		Agent2ExecutionID: params.Agent2ExecutionID,
		WinnerAgentID:     winnerID,
		Agent1EloBefore:   agent1Rating,
		Agent2EloBefore:   agent2Rating,
		Agent1EloAfter:    eloResult.Agent1NewRating,
		Agent2EloAfter:    eloResult.Agent2NewRating,
		KFactor:           sql.NullFloat64{Valid: true, Float64: avgKFactor},
		CompetitionType:   sql.NullString{Valid: true, String: "head_to_head"},
		Notes:             sql.NullString{Valid: params.Notes != "", String: params.Notes},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create competition record: %w", err)
	}

	agent1Wins := int64(0)
	agent1Losses := int64(0)
	agent1Draws := int64(0)
	
	agent2Wins := int64(0)
	agent2Losses := int64(0)
	agent2Draws := int64(0)

	switch params.Result {
	case Agent1Wins:
		agent1Wins = 1
		agent2Losses = 1
	case Agent2Wins:
		agent1Losses = 1
		agent2Wins = 1
	case Draw:
		agent1Draws = 1
		agent2Draws = 1
	}

	_, err = e.queries.UpdateAgentELO(ctx, db.UpdateAgentELOParams{
		EloRating: sql.NullFloat64{Valid: true, Float64: eloResult.Agent1NewRating},
		Wins:      sql.NullInt64{Valid: true, Int64: agent1Wins},
		Losses:    sql.NullInt64{Valid: true, Int64: agent1Losses},
		Draws:     sql.NullInt64{Valid: true, Int64: agent1Draws},
		ID:        params.Agent1ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update agent1 ELO: %w", err)
	}

	_, err = e.queries.UpdateAgentELO(ctx, db.UpdateAgentELOParams{
		EloRating: sql.NullFloat64{Valid: true, Float64: eloResult.Agent2NewRating},
		Wins:      sql.NullInt64{Valid: true, Int64: agent2Wins},
		Losses:    sql.NullInt64{Valid: true, Int64: agent2Losses},
		Draws:     sql.NullInt64{Valid: true, Int64: agent2Draws},
		ID:        params.Agent2ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update agent2 ELO: %w", err)
	}

	return &competition, nil
}

func (e *ELOCalculator) DetermineWinner(agent1Status, agent2Status string) (MatchResult, error) {
	if agent1Status == "completed" && agent2Status == "completed" {
		return Draw, nil
	}
	if agent1Status == "completed" && agent2Status != "completed" {
		return Agent1Wins, nil
	}
	if agent2Status == "completed" && agent1Status != "completed" {
		return Agent2Wins, nil
	}
	return Draw, fmt.Errorf("unable to determine winner: both agents have status %s and %s", agent1Status, agent2Status)
}

type TaskCompetitionResult struct {
	TaskID            int64
	Competitions      []db.AgentCompetition
	TotalCompetitions int
	NewCompetitions   int
}

// ProcessTaskCompetitionsWithWinner creates competitions between a specific winner agent and all other agents for a task
func (e *ELOCalculator) ProcessTaskCompetitionsWithWinner(ctx context.Context, taskID int64, winnerAgentID int64) (TaskCompetitionResult, error) {
	result := TaskCompetitionResult{
		TaskID: taskID,
	}

	// Get all executions for this task
	executions, err := e.queries.ListTaskExecutionsByTaskID(ctx, taskID)
	if err != nil {
		return result, fmt.Errorf("failed to list task executions: %w", err)
	}

	// Find the winner's execution and all other executions
	var winnerExecution *db.TaskExecution
	var otherExecutions []db.TaskExecution

	for i := range executions {
		if executions[i].AgentID == winnerAgentID {
			winnerExecution = &executions[i]
		} else {
			otherExecutions = append(otherExecutions, executions[i])
		}
	}

	if winnerExecution == nil {
		return result, fmt.Errorf("winner agent %d has no execution for task %d", winnerAgentID, taskID)
	}

	// Create competitions: winner vs each other agent
	for _, otherExecution := range otherExecutions {
		// Check if competition already exists for this task and agent pair
		existing, err := e.queries.GetExistingCompetition(ctx, db.GetExistingCompetitionParams{
			TaskID:            taskID,
			Agent1ID:          winnerAgentID,
			Agent2ID:          otherExecution.AgentID,
			Agent1ExecutionID: winnerExecution.ID,
			Agent2ExecutionID: otherExecution.ID,
		})
		if err == nil && existing.ID != 0 {
			continue // Competition already exists
		}

		// Also check reverse order
		existing, err = e.queries.GetExistingCompetition(ctx, db.GetExistingCompetitionParams{
			TaskID:            taskID,
			Agent1ID:          otherExecution.AgentID,
			Agent2ID:          winnerAgentID,
			Agent1ExecutionID: otherExecution.ID,
			Agent2ExecutionID: winnerExecution.ID,
		})
		if err == nil && existing.ID != 0 {
			continue // Competition already exists
		}

		// Record competition with winner agent winning
		competition, err := e.RecordCompetition(ctx, CompetitionParams{
			TaskID:            taskID,
			Agent1ID:          winnerAgentID,
			Agent2ID:          otherExecution.AgentID,
			Agent1ExecutionID: winnerExecution.ID,
			Agent2ExecutionID: otherExecution.ID,
			Result:            Agent1Wins, // Winner agent always wins
			Notes:             "Competition recorded from merge selection",
		})
		if err != nil {
			return result, fmt.Errorf("failed to record competition between agents %d and %d: %w", winnerAgentID, otherExecution.AgentID, err)
		}
		result.Competitions = append(result.Competitions, *competition)
		result.TotalCompetitions++
		result.NewCompetitions++
	}

	return result, nil
}

func (e *ELOCalculator) ProcessTaskCompetitions(ctx context.Context, taskID int64) (*TaskCompetitionResult, error) {
	executions, err := e.queries.ListTaskExecutionsByTaskID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task executions: %w", err)
	}

	if len(executions) < 2 {
		return &TaskCompetitionResult{
			TaskID:            taskID,
			Competitions:      []db.AgentCompetition{},
			TotalCompetitions: 0,
		}, nil
	}

	var competitions []db.AgentCompetition

	for i := 0; i < len(executions); i++ {
		for j := i + 1; j < len(executions); j++ {
			if executions[i].AgentID == executions[j].AgentID {
				continue
			}

			existing, err := e.queries.GetExistingCompetition(ctx, db.GetExistingCompetitionParams{
				TaskID:            taskID,
				Agent1ID:          executions[i].AgentID,
				Agent2ID:          executions[j].AgentID,
				Agent1ExecutionID: executions[i].ID,
				Agent2ExecutionID: executions[j].ID,
			})
			if err == nil && existing.ID != 0 {
				continue
			}

			result, err := e.DetermineWinner(executions[i].Status, executions[j].Status)
			if err != nil {
				continue
			}

			competition, err := e.RecordCompetition(ctx, CompetitionParams{
				TaskID:            taskID,
				Agent1ID:          executions[i].AgentID,
				Agent2ID:          executions[j].AgentID,
				Agent1ExecutionID: executions[i].ID,
				Agent2ExecutionID: executions[j].ID,
				Result:            result,
				Notes:             fmt.Sprintf("Auto-generated from task executions (agent1: %s, agent2: %s)", executions[i].Status, executions[j].Status),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to record competition: %w", err)
			}

			competitions = append(competitions, *competition)
		}
	}

	return &TaskCompetitionResult{
		TaskID:            taskID,
		Competitions:      competitions,
		TotalCompetitions: len(competitions),
	}, nil
}
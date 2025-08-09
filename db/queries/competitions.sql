-- name: CreateCompetition :one
INSERT INTO agent_competitions (
    task_id, agent1_id, agent2_id, 
    agent1_execution_id, agent2_execution_id,
    winner_agent_id, agent1_elo_before, agent2_elo_before,
    agent1_elo_after, agent2_elo_after, k_factor,
    competition_type, notes
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetCompetition :one
SELECT * FROM agent_competitions
WHERE id = ?;

-- name: ListCompetitions :many
SELECT * FROM agent_competitions
ORDER BY created_at DESC;

-- name: ListCompetitionsByTask :many
SELECT * FROM agent_competitions
WHERE task_id = ?
ORDER BY created_at DESC;

-- name: ListCompetitionsByAgent :many
SELECT * FROM agent_competitions
WHERE agent1_id = ? OR agent2_id = ?
ORDER BY created_at DESC;

-- name: GetCompetitionHistory :many
SELECT * FROM competition_history
ORDER BY created_at DESC;

-- name: GetCompetitionHistoryByAgent :many
SELECT * FROM competition_history
WHERE agent1_id = ? OR agent2_id = ?
ORDER BY created_at DESC;

-- name: GetHeadToHeadRecord :one
SELECT 
    COUNT(*) as total_games,
    SUM(CASE WHEN winner_agent_id = ? THEN 1 ELSE 0 END) as agent1_wins,
    SUM(CASE WHEN winner_agent_id = ? THEN 1 ELSE 0 END) as agent2_wins,
    SUM(CASE WHEN winner_agent_id IS NULL THEN 1 ELSE 0 END) as draws
FROM agent_competitions
WHERE (agent1_id = ? AND agent2_id = ?) OR (agent1_id = ? AND agent2_id = ?);

-- name: GetExistingCompetition :one
SELECT * FROM agent_competitions
WHERE task_id = ? AND 
      ((agent1_id = ? AND agent2_id = ? AND agent1_execution_id = ? AND agent2_execution_id = ?) OR
       (agent1_id = ? AND agent2_id = ? AND agent1_execution_id = ? AND agent2_execution_id = ?))
LIMIT 1;
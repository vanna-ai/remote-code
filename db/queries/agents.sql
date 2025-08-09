-- name: CreateAgent :one
INSERT INTO agents (root_id, name, command, params)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetAgent :one
SELECT * FROM agents
WHERE id = ?;

-- name: GetAgentsByRootID :many
SELECT * FROM agents
WHERE root_id = ?
ORDER BY name;

-- name: UpdateAgent :one
UPDATE agents
SET name = ?, command = ?, params = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteAgent :exec
DELETE FROM agents WHERE id = ?;

-- name: ListAgents :many
SELECT * FROM agents
ORDER BY name;

-- name: UpdateAgentELO :one
UPDATE agents
SET elo_rating = ?, games_played = games_played + 1, 
    wins = wins + ?, losses = losses + ?, draws = draws + ?,
    last_competed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: GetAgentLeaderboard :many
SELECT * FROM agent_leaderboard;

-- name: GetAgentELOHistory :many
SELECT 
    c.created_at,
    c.agent1_elo_after as elo_rating,
    (c.agent1_elo_after - c.agent1_elo_before) as elo_change
FROM agent_competitions c
WHERE c.agent1_id = ?
UNION ALL
SELECT 
    c.created_at,
    c.agent2_elo_after as elo_rating,
    (c.agent2_elo_after - c.agent2_elo_before) as elo_change
FROM agent_competitions c
WHERE c.agent2_id = ?
ORDER BY created_at ASC;
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
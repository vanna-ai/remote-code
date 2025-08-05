-- name: CreateProject :one
INSERT INTO projects (root_id, name)
VALUES (?, ?)
RETURNING *;

-- name: GetProject :one
SELECT * FROM projects
WHERE id = ?;

-- name: GetProjectsByRootID :many
SELECT * FROM projects
WHERE root_id = ?
ORDER BY name;

-- name: UpdateProject :one
UPDATE projects
SET name = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = ?;
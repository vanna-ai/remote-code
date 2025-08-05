-- name: CreateRoot :one
INSERT INTO roots (local_port, external_url)
VALUES (?, ?)
RETURNING *;

-- name: GetRoot :one
SELECT * FROM roots
WHERE id = ?;

-- name: GetRootWithAgentsAndProjects :one
SELECT 
    r.id,
    r.local_port,
    r.external_url,
    r.created_at,
    r.updated_at
FROM roots r
WHERE r.id = ?;

-- name: UpdateRoot :one
UPDATE roots
SET local_port = ?, external_url = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteRoot :exec
DELETE FROM roots WHERE id = ?;
-- name: CreateDirectoryDevServer :one
INSERT INTO directory_dev_servers (base_directory_id, tmux_session_id, status)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetDirectoryDevServer :one
SELECT * FROM directory_dev_servers
WHERE id = ?;

-- name: GetDirectoryDevServerByDirectoryID :one
SELECT * FROM directory_dev_servers
WHERE base_directory_id = ?;

-- name: ListRunningDirectoryDevServers :many
SELECT
    dds.*,
    bd.path as directory_path,
    bd.base_directory_id as directory_name,
    p.id as project_id,
    p.name as project_name
FROM directory_dev_servers dds
JOIN base_directories bd ON dds.base_directory_id = bd.id
JOIN projects p ON bd.project_id = p.id
WHERE dds.status = 'running'
ORDER BY dds.created_at DESC;

-- name: UpdateDirectoryDevServerStatus :one
UPDATE directory_dev_servers
SET status = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteDirectoryDevServer :exec
DELETE FROM directory_dev_servers WHERE id = ?;

-- name: DeleteDirectoryDevServerByDirectoryID :exec
DELETE FROM directory_dev_servers WHERE base_directory_id = ?;

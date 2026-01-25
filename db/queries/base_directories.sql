-- name: CreateBaseDirectory :one
INSERT INTO base_directories (
    project_id,
    base_directory_id,
    path,
    git_initialized,
    setup_commands,
    teardown_commands,
    dev_server_setup_commands,
    dev_server_teardown_commands
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetBaseDirectory :one
SELECT * FROM base_directories
WHERE id = ?;

-- name: GetBaseDirectoryByProjectAndID :one
SELECT * FROM base_directories
WHERE project_id = ? AND base_directory_id = ?;

-- name: GetBaseDirectoriesByProjectID :many
SELECT * FROM base_directories
WHERE project_id = ?
ORDER BY base_directory_id;

-- name: UpdateBaseDirectory :one
UPDATE base_directories
SET
    path = ?,
    git_initialized = ?,
    setup_commands = ?,
    teardown_commands = ?,
    dev_server_setup_commands = ?,
    dev_server_teardown_commands = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteBaseDirectory :exec
DELETE FROM base_directories WHERE id = ?;

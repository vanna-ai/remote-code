-- name: CreateTask :one
INSERT INTO tasks (project_id, base_directory_id, title, description, status)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = ?;

-- name: GetTasksByProjectID :many
SELECT * FROM tasks
WHERE project_id = ?
ORDER BY title;

-- name: GetTaskWithBaseDirectory :one
SELECT 
    t.*,
    bd.path as base_directory_path,
    bd.git_initialized,
    bd.worktree_setup_commands,
    bd.worktree_teardown_commands,
    bd.dev_server_setup_commands,
    bd.dev_server_teardown_commands
FROM tasks t
JOIN base_directories bd ON t.base_directory_id = bd.base_directory_id
WHERE t.id = ?;

-- name: UpdateTask :one
UPDATE tasks
SET 
    title = ?,
    description = ?,
    status = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = ?;

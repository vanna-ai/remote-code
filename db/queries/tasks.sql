-- name: CreateTask :one
INSERT INTO tasks (project_id, title, description, worktree_id)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = ?;

-- name: GetTasksByProjectID :many
SELECT * FROM tasks
WHERE project_id = ?
ORDER BY title;

-- name: GetTaskWithWorktree :one
SELECT 
    t.*,
    w.base_directory_id,
    w.path as worktree_path,
    w.agent_tmux_id,
    w.dev_server_tmux_id,
    w.external_url as worktree_external_url
FROM tasks t
LEFT JOIN worktrees w ON t.worktree_id = w.id
WHERE t.id = ?;

-- name: UpdateTask :one
UPDATE tasks
SET 
    title = ?,
    description = ?,
    worktree_id = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = ?;
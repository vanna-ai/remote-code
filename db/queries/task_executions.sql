-- name: CreateTaskExecution :one
INSERT INTO task_executions (task_id, agent_id, worktree_id, status)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetTaskExecution :one
SELECT * FROM task_executions
WHERE id = ?;

-- name: GetTaskExecutionsByTaskID :many
SELECT * FROM task_executions
WHERE task_id = ?
ORDER BY created_at;

-- name: GetTaskExecutionsByAgentID :many
SELECT * FROM task_executions
WHERE agent_id = ?
ORDER BY created_at DESC;

-- name: GetTaskExecutionWithDetails :one
SELECT 
    te.*,
    t.title as task_title,
    t.description as task_description,
    a.name as agent_name,
    a.command as agent_command,
    w.path as worktree_path,
    w.base_directory_id
FROM task_executions te
JOIN tasks t ON te.task_id = t.id
JOIN agents a ON te.agent_id = a.id
JOIN worktrees w ON te.worktree_id = w.id
WHERE te.id = ?;

-- name: UpdateTaskExecutionStatus :one
UPDATE task_executions
SET 
    status = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteTaskExecution :exec
DELETE FROM task_executions WHERE id = ?;

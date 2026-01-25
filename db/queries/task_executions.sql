-- name: CreateTaskExecution :one
INSERT INTO task_executions (task_id, agent_id, status, agent_tmux_id, dev_server_tmux_id)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetTaskExecution :one
SELECT * FROM task_executions
WHERE id = ?;

-- name: GetTaskExecutionsByTaskID :many
SELECT
    te.*,
    a.name as agent_name
FROM task_executions te
JOIN agents a ON te.agent_id = a.id
WHERE te.task_id = ?
ORDER BY te.created_at;

-- name: GetTaskExecutionsByAgentID :many
SELECT * FROM task_executions
WHERE agent_id = ?
ORDER BY created_at DESC;

-- name: GetTaskExecutionWithDetails :one
SELECT
    te.*,
    t.title as task_title,
    t.description as task_description,
    t.base_directory_id,
    a.name as agent_name,
    a.command as agent_command,
    bd.path as base_directory_path,
    p.id as project_id,
    p.name as project_name
FROM task_executions te
JOIN tasks t ON te.task_id = t.id
JOIN agents a ON te.agent_id = a.id
JOIN base_directories bd ON t.base_directory_id = bd.base_directory_id AND t.project_id = bd.project_id
JOIN projects p ON t.project_id = p.id
WHERE te.id = ?;

-- name: UpdateTaskExecutionStatus :one
UPDATE task_executions
SET
    status = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateTaskExecutionTmux :one
UPDATE task_executions
SET
    agent_tmux_id = ?,
    dev_server_tmux_id = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: ListTaskExecutions :many
SELECT
    te.*,
    t.title as task_title,
    a.name as agent_name,
    p.id as project_id,
    p.name as project_name
FROM task_executions te
JOIN tasks t ON te.task_id = t.id
JOIN agents a ON te.agent_id = a.id
JOIN projects p ON t.project_id = p.id
ORDER BY te.created_at DESC;

-- name: DeleteTaskExecution :exec
DELETE FROM task_executions WHERE id = ?;

-- name: ListTaskExecutionsByTaskID :many
SELECT * FROM task_executions
WHERE task_id = ?
ORDER BY created_at;

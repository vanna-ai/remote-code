-- name: CreateWorktree :one
INSERT INTO worktrees (base_directory_id, path, agent_tmux_id, dev_server_tmux_id, external_url)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetWorktree :one
SELECT * FROM worktrees
WHERE id = ?;

-- name: GetWorktreesByBaseDirectoryID :many
SELECT * FROM worktrees
WHERE base_directory_id = ?
ORDER BY path;

-- name: UpdateWorktree :one
UPDATE worktrees
SET 
    path = ?,
    agent_tmux_id = ?,
    dev_server_tmux_id = ?,
    external_url = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteWorktree :exec
DELETE FROM worktrees WHERE id = ?;
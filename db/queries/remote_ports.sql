-- name: CreateRemotePort :one
INSERT INTO remote_ports (port, tmux_session_id, status)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetRemotePort :one
SELECT * FROM remote_ports
WHERE id = ?;

-- name: ListRemotePorts :many
SELECT * FROM remote_ports
ORDER BY created_at DESC;

-- name: ListActiveRemotePorts :many
SELECT * FROM remote_ports
WHERE status IN ('starting', 'connected')
ORDER BY created_at DESC;

-- name: UpdateRemotePortExternalUrl :one
UPDATE remote_ports
SET external_url = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateRemotePortStatus :one
UPDATE remote_ports
SET status = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteRemotePort :exec
DELETE FROM remote_ports WHERE id = ?;

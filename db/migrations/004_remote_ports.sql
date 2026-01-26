-- Remote ports table for cloudflared tunnel management
CREATE TABLE IF NOT EXISTS remote_ports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    port INTEGER NOT NULL,
    tmux_session_id TEXT NOT NULL,
    external_url TEXT,
    status TEXT NOT NULL DEFAULT 'starting',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_remote_ports_status ON remote_ports(status);

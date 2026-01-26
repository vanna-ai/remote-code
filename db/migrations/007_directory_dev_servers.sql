-- Directory dev servers table for managing dev servers tied to base directories
CREATE TABLE IF NOT EXISTS directory_dev_servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_directory_id INTEGER NOT NULL UNIQUE,
    tmux_session_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'running',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (base_directory_id) REFERENCES base_directories(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_directory_dev_servers_base_directory_id ON directory_dev_servers(base_directory_id);
CREATE INDEX IF NOT EXISTS idx_directory_dev_servers_status ON directory_dev_servers(status);

-- Initial schema for web terminal application

CREATE TABLE IF NOT EXISTS roots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    local_port TEXT NOT NULL,
    external_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS agents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    command TEXT NOT NULL,
    params TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (root_id) REFERENCES roots(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (root_id) REFERENCES roots(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS base_directories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    base_directory_id TEXT NOT NULL,
    path TEXT NOT NULL,
    git_initialized BOOLEAN NOT NULL DEFAULT FALSE,
    worktree_setup_commands TEXT NOT NULL DEFAULT '',
    worktree_teardown_commands TEXT NOT NULL DEFAULT '',
    dev_server_setup_commands TEXT NOT NULL DEFAULT '',
    dev_server_teardown_commands TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE(project_id, base_directory_id)
);

CREATE TABLE IF NOT EXISTS worktrees (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_directory_id TEXT NOT NULL,
    path TEXT NOT NULL,
    agent_tmux_id TEXT,
    dev_server_tmux_id TEXT,
    external_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    worktree_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (worktree_id) REFERENCES worktrees(id) ON DELETE SET NULL
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_agents_root_id ON agents(root_id);
CREATE INDEX IF NOT EXISTS idx_projects_root_id ON projects(root_id);
CREATE INDEX IF NOT EXISTS idx_base_directories_project_id ON base_directories(project_id);
CREATE INDEX IF NOT EXISTS idx_base_directories_base_directory_id ON base_directories(base_directory_id);
CREATE INDEX IF NOT EXISTS idx_worktrees_base_directory_id ON worktrees(base_directory_id);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_worktree_id ON tasks(worktree_id);
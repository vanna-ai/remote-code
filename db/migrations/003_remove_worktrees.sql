-- Migration to remove git worktree isolation from task execution
-- Tasks will now run directly in the base directory instead of separate worktrees

-- Step 1: Add tmux columns to task_executions (migrating from worktrees table)
ALTER TABLE task_executions ADD COLUMN agent_tmux_id TEXT;
ALTER TABLE task_executions ADD COLUMN dev_server_tmux_id TEXT;

-- Step 2: Migrate tmux data from worktrees to task_executions
UPDATE task_executions
SET agent_tmux_id = (SELECT agent_tmux_id FROM worktrees WHERE worktrees.id = task_executions.worktree_id),
    dev_server_tmux_id = (SELECT dev_server_tmux_id FROM worktrees WHERE worktrees.id = task_executions.worktree_id);

-- Step 3: Create new task_executions table without worktree_id
CREATE TABLE task_executions_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    agent_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'running',
    agent_tmux_id TEXT,
    dev_server_tmux_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

-- Step 4: Copy data to new table
INSERT INTO task_executions_new (id, task_id, agent_id, status, agent_tmux_id, dev_server_tmux_id, created_at, updated_at)
SELECT id, task_id, agent_id, status, agent_tmux_id, dev_server_tmux_id, created_at, updated_at
FROM task_executions;

-- Step 5: Drop old table and rename new one
DROP TABLE task_executions;
ALTER TABLE task_executions_new RENAME TO task_executions;

-- Step 6: Recreate indexes for task_executions (without worktree_id index)
CREATE INDEX IF NOT EXISTS idx_task_executions_task_id ON task_executions(task_id);
CREATE INDEX IF NOT EXISTS idx_task_executions_agent_id ON task_executions(agent_id);

-- Step 7: Drop worktrees table and its index
DROP INDEX IF EXISTS idx_worktrees_base_directory_id;
DROP TABLE IF EXISTS worktrees;

-- Step 8: Rename columns in base_directories (worktree_* -> setup/teardown)
-- SQLite doesn't support RENAME COLUMN in older versions, so we recreate the table
CREATE TABLE base_directories_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    base_directory_id TEXT NOT NULL,
    path TEXT NOT NULL,
    git_initialized BOOLEAN NOT NULL DEFAULT FALSE,
    setup_commands TEXT NOT NULL DEFAULT '',
    teardown_commands TEXT NOT NULL DEFAULT '',
    dev_server_setup_commands TEXT NOT NULL DEFAULT '',
    dev_server_teardown_commands TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE(project_id, base_directory_id)
);

INSERT INTO base_directories_new (id, project_id, base_directory_id, path, git_initialized, setup_commands, teardown_commands, dev_server_setup_commands, dev_server_teardown_commands, created_at, updated_at)
SELECT id, project_id, base_directory_id, path, git_initialized, worktree_setup_commands, worktree_teardown_commands, dev_server_setup_commands, dev_server_teardown_commands, created_at, updated_at
FROM base_directories;

DROP TABLE base_directories;
ALTER TABLE base_directories_new RENAME TO base_directories;

-- Recreate base_directories indexes
CREATE INDEX IF NOT EXISTS idx_base_directories_project_id ON base_directories(project_id);
CREATE INDEX IF NOT EXISTS idx_base_directories_base_directory_id ON base_directories(base_directory_id);

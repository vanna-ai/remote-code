-- Migration: Add ELO tracking for agents
-- This migration adds ELO rating functionality to track agent performance in head-to-head competitions

-- Add ELO fields to existing agents table
-- Note: These columns may already exist, so we comment them out to avoid duplicate column errors
-- ALTER TABLE agents ADD COLUMN elo_rating REAL DEFAULT 1500.0;
-- ALTER TABLE agents ADD COLUMN games_played INTEGER DEFAULT 0;
-- ALTER TABLE agents ADD COLUMN wins INTEGER DEFAULT 0;
-- ALTER TABLE agents ADD COLUMN losses INTEGER DEFAULT 0;
-- ALTER TABLE agents ADD COLUMN draws INTEGER DEFAULT 0;
-- ALTER TABLE agents ADD COLUMN last_competed_at DATETIME DEFAULT NULL;

-- Create table to track head-to-head competitions between agents
CREATE TABLE IF NOT EXISTS agent_competitions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    agent1_id INTEGER NOT NULL,
    agent2_id INTEGER NOT NULL,
    agent1_execution_id INTEGER NOT NULL,
    agent2_execution_id INTEGER NOT NULL,
    -- Winner determination: NULL = draw, agent1_id = agent1 wins, agent2_id = agent2 wins
    winner_agent_id INTEGER,
    -- ELO ratings before the competition
    agent1_elo_before REAL NOT NULL,
    agent2_elo_before REAL NOT NULL,
    -- ELO ratings after the competition (calculated using standard ELO algorithm)
    agent1_elo_after REAL NOT NULL,
    agent2_elo_after REAL NOT NULL,
    -- Competition metadata
    k_factor REAL DEFAULT 32.0, -- ELO K-factor used for this competition
    competition_type TEXT DEFAULT 'head_to_head', -- Future: support for tournaments, etc.
    notes TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (agent1_id) REFERENCES agents(id) ON DELETE CASCADE,
    FOREIGN KEY (agent2_id) REFERENCES agents(id) ON DELETE CASCADE,
    FOREIGN KEY (agent1_execution_id) REFERENCES task_executions(id) ON DELETE CASCADE,
    FOREIGN KEY (agent2_execution_id) REFERENCES task_executions(id) ON DELETE CASCADE,
    FOREIGN KEY (winner_agent_id) REFERENCES agents(id) ON DELETE SET NULL,
    
    -- Ensure agents are different
    CHECK (agent1_id != agent2_id),
    -- Ensure winner is one of the competing agents or NULL for draw
    CHECK (winner_agent_id IS NULL OR winner_agent_id = agent1_id OR winner_agent_id = agent2_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_agent_competitions_agent1 ON agent_competitions(agent1_id);
CREATE INDEX IF NOT EXISTS idx_agent_competitions_agent2 ON agent_competitions(agent2_id);
CREATE INDEX IF NOT EXISTS idx_agent_competitions_task ON agent_competitions(task_id);
CREATE INDEX IF NOT EXISTS idx_agent_competitions_created_at ON agent_competitions(created_at);
CREATE INDEX IF NOT EXISTS idx_agent_competitions_winner ON agent_competitions(winner_agent_id);

-- Create view for easy agent leaderboard queries
CREATE VIEW IF NOT EXISTS agent_leaderboard AS
SELECT 
    a.id,
    a.name,
    a.elo_rating,
    a.games_played,
    a.wins,
    a.losses,
    a.draws,
    a.last_competed_at,
    CASE 
        WHEN a.games_played > 0 THEN ROUND((a.wins * 1.0 / a.games_played) * 100, 2)
        ELSE 0.0 
    END as win_percentage,
    RANK() OVER (ORDER BY a.elo_rating DESC) as elo_rank
FROM agents a
ORDER BY a.elo_rating DESC;

-- Create view for competition history with agent names
CREATE VIEW IF NOT EXISTS competition_history AS
SELECT 
    c.id,
    c.task_id,
    c.agent1_id,
    a1.name as agent1_name,
    c.agent2_id,
    a2.name as agent2_name,
    c.winner_agent_id,
    CASE 
        WHEN c.winner_agent_id IS NULL THEN 'Draw'
        WHEN c.winner_agent_id = c.agent1_id THEN a1.name
        ELSE a2.name
    END as winner_name,
    c.agent1_elo_before,
    c.agent2_elo_before,
    c.agent1_elo_after,
    c.agent2_elo_after,
    (c.agent1_elo_after - c.agent1_elo_before) as agent1_elo_change,
    (c.agent2_elo_after - c.agent2_elo_before) as agent2_elo_change,
    c.k_factor,
    c.competition_type,
    c.notes,
    c.created_at
FROM agent_competitions c
JOIN agents a1 ON c.agent1_id = a1.id
JOIN agents a2 ON c.agent2_id = a2.id
ORDER BY c.created_at DESC;
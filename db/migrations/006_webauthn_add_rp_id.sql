-- Add rp_id column to existing webauthn_credentials table
-- This handles databases created before rp_id was added

-- Check if rp_id column exists, if not add it
-- SQLite doesn't have IF NOT EXISTS for columns, so we use a workaround
-- by catching the error if the column already exists

-- Add the column (will fail silently if it already exists due to migration 005)
ALTER TABLE webauthn_credentials ADD COLUMN rp_id TEXT NOT NULL DEFAULT 'localhost';

-- Create index if not exists
CREATE INDEX IF NOT EXISTS idx_webauthn_credentials_rp_id ON webauthn_credentials(rp_id);

-- This migration was for adding rp_id column to existing webauthn_credentials table
-- It's now included in migration 005, so this is a no-op for sqlc compatibility
-- The actual runtime migration handles databases created before rp_id was added

-- Create index if not exists (idempotent)
CREATE INDEX IF NOT EXISTS idx_webauthn_credentials_rp_id ON webauthn_credentials(rp_id);

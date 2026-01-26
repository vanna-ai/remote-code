-- WebAuthn credentials table for passkey authentication
CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id TEXT PRIMARY KEY,
    rp_id TEXT NOT NULL,
    public_key BLOB NOT NULL,
    attestation_type TEXT NOT NULL,
    transport TEXT,
    aaguid BLOB,
    sign_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for looking up credentials by RP ID
CREATE INDEX IF NOT EXISTS idx_webauthn_credentials_rp_id ON webauthn_credentials(rp_id);

-- Sessions table for authenticated sessions
CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for cleaning up expired sessions
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

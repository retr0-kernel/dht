-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) NOT NULL UNIQUE,
    refresh_token VARCHAR(255) UNIQUE,
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_sessions_user_id ON sessions(user_id);

-- Create index on session_token for authentication
CREATE INDEX idx_sessions_session_token ON sessions(session_token) WHERE is_active = true AND revoked_at IS NULL;

-- Create index on refresh_token for token refresh
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token) WHERE is_active = true AND revoked_at IS NULL;

-- Create index on expires_at for cleanup jobs
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Create index on is_active
CREATE INDEX idx_sessions_is_active ON sessions(is_active) WHERE revoked_at IS NULL;

-- Create composite index for active sessions lookup
CREATE INDEX idx_sessions_user_active ON sessions(user_id, is_active) WHERE revoked_at IS NULL;

-- Create trigger for updated_at
CREATE TRIGGER update_sessions_updated_at 
    BEFORE UPDATE ON sessions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create api_keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);

-- Create index on key_hash for authentication
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash) WHERE is_active = true AND revoked_at IS NULL;

-- Create index on key_prefix for partial key matching
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);

-- Create index on is_active
CREATE INDEX idx_api_keys_is_active ON api_keys(is_active) WHERE revoked_at IS NULL;

-- Create index on expires_at for cleanup jobs
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at) WHERE expires_at IS NOT NULL;

-- Create trigger for updated_at
CREATE TRIGGER update_api_keys_updated_at 
    BEFORE UPDATE ON api_keys 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

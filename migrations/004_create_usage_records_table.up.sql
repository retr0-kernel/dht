-- Create usage_records table
CREATE TABLE IF NOT EXISTS usage_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_id BIGINT REFERENCES api_keys(id) ON DELETE SET NULL,
    operation VARCHAR(50) NOT NULL,
    key_accessed VARCHAR(255),
    request_size_bytes BIGINT NOT NULL DEFAULT 0,
    response_size_bytes BIGINT NOT NULL DEFAULT 0,
    status_code INTEGER NOT NULL,
    duration_ms INTEGER NOT NULL,
    ip_address INET,
    user_agent TEXT,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on user_id for usage queries
CREATE INDEX idx_usage_records_user_id ON usage_records(user_id);

-- Create index on api_key_id for API key usage tracking
CREATE INDEX idx_usage_records_api_key_id ON usage_records(api_key_id);

-- Create index on operation for filtering by operation type
CREATE INDEX idx_usage_records_operation ON usage_records(operation);

-- Create index on created_at for time-based queries
CREATE INDEX idx_usage_records_created_at ON usage_records(created_at DESC);

-- Create composite index for user usage over time
CREATE INDEX idx_usage_records_user_created ON usage_records(user_id, created_at DESC);

-- Create composite index for API key usage over time
CREATE INDEX idx_usage_records_api_key_created ON usage_records(api_key_id, created_at DESC);

-- Create index on status_code for error tracking
CREATE INDEX idx_usage_records_status_code ON usage_records(status_code);

-- Create partial index for errors
CREATE INDEX idx_usage_records_errors ON usage_records(user_id, created_at DESC) 
    WHERE status_code >= 400;

-- Add table partitioning comment for future optimization
COMMENT ON TABLE usage_records IS 'Consider partitioning by created_at (monthly) for better performance with large datasets';

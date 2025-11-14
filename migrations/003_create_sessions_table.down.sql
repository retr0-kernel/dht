-- Drop trigger
DROP TRIGGER IF EXISTS update_sessions_updated_at ON sessions;

-- Drop indexes
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_session_token;
DROP INDEX IF EXISTS idx_sessions_refresh_token;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_is_active;
DROP INDEX IF EXISTS idx_sessions_user_active;

-- Drop sessions table
DROP TABLE IF EXISTS sessions;

-- Drop trigger
DROP TRIGGER IF EXISTS update_api_keys_updated_at ON api_keys;

-- Drop indexes
DROP INDEX IF EXISTS idx_api_keys_user_id;
DROP INDEX IF EXISTS idx_api_keys_key_hash;
DROP INDEX IF EXISTS idx_api_keys_key_prefix;
DROP INDEX IF EXISTS idx_api_keys_is_active;
DROP INDEX IF EXISTS idx_api_keys_expires_at;

-- Drop api_keys table
DROP TABLE IF EXISTS api_keys;

-- Drop indexes
DROP INDEX IF EXISTS idx_usage_records_user_id;
DROP INDEX IF EXISTS idx_usage_records_api_key_id;
DROP INDEX IF EXISTS idx_usage_records_operation;
DROP INDEX IF EXISTS idx_usage_records_created_at;
DROP INDEX IF EXISTS idx_usage_records_user_created;
DROP INDEX IF EXISTS idx_usage_records_api_key_created;
DROP INDEX IF EXISTS idx_usage_records_status_code;
DROP INDEX IF EXISTS idx_usage_records_errors;

-- Drop usage_records table
DROP TABLE IF EXISTS usage_records;

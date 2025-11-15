# Database Migrations

SQL migration files for database schema evolution.

## Migration Files

### 001 - Create Users Table
**Purpose:** Core user authentication and account management

**Tables:**
- `users`: User accounts with email, username, password hash

**Features:**
- Email and username uniqueness constraints
- Soft delete support (`deleted_at`)
- Account verification tracking
- Last login timestamp
- Auto-updating `updated_at` trigger

**Indexes:**
- `idx_users_email`: Fast email lookups
- `idx_users_username`: Fast username lookups
- `idx_users_is_active`: Filter active users

---

### 002 - Create API Keys Table
**Purpose:** API key management for programmatic access

**Tables:**
- `api_keys`: API keys per user with scopes and expiration

**Features:**
- Foreign key to users (CASCADE delete)
- Key stored as bcrypt hash
- Key prefix for quick lookup
- Scopes array for permission control
- Expiration and revocation support
- Last used tracking

**Indexes:**
- `idx_api_keys_user_id`: List keys by user
- `idx_api_keys_key_hash`: Fast auth lookup
- `idx_api_keys_key_prefix`: Partial key matching
- `idx_api_keys_expires_at`: Cleanup expired keys

---

### 003 - Create Sessions Table
**Purpose:** User session management for web/mobile clients

**Tables:**
- `sessions`: JWT session tracking

**Features:**
- Session and refresh tokens
- IP address and user agent tracking
- Session expiration
- Session revocation support

**Indexes:**
- `idx_sessions_user_id`: List sessions by user
- `idx_sessions_session_token`: Fast token lookup
- `idx_sessions_refresh_token`: Token refresh
- `idx_sessions_expires_at`: Cleanup expired sessions

---

### 004 - Create Usage Records Table
**Purpose:** API usage tracking and analytics

**Tables:**
- `usage_records`: Per-request usage data

**Features:**
- Track operation type (GET/PUT/DELETE)
- Request/response sizes
- Latency tracking
- Error logging
- IP and user agent

**Indexes:**
- `idx_usage_records_user_id`: Usage by user
- `idx_usage_records_api_key_id`: Usage by API key
- `idx_usage_records_created_at`: Time-based queries
- `idx_usage_records_errors`: Error tracking

**Note:** Consider partitioning by `created_at` for large datasets

## Database Connection
```bash
DATABASE_URL="postgres://yourdht:yourdhtpass@localhost:5432/dht_db?sslmode=disable"
```

## Running Migrations

### Prerequisites
```bash
# Start PostgreSQL
podman run -d \
  --name yourdht-postgres \
  -e POSTGRES_USER=yourdht \
  -e POSTGRES_PASSWORD=yourdhtpass \
  -e POSTGRES_DB=dht_db \
  -p 5432:5432 \
  postgres:14

# Set environment variable
export DATABASE_URL="postgres://yourdht:yourdhtpass@localhost:5432/dht_db?sslmode=disable"
```

### Apply All Migrations (Up)
```bash
cd migrations

psql $DATABASE_URL -f 001_create_users_table.up.sql
psql $DATABASE_URL -f 002_create_api_keys_table.up.sql
psql $DATABASE_URL -f 003_create_sessions_table.up.sql
psql $DATABASE_URL -f 004_create_usage_records_table.up.sql

cd ..
```

### Rollback All Migrations (Down)
```bash
cd migrations

# Note: Reverse order!
psql $DATABASE_URL -f 004_create_usage_records_table.down.sql
psql $DATABASE_URL -f 003_create_sessions_table.down.sql
psql $DATABASE_URL -f 002_create_api_keys_table.down.sql
psql $DATABASE_URL -f 001_create_users_table.down.sql

cd ..
```

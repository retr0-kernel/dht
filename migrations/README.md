# Database Migrations

SQL migration files for database schema evolution.

## Migration Files

### 001 - Create Users Table
- **Up**: Creates users table with authentication fields
- **Down**: Drops users table and related indexes

### 002 - Create API Keys Table
- **Up**: Creates api_keys table for API authentication
- **Down**: Drops api_keys table and related indexes

### 003 - Create Sessions Table
- **Up**: Creates sessions table for user session management
- **Down**: Drops sessions table and related indexes

### 004 - Create Usage Records Table
- **Up**: Creates usage_records table for tracking API usage
- **Down**: Drops usage_records table and related indexes

## Database Connection
```bash
DATABASE_URL=postgres://yourdht:yourdhtpass@localhost:5432/dht_db?sslmode=disable
```

## Running Migrations Manually

### Apply All Migrations (Up)
```bash
psql $DATABASE_URL -f 001_create_users_table.up.sql
psql $DATABASE_URL -f 002_create_api_keys_table.up.sql
psql $DATABASE_URL -f 003_create_sessions_table.up.sql
psql $DATABASE_URL -f 004_create_usage_records_table.up.sql
```

### Rollback All Migrations (Down)
```bash
psql $DATABASE_URL -f 004_create_usage_records_table.down.sql
psql $DATABASE_URL -f 003_create_sessions_table.down.sql
psql $DATABASE_URL -f 002_create_api_keys_table.down.sql
psql $DATABASE_URL -f 001_create_users_table.down.sql
```

## Using golang-migrate CLI

### Install golang-migrate
```bash
# macOS
brew install golang-migrate

# Or with Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Apply Migrations
```bash
migrate -database "$DATABASE_URL" -path . up
```

### Rollback Migrations
```bash
migrate -database "$DATABASE_URL" -path . down
```

### Check Migration Version
```bash
migrate -database "$DATABASE_URL" -path . version
```

## Schema Overview

### users
- User accounts and authentication
- Soft delete support via `deleted_at`
- Email and username uniqueness
- Account verification tracking

### api_keys
- API key management per user
- Scopes for permission control
- Expiration and revocation support
- Key prefix for identification

### sessions
- User session management
- JWT token storage
- IP and user agent tracking
- Session expiration and revocation

### usage_records
- API usage tracking
- Request/response size monitoring
- Performance metrics (duration)
- Error tracking
- Consider partitioning for large datasets

## Notes

- All tables use `BIGSERIAL` for primary keys
- Timestamps use `TIMESTAMP WITH TIME ZONE`
- Soft deletes implemented where appropriate
- Indexes optimized for common query patterns
- Foreign keys use CASCADE or SET NULL appropriately

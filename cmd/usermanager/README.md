# User Manager Service

The User Manager service handles authentication, user management, and API key operations for the yourdht system.

## Overview

The User Manager provides:
- **User Registration**: Create new user accounts
- **Authentication**: JWT-based login system
- **Session Management**: Secure session tracking
- **API Key Management**: Create and manage API keys for programmatic access
- **API Key Validation**: Validate API keys for other services

## Architecture
![User Manager Architecture](./images/usermanager-architecture.png)

## Database Schema

### users Table
```sql
- id: BIGSERIAL PRIMARY KEY
- email: VARCHAR(255) UNIQUE
- username: VARCHAR(100) UNIQUE
- password_hash: VARCHAR(255)
- is_active: BOOLEAN
- is_verified: BOOLEAN
- created_at: TIMESTAMP
- updated_at: TIMESTAMP
- last_login_at: TIMESTAMP
- deleted_at: TIMESTAMP
```

### api_keys Table
```sql
- id: BIGSERIAL PRIMARY KEY
- user_id: BIGINT (FK to users)
- key_hash: VARCHAR(255)
- key_prefix: VARCHAR(20)
- name: VARCHAR(100)
- scopes: TEXT[]
- is_active: BOOLEAN
- last_used_at: TIMESTAMP
- expires_at: TIMESTAMP
- created_at: TIMESTAMP
- updated_at: TIMESTAMP
- revoked_at: TIMESTAMP
```

### sessions Table
```sql
- id: BIGSERIAL PRIMARY KEY
- user_id: BIGINT (FK to users)
- session_token: VARCHAR(255)
- refresh_token: VARCHAR(255)
- ip_address: INET
- user_agent: TEXT
- is_active: BOOLEAN
- expires_at: TIMESTAMP
- created_at: TIMESTAMP
- updated_at: TIMESTAMP
- revoked_at: TIMESTAMP
```

## Configuration

Environment variables:
```bash
DATABASE_URL="postgres://yourdht:yourdhtpass@localhost:5432/dht_db?sslmode=disable"
USERMANAGER_PORT="8081"
JWT_SECRET="your-secret-key-change-in-production"
JWT_EXPIRATION="1h"
```

## Running
```bash
# Development
go run cmd/usermanager/*.go

# Production build
go build -o bin/usermanager cmd/usermanager/*.go
./bin/usermanager
```

## Endpoints

### POST /signup

Create a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "SecurePass123!"
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "email": "user@example.com",
  "username": "johndoe",
  "is_active": true,
  "created_at": "2024-11-15T10:00:00Z"
}
```

**Errors:**
- `400`: Validation error (missing fields, invalid email, weak password)
- `409`: Email or username already exists

---

### POST /login

Authenticate user and receive JWT tokens.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "johndoe",
    "is_active": true,
    "created_at": "2024-11-15T10:00:00Z"
  }
}
```

**Errors:**
- `401`: Invalid credentials
- `401`: Account inactive

---

### POST /apikeys

Create a new API key (requires authentication).

**Headers:**
- `Authorization: Bearer <access_token>`

**Request:**
```json
{
  "name": "Production API Key",
  "scopes": ["read", "write"],
  "expires_in_days": 90
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "name": "Production API Key",
  "key_prefix": "ydht_abc",
  "key": "ydht_abcdefgh1234567890abcdefgh1234567890",
  "scopes": ["read", "write"],
  "is_active": true,
  "expires_at": "2025-02-15T10:00:00Z",
  "created_at": "2024-11-15T10:00:00Z"
}
```

**⚠️ Important:** The complete API key is only returned once during creation. Store it securely!

**Errors:**
- `401`: Missing or invalid JWT token
- `400`: Invalid request (missing name)

---

### GET /apikeys

List all API keys for the authenticated user.

**Headers:**
- `Authorization: Bearer <access_token>`

**Response:** `200 OK`
```json
{
  "api_keys": [
    {
      "id": 1,
      "name": "Production API Key",
      "key_prefix": "ydht_abc",
      "scopes": ["read", "write"],
      "is_active": true,
      "last_used_at": "2024-11-15T12:30:00Z",
      "expires_at": "2025-02-15T10:00:00Z",
      "created_at": "2024-11-15T10:00:00Z"
    },
    {
      "id": 2,
      "name": "Development Key",
      "key_prefix": "ydht_xyz",
      "scopes": ["read"],
      "is_active": true,
      "last_used_at": null,
      "expires_at": null,
      "created_at": "2024-11-14T15:00:00Z"
    }
  ],
  "count": 2
}
```

---

### POST /validate-key

Validate an API key (internal use by Gateway).

**Request:**
```json
{
  "api_key": "ydht_abcdefgh1234567890..."
}
```

**Response:** `200 OK`
```json
{
  "valid": true,
  "user_id": 1
}
```

**Errors:**
- `401`: Invalid or expired API key
- `400`: Missing API key in request

---

### GET /health

Health check endpoint.

**Response:** `200 OK`
```json
{
  "status": "healthy",
  "service": "usermanager"
}
```

## Security

### Password Hashing

Passwords are hashed using **bcrypt** with default cost (10 rounds).
```go
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

### JWT Tokens

**Access Token:**
- Algorithm: HS256
- Expiration: 1 hour (configurable)
- Claims: user_id, email
- Issuer: "dht"

**Refresh Token:**
- Algorithm: HS256
- Expiration: 7 days
- Claims: user_id
- Used to obtain new access tokens

### API Keys

**Generation:**
- 32 random bytes → base64 URL-encoded
- Prefix: `ydht_` for identification
- Stored as bcrypt hash in database
- First 8 characters stored as prefix for quick lookup

**Format:**
```
ydht_<44-character-base64-string>
```

**Validation:**
1. Extract prefix from key
2. Find candidates with matching prefix
3. Compare bcrypt hash
4. Check expiration and active status

## Authentication Flow
![Authentication Flow](./images/usermanager-auth-flow.png)

## Error Handling

| Error Code | Description |
|------------|-------------|
| 400 | Bad Request - Invalid input data |
| 401 | Unauthorized - Invalid credentials or token |
| 409 | Conflict - Email/username already exists |
| 500 | Internal Server Error - Database or system error |

## Testing
```bash
# Create user
curl -X POST http://localhost:8081/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"testuser","password":"password123"}'

# Login
TOKEN=$(curl -s -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.access_token')

# Create API key
curl -X POST http://localhost:8081/apikeys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Test Key","scopes":["read","write"]}'

# List API keys
curl -X GET http://localhost:8081/apikeys \
  -H "Authorization: Bearer $TOKEN"
```

## Monitoring

Key metrics:
- Active users
- Failed login attempts
- API keys created/revoked
- Session count
- Database connection pool stats

## Troubleshooting

### Database Connection Failed

**Symptom:**
```
Unable to create connection pool
```

**Solution:**
- Verify PostgreSQL is running
- Check DATABASE_URL is correct
- Ensure database `dht_db` exists
- Run migrations

### Password Hashing Errors

**Symptom:**
```
failed to hash password
```

**Solution:**
- Check bcrypt implementation
- Verify Go crypto libraries are installed
- Check memory availability

### JWT Token Issues

**Symptom:**
```
invalid token
```

**Solution:**
- Verify JWT_SECRET matches across services
- Check token expiration
- Ensure clock synchronization across servers
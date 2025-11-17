# dht - Distributed Key-Value Store SaaS

A production-ready distributed hash table (DHT) based key-value storage service built with Go, featuring multi-tenancy, replication, and configurable consistency levels.

[Medium](https://medium.com/@krizzsrivastava/building-a-distributed-key-value-store-in-a-weekend-dht-958bf00525da)

## Interface

- **Landing Page**

![Landing Page](./images/landing-page.gif)

- **Signup Page**

![Signup Page](./images/signup.gif)

- **Signin Page**

![Signin Page](./images/signin.gif)

## Architecture Overview
![Architecture](./images/architecture.png)


## Features

### Core Features
- **Distributed Storage**: Data distributed across multiple nodes using consistent hashing
- **Multi-Tenancy**: User authentication and API key management
- **Configurable Consistency**: Support for both strong and eventual consistency
- **Replication**: Automatic data replication across nodes (3 replicas by default)
- **TTL Support**: Optional time-to-live for key-value pairs
- **Write-Ahead Logging**: Durability through WAL with automatic recovery
- **Rate Limiting**: Token bucket algorithm (100 requests/minute with burst of 10)

### Consistency Levels
![Consistency](./images/consistency.png)

### Operational Features
- **Metrics & Monitoring**: Key count, WAL size, replication lag, ACK times
- **Automatic Retries**: Failed replications retry up to 3 times with backoff
- **Health Checks**: All services expose `/health` endpoints
- **Graceful Shutdown**: Clean shutdown with connection draining

## Quick Start

### Prerequisites

- **Go 1.21+**
- **PostgreSQL 14+** 

### 1. Database Setup
```bash
# Start PostgreSQL in Podman
podman run -d \
  --name yourdht-postgres \
  -e POSTGRES_USER=yourdht \
  -e POSTGRES_PASSWORD=yourdhtpass \
  -e POSTGRES_DB=dht_db \
  -p 5432:5432 \
  postgres:14
```

### 2. Run Migrations
```bash
cd migrations

# Apply all migrations
psql $DATABASE_URL -f 001_create_users_table.up.sql
psql $DATABASE_URL -f 002_create_api_keys_table.up.sql
psql $DATABASE_URL -f 003_create_sessions_table.up.sql
psql $DATABASE_URL -f 004_create_usage_records_table.up.sql

cd ..
```

### 3. Install Dependencies
```bash
go mod tidy
```

### 4. Start All Services

Open 6 terminal windows and run:
```bash
# Terminal 1: User Manager
go run ./cmd/usermanager

# Terminal 2: DHT Node 1
DHTNODE_PORT=8082 NODE_ID=node-1 go run ./cmd/dhtnode

# Terminal 3: DHT Node 2
DHTNODE_PORT=8083 NODE_ID=node-2 go run ./cmd/dhtnode

# Terminal 4: DHT Node 3
DHTNODE_PORT=8084 NODE_ID=node-3 go run ./cmd/dhtnode

# Terminal 5: Replicator
go run ./cmd/replicator

# Terminal 6: Gateway
go run ./cmd/gateway
```

### 5. Create User & Get API Key
```bash
# Create user account
curl -X POST http://localhost:8081/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "testuser",
    "password": "securepass123"
  }'

# Login to get JWT token
TOKEN=$(curl -s -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepass123"
  }' | jq -r '.access_token')

# Create API key
API_KEY=$(curl -s -X POST http://localhost:8081/apikeys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "My Application Key",
    "scopes": ["read", "write"],
    "expires_in_days": 90
  }' | jq -r '.key')

echo "Your API Key: $API_KEY"
```

### 6. Use the KV Store
```bash
# Store a value (eventual consistency - default)
curl -X PUT "http://localhost:8080/v1/kv/user:123" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}'

# Store with strong consistency
curl -X PUT "http://localhost:8080/v1/kv/config:app" \
  -H "X-API-Key: $API_KEY" \
  -H "X-Consistency: strong" \
  -H "Content-Type: application/json" \
  -d '{"theme":"dark","language":"en"}'

# Store with TTL (expires in 1 hour)
curl -X PUT "http://localhost:8080/v1/kv/session:abc123?ttl=1h" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"user_id":123,"expires":"2024-12-31"}'

# Retrieve a value
curl -X GET "http://localhost:8080/v1/kv/user:123" \
  -H "X-API-Key: $API_KEY"

# Delete a value
curl -X DELETE "http://localhost:8080/v1/kv/user:123" \
  -H "X-API-Key: $API_KEY"
```

## API Reference

### User Management Service (Port 8081)

#### POST /signup
Create a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "username": "username",
  "password": "password123"
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "email": "user@example.com",
  "username": "username",
  "is_active": true,
  "created_at": "2024-11-15T10:00:00Z"
}
```

---

#### POST /login
Authenticate and get JWT tokens.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "username",
    "is_active": true,
    "created_at": "2024-11-15T10:00:00Z"
  }
}
```

---

#### POST /apikeys
Create a new API key (requires JWT token).

**Headers:**
- `Authorization: Bearer <token>`

**Request:**
```json
{
  "name": "Production Key",
  "scopes": ["read", "write"],
  "expires_in_days": 90
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "name": "Production Key",
  "key_prefix": "ydht_abc",
  "key": "ydht_abcdefgh1234567890...",
  "scopes": ["read", "write"],
  "is_active": true,
  "expires_at": "2025-02-15T10:00:00Z",
  "created_at": "2024-11-15T10:00:00Z"
}
```

**Note:** The full API key is only shown once during creation. Store it securely!

---

#### GET /apikeys
List all API keys for authenticated user.

**Headers:**
- `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "api_keys": [
    {
      "id": 1,
      "name": "Production Key",
      "key_prefix": "ydht_abc",
      "scopes": ["read", "write"],
      "is_active": true,
      "last_used_at": "2024-11-15T12:00:00Z",
      "expires_at": "2025-02-15T10:00:00Z",
      "created_at": "2024-11-15T10:00:00Z"
    }
  ],
  "count": 1
}
```

### Gateway Service (Port 8080)

All gateway endpoints require the `X-API-Key` header.

#### PUT /v1/kv/{key}
Store a key-value pair.

**Headers:**
- `X-API-Key: <your-api-key>` (required)
- `X-Consistency: eventual|strong` (optional, default: eventual)
- `Content-Type: application/json` (or any type)

**Query Parameters:**
- `ttl` (optional): Time-to-live duration (e.g., "1h", "30m", "24h")

**Request Body:** Any binary data or JSON

**Response:** `200 OK`
```json
{
  "success": true,
  "key": "user:123",
  "primary_node": "http://localhost:8082",
  "replicas": 2
}
```

**Examples:**
```bash
# Store JSON
curl -X PUT "http://localhost:8080/v1/kv/user:123" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name":"John"}'

# Store with TTL
curl -X PUT "http://localhost:8080/v1/kv/session:abc?ttl=1h" \
  -H "X-API-Key: $API_KEY" \
  -d 'session-data'

# Store with strong consistency
curl -X PUT "http://localhost:8080/v1/kv/config:critical" \
  -H "X-API-Key: $API_KEY" \
  -H "X-Consistency: strong" \
  -d '{"setting":"value"}'
```

---

#### GET /v1/kv/{key}
Retrieve a value by key.

**Headers:**
- `X-API-Key: <your-api-key>` (required)
- `X-Consistency: eventual|strong` (optional, default: eventual)

**Response:** `200 OK`
Returns the raw value with `Content-Type: application/octet-stream`

**Example:**
```bash
curl -X GET "http://localhost:8080/v1/kv/user:123" \
  -H "X-API-Key: $API_KEY"
```

**Error Response:** `404 Not Found`
```json
{
  "error": "Key not found"
}
```

---

#### DELETE /v1/kv/{key}
Delete a key-value pair.

**Headers:**
- `X-API-Key: <your-api-key>` (required)
- `X-Consistency: eventual|strong` (optional, default: eventual)

**Response:** `200 OK`
```json
{
  "success": true,
  "key": "user:123",
  "primary_node": "http://localhost:8082",
  "replicas": 2
}
```

**Example:**
```bash
curl -X DELETE "http://localhost:8080/v1/kv/user:123" \
  -H "X-API-Key: $API_KEY"
```

### DHT Node Service (Ports 8082-8084)

#### GET /metrics
Get node metrics.

**Response:** `200 OK`
```json
{
  "node_id": "node-1",
  "key_count": 1247,
  "wal_size": 524288,
  "timestamp": 1700050000
}
```

#### GET /health
Health check endpoint.

**Response:** `200 OK`
```json
{
  "status": "healthy",
  "node_id": "node-1"
}
```

### Replicator Service (Port 8085)

#### GET /metrics
Get replication metrics.

**Response:** `200 OK`
```json
{
  "total_replications": 5420,
  "successful_replicas": 10782,
  "failed_replicas": 58,
  "queue_size": 23,
  "average_ack_time_ms": 45.7,
  "max_replication_lag_ms": 1234.5,
  "retries_in_progress": 2
}
```

#### GET /health
Health check endpoint.

**Response:** `200 OK`
```json
{
  "status": "healthy",
  "service": "replicator",
  "queue_size": 23
}
```

## 🔧 Configuration

All services can be configured via environment variables:

### Common Variables
```bash
DATABASE_URL="postgres://user:pass@localhost:5432/dht_db?sslmode=disable"
JWT_SECRET="your-secret-key-change-in-production"
JWT_EXPIRATION="1h"
```

### Service-Specific Ports
```bash
USERMANAGER_PORT="8081"
GATEWAY_PORT="8080"
DHTNODE_PORT="8082"  # or 8083, 8084 for other nodes
REPLICATOR_PORT="8085"
NODE_ID="node-1"     # for DHT nodes
```

## Architecture Deep Dive

### Consistent Hashing
![Consistent Hashing](./images/consistent_hashing.png)

- Uses **FNV-1a hash** for fast, uniform distribution
- **150 virtual nodes** per physical node for balanced load
- Automatically handles node addition/removal
- Keys are mapped to primary node + 2 replicas (configurable)

### Replication Strategy

**Eventual Consistency:**
1. Client writes to primary node
2. Primary acknowledges immediately
3. Replication happens asynchronously in background
4. Failed replications retry up to 3 times with exponential backoff
5. **Latency:** ~100ms (primary only)

**Strong Consistency:**
1. Client writes to primary node
2. Replicator writes to all replica nodes in parallel
3. Wait for majority (2 out of 3) to acknowledge
4. Return success only after majority ACK
5. **Latency:** ~300-500ms (includes replication)

### Write-Ahead Logging (WAL)
![WAL](./images/wal.png)

- Every write is logged to disk before applying to storage
- Uses `encoding/gob` for efficient serialization
- `fsync()` called after each write for durability
- Automatic recovery on node restart
- Supports log compaction/truncation

### Rate Limiting
- **Token Bucket Algorithm**
- 100 requests per minute per user
- Burst capacity of 10 requests
- Independent bucket per user ID
- Refills at 1.67 tokens/second

## Monitoring & Metrics

### Key Metrics to Monitor

**Gateway:**
- Request rate per user
- Request latency (p50, p95, p99)
- Rate limit hits

**DHT Nodes:**
- Key count
- WAL file size
- Disk usage

**Replicator:**
- Replication queue size
- Average ACK time
- Replication lag
- Failed replication count

### Example Monitoring Script
```bash
#!/bin/bash
# monitor.sh - Simple monitoring script

while true; do
  echo "DHT Nodes"
  curl -s http://localhost:8082/metrics | jq
  curl -s http://localhost:8083/metrics | jq
  curl -s http://localhost:8084/metrics | jq
  
  echo "Replicator"
  curl -s http://localhost:8085/metrics | jq
  
  sleep 10
done
```

## Testing

### Manual Testing
```bash
# Test eventual consistency
for i in {1..100}; do
  curl -X PUT "http://localhost:8080/v1/kv/test:$i" \
    -H "X-API-Key: $API_KEY" \
    -H "X-Consistency: eventual" \
    -d "value-$i"
done

# Test strong consistency
for i in {1..100}; do
  curl -X PUT "http://localhost:8080/v1/kv/critical:$i" \
    -H "X-API-Key: $API_KEY" \
    -H "X-Consistency: strong" \
    -d "value-$i"
done

# Verify replication
curl http://localhost:8082/store/test:1
curl http://localhost:8083/store/test:1
curl http://localhost:8084/store/test:1
```

## Troubleshooting

### Service Won't Start

**Problem:** Port already in use
```
bind: address already in use
```

**Solution:**
```bash
# Find process using the port
lsof -i :8080

# Kill the process
kill -9 <PID>
```

---

**Problem:** Database connection failed
```
Unable to create connection pool
```

**Solution:**
```bash
# Verify PostgreSQL is running
podman ps | grep postgres

# Test connection
psql $DATABASE_URL -c "SELECT 1"
```

### Replication Not Working

**Problem:** Replication fails with connection refused

**Solution:**
- Ensure replicator is running on port 8085 (not a DHT node)
- Check DHT nodes are running on ports 8082-8084
- Verify network connectivity between services

### Data Not Persisting

**Problem:** Data lost after node restart

**Solution:**
- Check WAL file exists in `data/` directory
- Ensure write permissions on `data/` directory
- Check logs for WAL recovery messages on startup

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Contact

For questions or support, please open an issue on GitHub.

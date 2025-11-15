# Gateway Service

The Gateway service is the entry point for all client requests to the dht distributed key-value store.

## Overview

The Gateway handles:
- **API Key Authentication**: Validates requests against the User Manager service
- **Rate Limiting**: Token bucket algorithm (100 req/min per user)
- **Request Routing**: Uses consistent hashing to route requests to appropriate DHT nodes
- **Replication Coordination**: Triggers data replication based on consistency level
- **Consistency Management**: Supports both strong and eventual consistency

## Architecture
![Gateway Architecture](./images/gateway-architecture.png)

## Configuration

Environment variables:
```bash
GATEWAY_PORT="8080"              # Gateway listen port
USERMANAGER_PORT="8081"          # User Manager port for auth
REPLICATOR_PORT="8085"           # Replicator port
```

## Running
```bash
go run cmd/gateway/*.go
```

## Request Flow

### Write Operation (PUT)
![Gateway Put Flow](./images/gateway-put-flow.png)

### Read Operation (GET)

1. Validate API key → Get user ID
2. Check rate limit
3. Use hash ring to find primary node
4. Forward request to primary node
5. Return response to client

## Endpoints

### PUT /v1/kv/{key}

Store a key-value pair.

**Headers:**
- `X-API-Key`: API key (required)
- `X-Consistency`: `eventual` or `strong` (optional, default: eventual)
- `Content-Type`: Any (e.g., `application/json`)

**Query Parameters:**
- `ttl`: Time-to-live (e.g., `1h`, `30m`, `24h`)

**Example:**
```bash
curl -X PUT "http://localhost:8080/v1/kv/user:123?ttl=1h" \
  -H "X-API-Key: ydht_abc123..." \
  -H "X-Consistency: strong" \
  -H "Content-Type: application/json" \
  -d '{"name":"John","age":30}'
```

### GET /v1/kv/{key}

Retrieve a value by key.

**Headers:**
- `X-API-Key`: API key (required)
- `X-Consistency`: `eventual` or `strong` (optional)

**Example:**
```bash
curl -X GET "http://localhost:8080/v1/kv/user:123" \
  -H "X-API-Key: ydht_abc123..."
```

### DELETE /v1/kv/{key}

Delete a key-value pair.

**Headers:**
- `X-API-Key`: API key (required)
- `X-Consistency`: `eventual` or `strong` (optional)

**Example:**
```bash
curl -X DELETE "http://localhost:8080/v1/kv/user:123" \
  -H "X-API-Key: ydht_abc123..."
```

### GET /health

Health check endpoint.

**Example:**
```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "gateway",
  "nodes": [
    "http://localhost:8082",
    "http://localhost:8083",
    "http://localhost:8084"
  ]
}
```

## Rate Limiting

### Token Bucket Algorithm
![Token Bucket](./images/token-bucket.png)

**Configuration:**
- Max tokens: 10 (burst capacity)
- Refill rate: 100 tokens/minute = 1.67 tokens/second
- Per-user buckets

## Consistent Hashing

The Gateway uses a hash ring to determine which DHT node should store each key.

**Key Features:**
- 150 virtual nodes per physical node
- FNV-1a hash function
- Returns primary + 2 replica nodes
- Automatic rebalancing when nodes join/leave

**Example:**
```
Key: "user:123" 
→ Hash: 0x7a3f9e12
→ Primary: http://localhost:8082
→ Replicas: [http://localhost:8083, http://localhost:8084]
```

## Error Handling

| Error | Status Code | Description |
|-------|-------------|-------------|
| Missing API key | 401 | X-API-Key header required |
| Invalid API key | 401 | Key not found or expired |
| Rate limit exceeded | 429 | Too many requests |
| Invalid consistency | 400 | Must be "strong" or "eventual" |
| No nodes available | 503 | All DHT nodes down |
| Primary node unavailable | 503 | Cannot reach primary node |

## Monitoring

Key metrics to monitor:
- Request rate (per user, per endpoint)
- Request latency (p50, p95, p99)
- Rate limit hits
- Node health status
- Error rates

## Testing
```bash
# Get API key
API_KEY="ydht_your_key_here"

# Test eventual consistency
time curl -X PUT "http://localhost:8080/v1/kv/test:1" \
  -H "X-API-Key: $API_KEY" \
  -H "X-Consistency: eventual" \
  -d "test-value"

# Test strong consistency
time curl -X PUT "http://localhost:8080/v1/kv/test:2" \
  -H "X-API-Key: $API_KEY" \
  -H "X-Consistency: strong" \
  -d "test-value"

# Compare latencies
```

## Troubleshooting

### Connection Refused to User Manager

**Symptom:** API key validation fails
```
Failed to trigger replication: dial tcp [::1]:8081: connect: connection refused
```

**Solution:** Ensure User Manager is running on port 8081

### Connection Refused to Replicator

**Symptom:** Replication fails
```
Failed to trigger replication: dial tcp [::1]:8085: connect: connection refused
```

**Solution:** Ensure Replicator is running on port 8085 (not a DHT node)

### Rate Limit Issues

**Symptom:** Getting 429 errors

**Solution:**
- Check rate limiter configuration
- Verify user ID is being correctly extracted
- Monitor rate limiter metrics
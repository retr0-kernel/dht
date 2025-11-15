# DHT Node Service

The DHT Node service provides distributed key-value storage with write-ahead logging for durability.

## Overview

Each DHT Node:
- **Stores Data**: In-memory key-value storage with TTL support
- **WAL (Write-Ahead Log)**: Logs all writes for crash recovery
- **Metrics**: Exposes key count and WAL size
- **HTTP API**: Simple REST interface for CRUD operations

## Architecture
![DHTNode Architecture](./images/dhtnode-architecture.png)

## Storage Engine

### In-Memory Store
![In Memory Store](./images/inmemory-store.png)

**Features:**
- Thread-safe with RWMutex
- Optional TTL per key
- Automatic cleanup of expired entries (every 60 seconds)
- Soft delete support

### Write-Ahead Log (WAL)

**Purpose:** Ensure durability - data survives crashes and restarts

**Operations Logged:**
- SET: Store/update a key
- DELETE: Remove a key

**File Format:**
- Encoding: Go's `encoding/gob`
- Location: `data/<node-id>-wal.log`
- Append-only

**WAL Entry Structure:**
```go
type WALEntry struct {
    Operation string        // "SET" or "DELETE"
    Key       string
    Value     []byte
    TTL       time.Duration
    Timestamp time.Time
}
```

## Configuration

Environment variables:
```bash
DHTNODE_PORT="8082"    # HTTP server port
NODE_ID="node-1"       # Unique node identifier
```

## Running
```bash
# Node 1
DHTNODE_PORT=8082 NODE_ID=node-1 go run ./cmd/dhtnode

# Node 2
DHTNODE_PORT=8083 NODE_ID=node-2 go run ./cmd/dhtnode

# Node 3
DHTNODE_PORT=8084 NODE_ID=node-3 go run ./cmd/dhtnode
```

## Endpoints

### PUT /store/{key}

Store or update a key-value pair.

**Query Parameters:**
- `ttl` (optional): Time-to-live duration (e.g., "1h", "30m")

**Request Body:** Raw bytes (any content type)

**Example:**
```bash
# Store JSON with 1 hour TTL
curl -X PUT "http://localhost:8082/store/user:123?ttl=1h" \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}'

# Store plain text
curl -X PUT "http://localhost:8082/store/message:abc" \
  -d "Hello World"
```

**Response:** `200 OK`
```json
{
  "success": true,
  "key": "user:123",
  "node": "node-1"
}
```

**Process:**
1. Write operation to WAL
2. Sync WAL to disk (`fsync`)
3. Update in-memory store
4. Return success

---

### GET /store/{key}

Retrieve a value by key.

**Example:**
```bash
curl -X GET "http://localhost:8082/store/user:123"
```

**Response:** `200 OK`
- Returns raw value
- Header: `X-Node-ID: node-1`
- Header: `Content-Type: application/octet-stream`

**Error:** `404 Not Found`
```json
{
  "error": "Key not found"
}
```

---

### DELETE /store/{key}

Delete a key-value pair.

**Example:**
```bash
curl -X DELETE "http://localhost:8082/store/user:123"
```

**Response:** `200 OK`
```json
{
  "success": true,
  "key": "user:123",
  "node": "node-1"
}
```

**Error:** `404 Not Found`
```json
{
  "error": "Key not found"
}
```

**Process:**
1. Write DELETE to WAL
2. Sync WAL to disk
3. Remove from in-memory store
4. Return success

---

### GET /metrics

Get node metrics.

**Example:**
```bash
curl http://localhost:8082/metrics
```

**Response:** `200 OK`
```json
{
  "node_id": "node-1",
  "key_count": 1247,
  "wal_size": 524288,
  "timestamp": 1700050000
}
```

**Metrics:**
- `key_count`: Number of keys in storage (excluding expired)
- `wal_size`: WAL file size in bytes
- `timestamp`: Current Unix timestamp

---

### GET /health

Health check endpoint.

**Example:**
```bash
curl http://localhost:8082/health
```

**Response:** `200 OK`
```json
{
  "status": "healthy",
  "node_id": "node-1"
}
```

## Write-Ahead Log Details

### Startup Recovery
![WAL Recovery](./images/wal-recovery.png)

**Features:**
- Skips expired entries during recovery
- Handles corrupted entries gracefully
- Reports number of entries restored

**Example Log Output:**
```
WAL: Restored 1247 entries from data/node-1-wal.log
DHT Node node-1 starting on port 8082
```

### WAL Compaction

**When to compact:**
- WAL file exceeds size threshold (e.g., 100MB)
- Periodic maintenance (e.g., daily)

**Process:**
1. Create snapshot of current in-memory state
2. Write snapshot to new file
3. Truncate/delete old WAL
4. Start fresh WAL

*Note: Compaction not yet implemented in current version*

## TTL (Time-To-Live) Support

### How It Works
```go
// Store with TTL
storage.Set("session:abc", data, 1*time.Hour)

// Entry structure
Entry {
    Key: "session:abc"
    Value: []byte{...}
    ExpiresAt: time.Now().Add(1 * time.Hour)
    CreatedAt: time.Now()
    UpdatedAt: time.Now()
}
```

### Cleanup Process

**Background goroutine** runs every 60 seconds:
1. Lock storage
2. Iterate over all entries
3. Check if `ExpiresAt < time.Now()`
4. Delete expired entries
5. Unlock storage

### Behavior

- Keys with no TTL never expire
- Expired keys return "Key not found" error
- Expired keys not counted in metrics
- WAL entries for expired keys skipped during recovery

## Data Persistence

### What Survives Crashes

✅ **Persisted** (in WAL):
- All SET operations
- All DELETE operations
- TTL values
- Timestamps

❌ **Lost** (in-memory only):
- Last access times
- Runtime metrics
- Active connections

### Recovery Guarantees

- **Durability**: All acknowledged writes survive crashes
- **Consistency**: WAL replay maintains operation order
- **Atomicity**: Each WAL entry is atomic (gob encoding)

## Performance Considerations

### Read Performance
- **In-memory**: ~1-5 microseconds
- **No disk I/O** for reads

### Write Performance
- **WAL write**: ~100-500 microseconds (depends on disk)
- **Memory update**: ~1-5 microseconds
- **Total**: ~100-500 microseconds

**Bottleneck:** WAL fsync (waiting for disk)


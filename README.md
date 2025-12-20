# StrangeDB

<div align="center">
  <pre>
███████╗████████╗██████╗  █████╗ ███╗   ██╗ ██████╗ ███████╗
██╔════╝╚══██╔══╝██╔══██╗██╔══██╗████╗  ██║██╔════╝ ██╔════╝
███████╗   ██║   ██████╔╝███████║██╔██╗ ██║██║  ███╗█████╗  
╚════██║   ██║   ██╔══██╗██╔══██║██║╚██╗██║██║   ██║██╔══╝  
███████║   ██║   ██║  ██║██║  ██║██║ ╚████║╚██████╔╝███████╗
╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝
                                                         DB
  </pre>
  
  <strong>A distributed key-value store with eventual consistency</strong>
  
  <br/>
  
  <em>Phase 1: Foundation</em>
</div>

---

## Task Breakdown

### 1.1 Project Setup
- [x] Initialize Go module (`github.com/yourusername/strangedb`)
- [x] Create directory structure as per design doc
- [x] Setup `.gitignore`, `Makefile`, `LICENSE`
- [x] Configure linting (golangci-lint)
- [x] Create initial `go.mod` with dependencies

### 1.2 Configuration Management
- [x] Create `internal/config/config.go`
- [x] Environment variable parsing
- [x] Command-line flag parsing
- [x] Default values
- [x] Validation logic

### 1.3 Logging Infrastructure
- [x] Setup zerolog for structured logging
- [x] Log levels (debug, info, warn, error)
- [x] Request ID propagation
- [x] Pretty printing for development

### 1.4 Hybrid Logical Clock (HLC)
- [x] Create `internal/hlc/clock.go`
- [x] Implement `Now()` - generate new timestamp
- [x] Implement `Update()` - update on receive
- [x] Implement `Compare()` - compare two HLC timestamps
- [x] Thread-safe implementation
- [x] Unit tests

### 1.5 Storage Layer
- [x] Create `internal/storage/storage.go` - interface
- [x] Create `internal/storage/badger.go` - BadgerDB impl
- [x] Key encoding (`d:`, `m:`, `t:` prefixes)
- [x] Value encoding (protobuf or JSON)
- [x] Basic operations: Get, Set, Delete
- [x] Tombstone support
- [x] Unit tests

### 1.6 HTTP API Server
- [ ] Create `internal/transport/http/server.go`
- [ ] Create `internal/transport/http/handlers.go`
- [ ] Implement endpoints:
  - `POST /api/v1/kv` - Set key
  - `GET /api/v1/kv/{key}` - Get key
  - `DELETE /api/v1/kv/{key}` - Delete key
  - `GET /health` - Health check
  - `GET /api/v1/status` - Node status
- [ ] Request/response logging middleware
- [ ] Error handling middleware
- [ ] Unit and integration tests

### 1.7 Node Lifecycle
- [ ] Create `internal/node/node.go`
- [ ] State machine (New → Starting → Ready → Stopping → Stopped)
- [ ] Graceful startup
- [ ] Graceful shutdown

### 1.8 Main Entry Point
- [ ] Create `cmd/strangedb/main.go`
- [ ] Wire all components together
- [ ] Signal handling (SIGINT, SIGTERM)

### 1.9 Docker Support
- [ ] Create `Dockerfile` (multi-stage build)
- [ ] Create `.dockerignore`
- [ ] Build script in Makefile

### 1.10 Testing & Documentation
- [ ] Unit tests for all packages
- [ ] Integration test for HTTP API
- [ ] Basic README with quickstart

---

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker (optional)

### Build & Run

```bash
# Build
make build

# Run single node
./build/strangedb --http-port 9000 --data-dir ./data

# Or with Docker
docker build -t strangedb .
docker run -p 9000:9000 strangedb
```

### Test the API

```bash
# Set a key
curl -X POST http://localhost:9000/api/v1/kv \
  -H "Content-Type: application/json" \
  -d '{"key": "hello", "value": "d29ybGQ="}'

# Get a key  
curl http://localhost:9000/api/v1/kv/hello

# Delete a key
curl -X DELETE http://localhost:9000/api/v1/kv/hello

# Health check
curl http://localhost:9000/health
```

---

## 📋 Current Phase: Foundation

This version implements a single-node key-value store with:
- ✅ BadgerDB storage engine
- ✅ HTTP REST API
- ✅ Hybrid Logical Clock (HLC) timestamps
- ✅ Graceful shutdown
- ✅ Docker support

---

## 🗺️ Roadmap

| Phase | Status | Description |
|-------|--------|-------------|
| **Phase 1** | 🔨 In Progress | Single-node KV store |
| Phase 2 | ⏳ Planned | Distributed cluster |
| Phase 3 | ⏳ Planned | Consistency & reliability |
| Phase 4 | ⏳ Planned | CLI & observability |
| Phase 5 | ⏳ Planned | Web dashboard |
| Phase 6 | ⏳ Planned | AI query patterns |

---

## 🔧 Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `STRANGE_HTTP_PORT` | `9000` | HTTP API port |
| `STRANGE_DATA_DIR` | `./data` | Data directory |
| `STRANGE_LOG_LEVEL` | `info` | Log level |

---

## 📖 API Reference

### Set Key
```http
POST /api/v1/kv
Content-Type: application/json

{"key": "mykey", "value": "base64_encoded_value"}
```

### Get Key
```http
GET /api/v1/kv/{key}
```

### Delete Key
```http
DELETE /api/v1/kv/{key}
```

### Health Check
```http
GET /health
```

---

## 📄 License

MIT License

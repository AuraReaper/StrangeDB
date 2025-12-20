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

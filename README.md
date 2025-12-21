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
  
  <strong>A highly-scalable distributed key-value store with peer-to-peer topology</strong>
  
  <br/>
  
  <em>Phase 2: Distributed Core</em>
</div>

---

## 🎯 Overview

**StrangeDB** is a distributed key-value store featuring:
- Peer-to-peer architecture (no single point of failure)
- Consistent hashing with virtual nodes
- Configurable replication (N=3, R=2, W=2)
- Eventual consistency with Last-Write-Wins

---

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose (optional)

### Start a 3-Node Cluster

```bash
# Build
make build

# Start cluster
./scripts/run_cluster.sh

# Or with Docker Compose
docker-compose up -d
```

### Test the Cluster

```bash
# Set a key (connect to ANY node)
curl -X POST http://localhost:9000/api/v1/kv \
  -H "Content-Type: application/json" \
  -d '{"key": "user:123", "value": "eyJuYW1lIjoiSm9obiJ9"}'

# Get a key (from any node - it routes automatically)
curl http://localhost:9001/api/v1/kv/user:123

# Check cluster status
curl http://localhost:9000/api/v1/cluster/status
```

---

## 🏗️ Architecture

```
         Client App
              |
              | Can connect to ANY node
              |
    +---------+---------+---------+
    |         |         |         |
    v         v         v         v
+-------+  +-------+  +-------+
| Node 1|  | Node 2|  | Node 3|  
|:9000  |  |:9001  |  |:9002  |  
|       |  |       |  |       |  
| Routes|<-| Routes|<-| Routes|  Each node can:
| to    |->| to    |->| to    |  - Handle requests locally
| peers |  | peers |  | peers |  - Route to correct peer
+-------+  +-------+  +-------+  - No single point of failure
```

---

## 📋 Current Phase: Distributed Core

This version implements:
- ✅ Consistent hashing with virtual nodes
- ✅ gRPC inter-node communication
- ✅ Gossip protocol for membership
- ✅ Automatic request routing
- ✅ Replication (N=3 default)
- ✅ Quorum reads/writes (R=2, W=2)

---

## 🗺️ Roadmap

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1 | ✅ Complete | Single-node KV store |
| **Phase 2** | 🔨 In Progress | Distributed cluster |
| Phase 3 | ⏳ Planned | Consistency & reliability |
| Phase 4 | ⏳ Planned | CLI & observability |
| Phase 5 | ⏳ Planned | Web dashboard |
| Phase 6 | ⏳ Planned | AI query patterns |

---

## 🔧 Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `STRANGE_NODE_ID` | auto | Unique node ID |
| `STRANGE_HTTP_PORT` | `9000` | HTTP API port |
| `STRANGE_GRPC_PORT` | `9001` | gRPC port |
| `STRANGE_DATA_DIR` | `./data` | Data directory |
| `STRANGE_SEEDS` | `` | Seed node URLs |
| `STRANGE_REPLICATION_N` | `3` | Replication factor |
| `STRANGE_READ_QUORUM` | `2` | Read quorum |
| `STRANGE_WRITE_QUORUM` | `2` | Write quorum |

---

## 📖 API Reference

### Key Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/kv` | Set key-value |
| `GET` | `/api/v1/kv/{key}` | Get value |
| `DELETE` | `/api/v1/kv/{key}` | Delete key |

### Cluster Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/cluster/status` | Cluster health |
| `GET` | `/api/v1/cluster/ring` | Hash ring info |
| `GET` | `/health` | Node health |

---

## 🐳 Docker Deployment

```bash
# Start 3-node cluster
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f

# Stop cluster
docker-compose down
```

---

## 📄 License

Apache 2.0

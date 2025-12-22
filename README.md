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
  
  <em>Phase 4: Observability & Operations</em>
</div>

---

## 🎯 Overview

**StrangeDB** is a distributed key-value store featuring:
- Peer-to-peer architecture (no single point of failure)
- Consistent hashing with virtual nodes
- Configurable replication & quorum
- Eventual consistency with automatic conflict resolution
- **Full observability with Prometheus & OpenTelemetry**
- **Beautiful Terminal UI (strange-cli)**

---

## 🚀 Quick Start

### Start Cluster

```bash
docker-compose up -d
```

### Use strange-cli

```bash
# Install CLI
go install ./cmd/strange-cli

# Connect to cluster
strange-cli --urls localhost:9000,localhost:9001,localhost:9002
```

### Test API

```bash
# Set a key
curl -X POST http://localhost:9000/api/v1/kv \
  -d '{"key": "hello", "value": "d29ybGQ="}'

# Get metrics
curl http://localhost:9000/metrics
```

---

## 📊 strange-cli

Beautiful terminal client for StrangeDB:

```
╔═══════════════════════════════════════════════════════════════════╗
║   ███████╗████████╗██████╗  █████╗ ███╗   ██╗ ██████╗ ███████╗    ║
║   ██╔════╝╚══██╔══╝██╔══██╗██╔══██╗████╗  ██║██╔════╝ ██╔════╝    ║
║   ███████╗   ██║   ██████╔╝███████║██╔██╗ ██║██║  ███╗█████╗      ║
║   ╚════██║   ██║   ██╔══██╗██╔══██║██║╚██╗██║██║   ██║██╔══╝      ║
║   ███████║   ██║   ██║  ██║██║  ██║██║ ╚████║╚██████╔╝███████╗    ║
║   ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝    ║
║                                                         DB v0.4.0  ║
╠═══════════════════════════════════════════════════════════════════╣
║  [Cluster]  [Keys]  [Metrics]  [Help]                    [q] Quit  ║
╠═══════════════════════════════════════════════════════════════════╣
║   ● node-1  localhost:9000  [HEALTHY]   Keys: 1,234                ║
║   ● node-2  localhost:9001  [HEALTHY]   Keys: 1,198                ║
║   ● node-3  localhost:9002  [HEALTHY]   Keys: 1,245                ║
╚═══════════════════════════════════════════════════════════════════╝
```

### CLI Features
- Real-time cluster health monitoring
- Interactive key-value operations
- Performance metrics dashboard
- Keyboard navigation

---

## 📈 Observability

### Prometheus Metrics
```bash
curl http://localhost:9000/metrics
```

Key metrics:
- `strangedb_requests_total` - Request count by operation
- `strangedb_request_duration_seconds` - Latency histogram
- `strangedb_keys_total` - Total keys stored
- `strangedb_gossip_messages_total` - Gossip activity

### OpenTelemetry Tracing
Configure OTLP exporter:
```bash
STRANGE_OTEL_ENDPOINT=localhost:4317 ./strangedb
```

---

## 📋 Current Phase: Observability & CLI

This version adds:
- ✅ Prometheus metrics endpoint
- ✅ OpenTelemetry tracing
- ✅ strange-cli terminal UI
- ✅ Structured JSON logging
- ✅ Health check endpoints

---

## 🗺️ Roadmap

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1 | ✅ Complete | Single-node KV store |
| Phase 2 | ✅ Complete | Distributed cluster |
| Phase 3 | ✅ Complete | Consistency & reliability |
| **Phase 4** | 🔨 In Progress | CLI & observability |
| Phase 5 | ⏳ Planned | Web dashboard |
| Phase 6 | ⏳ Planned | AI query patterns |

---

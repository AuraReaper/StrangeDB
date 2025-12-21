#!/bin/bash

set -e
mkdir -p logs

echo "Building StrangeDB..."
go build -o build/strangedb ./cmd/strangedb

mkdir -p data/node1 data/node2 data/node3

pkill -f "strangedb" 2>/dev/null || true
sleep 1

# FIXED: Use gRPC ports for seeds (inter-node communication)
NODE1_GRPC="localhost:9001"
NODE2_GRPC="localhost:9011"
NODE3_GRPC="localhost:9021"
ALL_SEEDS="${NODE1_GRPC},${NODE2_GRPC},${NODE3_GRPC}"

echo "Starting Node 1..."
./build/strangedb \
    --node-id node-1 \
    --http-port 9000 \
    --grpc-port 9001 \
    --data-dir ./data/node1 \
    --seeds "${ALL_SEEDS}" \
    > logs/node1.log 2>&1 &

echo "Starting Node 2..."
./build/strangedb \
    --node-id node-2 \
    --http-port 9010 \
    --grpc-port 9011 \
    --data-dir ./data/node2 \
    --seeds "${ALL_SEEDS}" \
    > logs/node2.log 2>&1 &

echo "Starting Node 3..."
./build/strangedb \
    --node-id node-3 \
    --http-port 9020 \
    --grpc-port 9021 \
    --data-dir ./data/node3 \
    --seeds "${ALL_SEEDS}" \
    > logs/node3.log 2>&1 &

sleep 3

echo ""
echo "Cluster started!"
echo "  Node 1: HTTP=localhost:9000, gRPC=localhost:9001"
echo "  Node 2: HTTP=localhost:9010, gRPC=localhost:9011"
echo "  Node 3: HTTP=localhost:9020, gRPC=localhost:9021"
echo ""
echo "Test with:"
echo "  curl -X POST http://localhost:9000/api/v1/kv -d '{\"key\":\"test\",\"value\":\"aGVsbG8=\"}'"
echo "  curl http://localhost:9010/api/v1/kv/test  # Should work from any node!"
echo ""
echo "Logs: tail -f logs/node1.log"
echo "Stop: pkill -f strangedb"

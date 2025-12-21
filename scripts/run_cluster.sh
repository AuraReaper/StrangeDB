#!/bin/bash

# Starts a 3-node local cluster

set -e

mkdir -p logs

# Build first
echo "Building StrangeDB..."
go build -o build/strangedb ./cmd/strangedb

# Create data directories
mkdir -p data/node1 data/node2 data/node3

# Kill any existing processes
pkill -f "strangedb" 2>/dev/null || true
sleep 1

# Node URLs
NODE1_URL="http://localhost:9000"
NODE2_URL="http://localhost:9010"
NODE3_URL="http://localhost:9020"
ALL_SEEDS="${NODE1_URL},${NODE2_URL},${NODE3_URL}"

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

# Wait for startup
sleep 3

echo ""
echo "Cluster started!"
echo "  Node 1: http://localhost:9000"
echo "  Node 2: http://localhost:9010"
echo "  Node 3: http://localhost:9020"
echo ""
echo "Test with:"
echo "  curl -X POST http://localhost:9000/api/v1/kv -d '{\"key\":\"test\",\"value\":\"aGVsbG8=\"}'"
echo "  curl http://localhost:9010/api/v1/kv/test"
echo ""
echo "Logs in: logs/"
echo "Stop with: pkill -f strangedb"

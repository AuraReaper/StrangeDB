# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o strangedb ./cmd/strangedb

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/strangedb .

# Create data directory
RUN mkdir -p /data

# Expose ports
EXPOSE 9000 9001

# Default environment variables
ENV STRANGE_DATA_DIR=/data
ENV STRANGE_HTTP_PORT=9000
ENV STRANGE_GRPC_PORT=9001

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9000/health || exit 1

# Run the binary
ENTRYPOINT ["./strangedb"]

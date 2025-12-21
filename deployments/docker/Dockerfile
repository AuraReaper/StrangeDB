# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o strangedb ./cmd/strangedb

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Copy binary
COPY --from=builder /app/strangedb .

# Create data directory
RUN mkdir -p /data

# Expose ports
EXPOSE 9000 9001

# Set environment
ENV STRANGE_DATA_DIR=/data
ENV STRANGE_HTTP_PORT=9000
ENV STRANGE_GRPC_PORT=9001

# Run
CMD ["./strangedb"]

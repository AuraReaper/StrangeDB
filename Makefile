.PHONY: build test test-coverage clean run fmt lint docker-build

BINARY_NAME := strangedb
BUILD_DIR := build

# Build the server binary
build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/strangedb

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run the server
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run ./...

# Build Docker image
docker-build:
	docker build -t strangedb:latest -f deployments/docker/Dockerfile .

# Makefile

.PHONY: build test clean run

# Build the server binary
build:
	go build -o build/strangedb ./cmd/strangedb

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf build/
	rm -f coverage.out coverage.html

# Run the server
run: build
	./build/strangedb

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run ./...

# Build Docker image
docker:
	docker build -t strangedb:latest -f deployments/docker/Dockerfile .

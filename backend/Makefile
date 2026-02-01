.PHONY: build run test clean deps lint migrate

# Variables
BINARY_NAME=sherlockos
BUILD_DIR=./bin

# Build the application
build:
	@echo "Building..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

# Run the application
run:
	@echo "Running..."
	go run ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run database migrations (requires migrate tool)
migrate-up:
	@echo "Running migrations..."
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	@echo "Rolling back migrations..."
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	@echo "Creating new migration..."
	migrate create -ext sql -dir ./migrations -seq $(name)

# Generate mocks (requires mockgen)
mocks:
	@echo "Generating mocks..."
	mockgen -source=./internal/clients/interfaces.go -destination=./internal/clients/mocks/mock_clients.go -package=mocks

# Development: run with hot reload (requires air)
dev:
	@echo "Running with hot reload..."
	air

# Docker build
docker-build:
	docker build -t sherlockos-backend .

# Docker run
docker-run:
	docker run -p 8080:8080 --env-file .env sherlockos-backend

# MailRaven Makefile

.PHONY: all build test lint clean run help

# Build variables
BINARY_NAME=mailraven
BUILD_DIR=bin
MAIN_PATH=./cmd/mailraven

all: lint test build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

## test: Run all tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -race -tags=integration ./tests/...

## lint: Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/"; \
		go fmt ./...; \
		go vet ./...; \
	fi

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.txt coverage.html
	@rm -f *.db *.db-shm *.db-wal

## run: Build and run the server
run: build
	@echo "Starting $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

## docker-build: Build docker image
docker-build:
	@echo "Building docker image..."
	docker build -f build/Dockerfile -t mailraven:latest .

## docker-run: Run docker container with defaults
docker-run:
	@echo "Running docker container..."
	docker run -p 25:25 -p 80:80 -p 443:443 -v $$(PWD)/data:/data -v $$(PWD)/deployment/config.example.yaml:/app/config.yaml mailraven:latest

## coverage: Generate and open coverage report
coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Opening coverage.html..."

## help: Show this help message
help:
	@echo "MailRaven Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' Makefile

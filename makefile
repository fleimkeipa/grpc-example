.PHONY: all proto test build docker-up docker-down lint clean run help

# Variables
BINARY_NAME=explore-service
PROTO_DIR=proto
COVERAGE_FILE=coverage.out

# Default target
.DEFAULT_GOAL := help

## test: Run tests with coverage
test:
	@echo "🧪 Running tests..."
	go test -v -race -coverprofile=$(COVERAGE_FILE) ./...
	@echo "📊 Generating coverage report..."
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "✅ Tests complete. Open coverage.html to view results"

## build: Build the application binary
build:
	@echo "🏗️  Building..."
	mkdir -p bin
	go build -o bin/$(BINARY_NAME) ./cmd
	@echo "✅ Build complete: bin/$(BINARY_NAME)"

## docker-up: Start Docker containers
docker-up:
	@echo "🐳 Starting Docker containers..."
	docker-compose up --build -d
	@echo "✅ Containers started"

## docker-down: Stop Docker containers
docker-down:
	@echo "🛑 Stopping Docker containers..."
	docker-compose down -v
	@echo "✅ Containers stopped"

## run: Run the application
run:
	@echo "🚀 Starting application..."
	go run cmd/main.go


## clean: Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -rf bin/
	rm -f $(COVERAGE_FILE) coverage.html
	@echo "✅ Clean complete"

## logs: Show Docker logs
logs:
	docker-compose logs -f

## db: Connect to database
db:
	docker exec -it $$(docker-compose ps -q db) psql -U postgres -d explore

## all: Run proto, test, and build
all: proto test build

## help: Show this help message
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
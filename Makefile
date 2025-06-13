.PHONY: build test clean run dev lint fmt vet deps example integration

# Build the main application
build:
	go build -o bin/commentific ./cmd/commentific

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run the main application
run: build
	./bin/commentific

# Run in development mode (with auto-reload would require additional tools)
dev:
	go run ./cmd/commentific/main.go

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	go fmt ./...

# Vet the code
vet:
	go vet ./...

# Tidy dependencies
deps:
	go mod tidy
	go mod download

# Run basic example
example:
	go run ./examples/basic/main.go

# Run integration example
integration:
	go run ./examples/integration/main.go

# Database migration (requires DATABASE_URL environment variable)
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

# Docker commands (if you add Docker support later)
docker-build:
	docker build -t commentific .

docker-run:
	docker run -p 8080:8080 commentific

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  run            - Build and run the application"
	@echo "  dev            - Run in development mode"
	@echo "  lint           - Lint the code"
	@echo "  fmt            - Format the code"
	@echo "  vet            - Vet the code"
	@echo "  deps           - Tidy and download dependencies"
	@echo "  example        - Run basic example"
	@echo "  integration    - Run integration example"
	@echo "  migrate-up     - Run database migrations up"
	@echo "  migrate-down   - Run database migrations down"
	@echo "  help           - Show this help message" 
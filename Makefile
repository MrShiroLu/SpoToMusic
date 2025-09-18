.PHONY: build test clean install deps

# Build the application
build:
	go build -o bin/spotomusic .

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/spotomusic-linux .
	GOOS=windows GOARCH=amd64 go build -o bin/spotomusic.exe .
	GOOS=darwin GOARCH=amd64 go build -o bin/spotomusic-macos .

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install the application
install: build
	go install .

# Run the application
run:
	go run .

# Run with verbose logging
run-verbose:
	go run . --verbose

# Run tests in watch mode (requires entr)
test-watch:
	find . -name "*.go" | entr -c go test ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate documentation
docs:
	godoc -http=:6060

# Setup development environment
setup: deps
	@echo "Setting up development environment..."
	@echo "1. Copy .env.example to .env and fill in your credentials"
	@echo "2. Run 'make test' to verify everything works"
	@echo "3. Run 'make build' to build the application"

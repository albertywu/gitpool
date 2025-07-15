.PHONY: build test test-unit test-integration clean install

# Default target
all: build

# Build the gp binary (gitpool)
build:
	go build -o gp ./gitpool

# Run all tests
test: test-unit test-integration

# Run unit tests
test-unit:
	go test -v ./...

# Run integration tests
test-integration:
	cd tests && ./run_tests.sh

# Run integration tests with verbose output
test-integration-verbose:
	cd tests && VERBOSE=true ./run_tests.sh

# Clean up build artifacts and test files
clean:
	rm -f gp gitpool/gitpool
	rm -f tests/integration_test
	pkill -f "gitpool.*daemon" || true

# Install the binary
install: build
	cp gp /usr/local/bin/gitpool

# Development helpers
dev-build:
	go build -race -o gp ./gitpool

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run

# Help target
help:
	@echo "Available targets:"
	@echo "  build                 - Build the gp binary (gitpool)"
	@echo "  test                  - Run all tests (unit + integration)"
	@echo "  test-unit             - Run unit tests only"
	@echo "  test-integration      - Run integration tests"
	@echo "  test-integration-verbose - Run integration tests with verbose output"
	@echo "  clean                 - Clean up build artifacts"
	@echo "  install               - Install binary to /usr/local/bin"
	@echo "  dev-build             - Build with race detector"
	@echo "  fmt                   - Format code"
	@echo "  vet                   - Run go vet"
	@echo "  lint                  - Run golangci-lint"
	@echo "  help                  - Show this help message" 
#!/bin/bash

# Integration test runner for treefarm CLI

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
TEST_TIMEOUT="300s"
VERBOSE=${VERBOSE:-false}

echo -e "${YELLOW}Running Treefarm Integration Tests${NC}"
echo "=================================="

# Change to tests directory
cd "$(dirname "$0")"

# Clean up any existing test processes
echo "Cleaning up any existing test processes..."
pkill -f "treefarm.*daemon" || true
sleep 1

# Run the tests with timeout
echo "Running integration tests..."
echo

if [ "$VERBOSE" = "true" ]; then
    go test -v -timeout $TEST_TIMEOUT ./...
else
    go test -timeout $TEST_TIMEOUT ./...
fi

exit_code=$?

# Clean up
echo
echo "Cleaning up..."
pkill -f "treefarm.*daemon" || true

if [ $exit_code -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
else
    echo -e "${RED}Some tests failed!${NC}"
fi

exit $exit_code 
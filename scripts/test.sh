#!/bin/bash
# Test script for blockchain indexer

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Running tests for blockchain indexer...${NC}"
echo

# Run tests with coverage
echo -e "${YELLOW}Running unit tests...${NC}"
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

if [ $? -eq 0 ]; then
    echo
    echo -e "${GREEN}✓ All tests passed${NC}"

    # Generate coverage report
    echo
    echo -e "${YELLOW}Generating coverage report...${NC}"
    go tool cover -html=coverage.out -o coverage.html

    # Show coverage summary
    echo
    echo -e "${YELLOW}Coverage summary:${NC}"
    go tool cover -func=coverage.out | tail -n 1

    echo
    echo -e "${GREEN}Coverage report: coverage.html${NC}"
else
    echo
    echo -e "${RED}✗ Tests failed${NC}"
    exit 1
fi

echo
echo -e "${GREEN}Testing complete!${NC}"

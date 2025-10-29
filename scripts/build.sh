#!/bin/bash
# Build script for blockchain indexer

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="blockchain-indexer"
BIN_DIR="./bin"
CMD_DIR="./cmd/indexer"

# Get version info
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo -e "${GREEN}Building ${APP_NAME}...${NC}"
echo "Version: ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Build Date: ${BUILD_DATE}"
echo

# Create bin directory
mkdir -p "${BIN_DIR}"

# Build flags
LDFLAGS="-w -s"
LDFLAGS="${LDFLAGS} -X main.version=${VERSION}"
LDFLAGS="${LDFLAGS} -X main.commit=${COMMIT}"
LDFLAGS="${LDFLAGS} -X main.date=${BUILD_DATE}"

# Build for current platform
echo -e "${YELLOW}Building for current platform...${NC}"
CGO_ENABLED=1 go build \
    -ldflags="${LDFLAGS}" \
    -o "${BIN_DIR}/${APP_NAME}" \
    "${CMD_DIR}"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Build successful${NC}"
    echo "Binary: ${BIN_DIR}/${APP_NAME}"

    # Show binary info
    ls -lh "${BIN_DIR}/${APP_NAME}"

    # Test binary
    echo
    echo -e "${YELLOW}Testing binary...${NC}"
    "${BIN_DIR}/${APP_NAME}" version
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

echo
echo -e "${GREEN}Build complete!${NC}"

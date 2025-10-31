#!/bin/bash
# Docker build script for blockchain indexer

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
IMAGE_NAME="blockchain-indexer"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
REGISTRY="${REGISTRY:-}"

echo -e "${GREEN}Building Docker image for ${IMAGE_NAME}...${NC}"
echo "Version: ${VERSION}"
echo

# Build image
echo -e "${YELLOW}Building Docker image...${NC}"
docker build \
    -t "${IMAGE_NAME}:${VERSION}" \
    -t "${IMAGE_NAME}:latest" \
    .

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Docker build successful${NC}"

    # Show image info
    docker images | grep "${IMAGE_NAME}"

    # Tag for registry if specified
    if [ -n "${REGISTRY}" ]; then
        echo
        echo -e "${YELLOW}Tagging for registry ${REGISTRY}...${NC}"
        docker tag "${IMAGE_NAME}:${VERSION}" "${REGISTRY}/${IMAGE_NAME}:${VERSION}"
        docker tag "${IMAGE_NAME}:latest" "${REGISTRY}/${IMAGE_NAME}:latest"
        echo -e "${GREEN}✓ Tagged for registry${NC}"
    fi
else
    echo -e "${RED}✗ Docker build failed${NC}"
    exit 1
fi

echo
echo -e "${GREEN}Docker build complete!${NC}"
echo
echo "Run with:"
echo "  docker run -p 8080:8080 -p 50051:50051 -p 9091:9091 ${IMAGE_NAME}:${VERSION}"
echo
echo "Or use docker-compose:"
echo "  docker-compose up -d"

# Blockchain Indexer

> Multi-chain blockchain indexer with SOLID principles and Clean Architecture

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Architecture](https://img.shields.io/badge/Architecture-Clean-success)](docs/ARCHITECTURE.md)

**blockchain-indexer**는 멀티체인을 지원하는 확장 가능한 블록체인 인덱서입니다. EVM, Solana, Cosmos, Polkadot, Avalanche, Ripple 등 다양한 블록체인의 블록 및 트랜잭션 데이터를 인덱싱하고, GraphQL, gRPC, REST API를 통해 제공합니다.

---

## Features

### Multi-Chain Support
- **EVM Chains**: Ethereum, BSC, Polygon, Arbitrum, Optimism, etc.
- **Solana**: Solana mainnet, devnet, testnet
- **Cosmos**: Cosmos Hub, Osmosis, and other Cosmos SDK chains
- **Polkadot**: Polkadot, Kusama, and Substrate-based chains
- **Avalanche**: C-Chain, X-Chain, P-Chain
- **Ripple**: XRPL (XRP Ledger)

### Core Features
- **SOLID Architecture**: Follows SOLID principles for maintainability
- **Clean Architecture**: Separated layers (Domain, Application, Infrastructure, Presentation)
- **Pluggable Adapters**: Easy to add new blockchain support
- **Efficient Storage**: PebbleDB for high-performance embedded storage
- **High-Performance Event Bus**: 100M+ events/sec delivery
- **Flexible Querying**: GraphQL, gRPC, and REST API support
- **Real-time Statistics**: Live indexing metrics and chain statistics
- **Gap Detection & Recovery**: Automatic detection and filling of missing blocks
- **Performance Optimization**: LRU caching, batch operations, configurable profiles
- **Health Monitoring**: Comprehensive health checks and observability

### API Support
- **GraphQL API**: Flexible querying with filtering and pagination
- **gRPC API**: High-performance RPC with TLS 1.2+
- **REST API**: Standard HTTP API with TLS 1.2+
- **TLS 1.2+ Support**: Secure communication for all APIs
- **Authentication**: JWT-based authentication support

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│           Presentation Layer (APIs)                 │
│   GraphQL    │    gRPC     │     REST API           │
└──────────────┴─────────────┴────────────────────────┘
                      │
┌─────────────────────▼─────────────────────────────┐
│          Application Layer (Use Cases)             │
│   Indexer │ Block Processor │ Query Handler       │
└──────────────┬────────────────────────────────────┘
               │
┌──────────────▼────────────────────────────────────┐
│           Domain Layer (Business Logic)            │
│   Models  │  Repositories  │  Services            │
└──────────────┬────────────────────────────────────┘
               │
┌──────────────▼────────────────────────────────────┐
│      Infrastructure Layer (External)               │
│  Chain Adapters │ Storage │ Event Bus             │
│  EVM │ Solana │ Cosmos │ Polkadot │ ...           │
└───────────────────────────────────────────────────┘
```

For detailed architecture, see [ARCHITECTURE.md](docs/ARCHITECTURE.md)

---

## Documentation

### Core Documentation
- [Architecture](docs/ARCHITECTURE.md) - System architecture and design principles
- [Directory Structure](docs/DIRECTORY_STRUCTURE.md) - Project layout and organization
- [Implementation Plan](docs/IMPLEMENTATION_PLAN.md) - Phased development plan
- [Deployment Guide](docs/DEPLOYMENT.md) - Docker, systemd, and production deployment
- [CI/CD Guide](docs/CI_CD.md) - GitHub Actions workflows and automation

### API Documentation
- [API Reference](docs/API_REFERENCE.md) - Complete API documentation (GraphQL, gRPC, REST)
- [GraphQL Schema](pkg/presentation/graphql/schema/schema.graphql) - GraphQL schema definition
- [gRPC Protocol Buffers](api/proto/indexer/v1/indexer.proto) - Protocol Buffer definitions

### Operational Guides
- [Chain Adapter Guide](docs/CHAIN_ADAPTER_GUIDE.md) - How to use and add new blockchain support
- [Performance Guide](docs/PERFORMANCE.md) - Optimization strategies and configuration
- [Monitoring Guide](docs/MONITORING.md) - Health checks, metrics, and troubleshooting

---

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Git

### Installation

```bash
# Clone repository
git clone https://github.com/sage-x-project/blockchain-indexer.git
cd blockchain-indexer

# Install dependencies
go mod download

# Copy example config and customize
cp config/config.example.yaml config/config.yaml

# OR use chain-specific configs:
# For Solana:    cp config/config-solana.example.yaml config/config.yaml
# For Cosmos:    cp config/config-cosmos.example.yaml config/config.yaml
# For Polkadot:  cp config/config-polkadot.example.yaml config/config.yaml
# For Avalanche: cp config/config-avalanche.example.yaml config/config.yaml
# For Ripple:    cp config/config-ripple.example.yaml config/config.yaml

# Build the indexer
go build -o bin/indexer ./cmd/indexer
```

### Running the Server

```bash
# Start the API server with your configuration
./bin/indexer server --config config/config.yaml

# The server will start with the following endpoints:
# - REST API:    http://localhost:8080/api
# - GraphQL:     http://localhost:8080/graphql
# - gRPC:        localhost:50051
# - Health:      http://localhost:8080/health
# - Metrics:     http://localhost:9091/metrics
```

### Testing the APIs

```bash
# Test health endpoint
curl http://localhost:8080/health

# Test REST API
curl http://localhost:8080/api/

# Test GraphQL (access GraphQL Playground)
open http://localhost:8080/graphql

# Run integration tests
go test -v ./test/integration/...
```

### Available Commands

```bash
# Show all available commands
./bin/indexer --help

# Start API server
./bin/indexer server --config config/config.yaml

# Start blockchain indexing
./bin/indexer index --config config/config.yaml

# Show version information
./bin/indexer version

# Manage configuration
./bin/indexer config --help
```

### Docker Deployment

```bash
# Build and run with Docker Compose (includes Prometheus & Grafana)
docker-compose up -d

# Access services:
# - Indexer API: http://localhost:8080
# - Prometheus: http://localhost:9090
# - Grafana: http://localhost:3000 (admin/admin)

# View logs
docker-compose logs -f indexer

# Stop services
docker-compose down
```

For detailed deployment instructions, see [DEPLOYMENT.md](docs/DEPLOYMENT.md)

---

## Implementation Status

The blockchain indexer has been developed through a phased approach, with all core phases now complete:

**Phase 1-4**: Foundation, core functionality, and APIs
**Phase 5-6**: Multi-chain support (Solana, Cosmos, Polkadot, Avalanche, Ripple)
**Phase 7-8**: Production deployment and CI/CD automation
**Phase 9**: Advanced features including statistics, gap detection, performance optimization, and monitoring

For detailed implementation history, see [IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md) and [PHASE_9_PLAN.md](docs/PHASE_9_PLAN.md).

**Current Version**: 1.0.0-rc1
**Status**: Production ready - all core features implemented

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Status**: Production Ready (Phase 9 Complete - Fully Operational)

**Current Version**: 1.0.0-rc1

**Last Updated**: 2025-10-31
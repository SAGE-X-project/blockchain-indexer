# Blockchain Indexer

> Multi-chain blockchain indexer with SOLID principles and Clean Architecture

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Architecture](https://img.shields.io/badge/Architecture-Clean-success)](docs/ARCHITECTURE.md)

**blockchain-indexer**는 멀티체인을 지원하는 확장 가능한 블록체인 인덱서입니다. EVM, Solana, Cosmos, Polkadot, Avalanche, Ripple 등 다양한 블록체인의 블록 및 트랜잭션 데이터를 인덱싱하고, GraphQL, gRPC, REST API를 통해 제공합니다.

---

## 🎯 Features

### Multi-Chain Support
- ✅ **EVM Chains**: Ethereum, BSC, Polygon, Arbitrum, Optimism, etc.
- ✅ **Solana**: Solana mainnet, devnet, testnet
- 🚧 **Cosmos**: Cosmos Hub, Osmosis, and other Cosmos SDK chains
- 🚧 **Polkadot**: Polkadot, Kusama, and Substrate-based chains
- 🚧 **Avalanche**: C-Chain, X-Chain, P-Chain
- 🚧 **Ripple**: XRPL (XRP Ledger)

### Core Features
- 🏗️ **SOLID Architecture**: Follows SOLID principles for maintainability
- 🧩 **Clean Architecture**: Separated layers (Domain, Application, Infrastructure, Presentation)
- 🔌 **Pluggable Adapters**: Easy to add new blockchain support
- 💾 **Efficient Storage**: PebbleDB for high-performance embedded storage
- ⚡ **High-Performance Event Bus**: 100M+ events/sec delivery
- 🔍 **Flexible Querying**: GraphQL, gRPC, and REST API support

### API Support
- ✅ **GraphQL API**: Flexible querying with filtering and pagination
- ✅ **gRPC API**: High-performance RPC with TLS 1.2+
- ✅ **REST API**: Standard HTTP API with TLS 1.2+
- 🔒 **TLS 1.2+ Support**: Secure communication for all APIs
- 🔑 **Authentication**: JWT-based authentication support

---

## 🏗️ Architecture

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

## 📖 Documentation

### Core Documentation
- 📄 [Architecture](docs/ARCHITECTURE.md) - System architecture and design principles
- 📄 [Directory Structure](docs/DIRECTORY_STRUCTURE.md) - Project layout and organization
- 📄 [Implementation Plan](docs/IMPLEMENTATION_PLAN.md) - Phased development plan
- 📄 [Deployment Guide](docs/DEPLOYMENT.md) - Docker, systemd, and production deployment
- 📄 [CI/CD Guide](docs/CI_CD.md) - GitHub Actions workflows and automation

### API Documentation (Coming Soon)
- 📄 API Reference - Complete API documentation
- 📄 GraphQL Schema - GraphQL schema definition
- 📄 gRPC Protocol Buffers - Protocol Buffer definitions

### Development Guides (Coming Soon)
- 📄 Chain Adapter Guide - How to add new blockchain support

---

## 🚀 Quick Start

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
# OR for Solana
cp config/config-solana.example.yaml config/config.yaml

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

## 🛣️ Roadmap

### Phase 1: Foundation ✅ 100%
- [x] Core architecture design
- [x] Domain models and interfaces
- [x] PebbleDB storage implementation
- [x] Configuration management
- [x] Logging infrastructure
- [x] Metrics and monitoring

### Phase 2: EVM Support ✅ 100%
- [x] EVM chain adapter
- [x] Block indexer
- [x] Transaction processor
- [x] Integration tests

### Phase 3: APIs ✅ 100%
- [x] Event bus (100M+ events/sec)
- [x] GraphQL API with subscriptions
- [x] gRPC API with streaming
- [x] REST API with comprehensive endpoints

### Phase 4: Main Application ✅ 100%
- [x] CLI interface
- [x] Server initialization
- [x] Graceful shutdown
- [x] Integration tests

### Phase 5: Solana Support ✅ 100%
- [x] Solana adapter architecture
- [x] Solana RPC client wrapper
- [x] Solana data normalizer
- [x] Configuration support

### Phase 7: Production Features ✅ 100%
- [x] Docker containerization
- [x] Docker Compose with monitoring stack
- [x] Prometheus metrics integration
- [x] Grafana dashboards provisioning
- [x] Systemd service configuration
- [x] Deployment scripts and automation
- [x] Deployment documentation

### Phase 8: CI/CD & Automation ✅ 100%
- [x] GitHub Actions CI workflow
- [x] GitHub Actions release workflow
- [x] Multi-platform binary builds
- [x] Docker image publishing (GHCR)
- [x] Dependabot configuration
- [x] Automated testing and coverage
- [x] Security scanning
- [x] CI/CD documentation

### Phase 6: Additional Chains 🔜
- [ ] Cosmos adapter (Tendermint-based chains)
- [ ] Polkadot adapter (Substrate-based chains)
- [ ] Avalanche adapter (C-Chain, X-Chain, P-Chain)
- [ ] Ripple adapter (XRPL)
- [ ] Chain adapter registry
- [ ] Comprehensive API documentation

See [IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md) for full roadmap.

**Current Status**: Phase 8 Complete - Fully Automated CI/CD Pipeline

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Status**: ✅ Production Ready (Phase 4 Complete)

**Current Version**: 0.1.0-beta

**Last Updated**: 2025-10-30
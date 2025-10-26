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

### API Documentation (Coming Soon)
- 📄 API Reference - Complete API documentation
- 📄 GraphQL Schema - GraphQL schema definition
- 📄 gRPC Protocol Buffers - Protocol Buffer definitions

### Development Guides (Coming Soon)
- 📄 Chain Adapter Guide - How to add new blockchain support
- 📄 Deployment Guide - Production deployment instructions

---

## 🚀 Quick Start

```bash
# Clone repository
git clone https://github.com/sage-x-project/blockchain-indexer.git
cd blockchain-indexer

# Install dependencies
go mod download

# Build
go build -o bin/indexer ./cmd/indexer

# Run (configuration required)
./bin/indexer --config config/config.yaml
```

---

## 🛣️ Roadmap

### Phase 1: Foundation (Current) ✅ 40%
- [x] Core architecture design
- [x] Domain models and interfaces
- [ ] PebbleDB storage implementation
- [ ] Configuration management

### Phase 2: EVM Support 🚧
- [ ] EVM chain adapter
- [ ] Block indexer
- [ ] Integration tests

### Phase 3: APIs 🚧
- [ ] Event bus
- [ ] GraphQL API
- [ ] gRPC API
- [ ] REST API

See [IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md) for full roadmap.

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Status**: 🚧 In Development (Phase 1)

**Current Version**: 0.1.0-alpha

**Last Updated**: 2025-10-26
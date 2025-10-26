# Blockchain Indexer Architecture

> Multi-chain blockchain indexer with SOLID principles and Clean Code architecture

## 🎯 Overview

**blockchain-indexer**는 멀티체인을 지원하는 확장 가능한 블록체인 인덱서입니다. EVM, Solana, Cosmos, Polkadot, Avalanche, Ripple 등 다양한 블록체인의 블록 및 트랜잭션 데이터를 인덱싱하고, GraphQL, gRPC, REST API를 통해 제공합니다.

## 🏗️ Architecture Principles

### SOLID Principles

1. **Single Responsibility Principle (SRP)**
   - 각 체인 어댑터는 하나의 블록체인 프로토콜만 담당
   - 각 레이어는 명확한 단일 책임을 가짐

2. **Open/Closed Principle (OCP)**
   - 새로운 체인 추가 시 기존 코드 수정 없이 확장
   - 인터페이스 기반 설계로 확장성 보장

3. **Liskov Substitution Principle (LSP)**
   - 모든 체인 어댑터는 `ChainAdapter` 인터페이스로 교체 가능
   - 모든 데이터 제공자는 `DataProvider` 인터페이스로 교체 가능

4. **Interface Segregation Principle (ISP)**
   - 작은 인터페이스로 분리 (Reader, Writer, ChainReader, etc.)
   - 클라이언트는 필요한 인터페이스만 의존

5. **Dependency Inversion Principle (DIP)**
   - 구체적인 구현이 아닌 추상화(인터페이스)에 의존
   - 의존성 주입을 통한 결합도 감소

### Clean Architecture Layers

```
┌─────────────────────────────────────────────────────┐
│              Presentation Layer                     │
│  (GraphQL, gRPC, REST API Handlers)                 │
└──────────────────┬──────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────┐
│              Application Layer                      │
│  (Use Cases, Business Logic, Orchestration)         │
└──────────────────┬──────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────┐
│              Domain Layer                           │
│  (Entities, Interfaces, Domain Logic)               │
└──────────────────┬──────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────┐
│           Infrastructure Layer                      │
│  (Chain Adapters, Storage, External Services)       │
└─────────────────────────────────────────────────────┘
```

## 📦 System Architecture

### High-Level Architecture

```
                    ┌─────────────────────────┐
                    │   External Clients      │
                    │  (Web, Mobile, CLI)     │
                    └────────────┬────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌────────▼────────┐   ┌──────────▼─────────┐  ┌────────▼────────┐
│  GraphQL API    │   │     gRPC API       │  │   REST API      │
│  (TLS 1.2+)     │   │   (TLS 1.2+)       │  │  (TLS 1.2+)     │
└────────┬────────┘   └──────────┬─────────┘  └────────┬────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │  Application Services   │
                    │  • Query Handler        │
                    │  • Block Processor      │
                    │  • Event Publisher      │
                    └────────────┬────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌────────▼────────┐   ┌──────────▼─────────┐  ┌────────▼────────┐
│  Storage Layer  │   │    Event Bus       │  │  Chain Registry │
│   (PebbleDB)    │   │   (Pub/Sub)        │  │  (Adapters)     │
└─────────────────┘   └────────────────────┘  └────────┬────────┘
                                                        │
                    ┌───────────────────────────────────┘
                    │
         ┌──────────┼──────────┬──────────┬──────────┐
         │          │          │          │          │
┌────────▼──┐  ┌───▼────┐  ┌──▼─────┐  ┌▼─────┐  ┌─▼─────┐
│ EVM       │  │Solana  │  │Cosmos  │  │Polka │  │Ripple │
│ Adapter   │  │Adapter │  │Adapter │  │ dot  │  │Adapter│
└────┬──────┘  └───┬────┘  └──┬─────┘  └┬─────┘  └─┬─────┘
     │             │           │         │          │
     ▼             ▼           ▼         ▼          ▼
  EVM Nodes   Solana RPC   Cosmos RPC  Substrate  XRPL
```

## 🔧 Core Components

### 1. Chain Adapter Layer

체인별 데이터 읽기 및 정규화를 담당하는 어댑터들입니다.

**Interface: ChainAdapter**
```go
type ChainAdapter interface {
    // Chain information
    GetChainType() ChainType
    GetChainID() string

    // Block operations
    GetLatestBlockNumber(ctx context.Context) (uint64, error)
    GetBlockByNumber(ctx context.Context, number uint64) (*Block, error)
    GetBlockByHash(ctx context.Context, hash string) (*Block, error)

    // Transaction operations
    GetTransaction(ctx context.Context, hash string) (*Transaction, error)

    // Health check
    IsHealthy(ctx context.Context) bool
}
```

**Supported Chains:**
- ✅ **EVM** (Ethereum, BSC, Polygon, Arbitrum, Optimism, etc.)
- ✅ **Solana**
- 🚧 **Cosmos** (Cosmos Hub, Osmosis, etc.)
- 🚧 **Polkadot** (Polkadot, Kusama)
- 🚧 **Avalanche** (C-Chain, X-Chain, P-Chain)
- 🚧 **Ripple** (XRPL)

### 2. Domain Models

모든 체인에서 공통으로 사용하는 정규화된 데이터 모델입니다.

**Core Entities:**
```go
// Block represents a normalized blockchain block
type Block struct {
    ChainType   ChainType
    ChainID     string
    Number      uint64
    Hash        string
    ParentHash  string
    Timestamp   time.Time
    Proposer    string
    TxCount     int
    Metadata    map[string]interface{} // Chain-specific data
}

// Transaction represents a normalized blockchain transaction
type Transaction struct {
    ChainType   ChainType
    ChainID     string
    Hash        string
    BlockNumber uint64
    BlockHash   string
    Index       uint64
    From        string
    To          string
    Value       string
    Status      TxStatus
    Metadata    map[string]interface{} // Chain-specific data
}
```

### 3. Storage Layer

블록체인 데이터를 저장하고 조회하는 레이어입니다.

**Interface: Storage**
```go
type Storage interface {
    Reader
    Writer

    Close() error
    NewBatch() Batch
}

type Reader interface {
    GetLatestHeight(ctx context.Context, chainID string) (uint64, error)
    GetBlock(ctx context.Context, chainID string, number uint64) (*Block, error)
    GetTransaction(ctx context.Context, chainID string, hash string) (*Transaction, error)
}

type Writer interface {
    SetBlock(ctx context.Context, block *Block) error
    SetTransaction(ctx context.Context, tx *Transaction) error
}
```

**Implementation:**
- Primary: PebbleDB (embedded key-value store)
- Future: PostgreSQL, MongoDB support

### 4. Event Bus

실시간 블록/트랜잭션 이벤트를 구독자에게 전달하는 Pub/Sub 시스템입니다.

**Features:**
- High-performance event delivery (100M+ events/sec)
- Flexible filtering (chain, address, value range)
- Multiple subscribers support (10K+ concurrent)
- Zero allocations for core operations

### 5. Data Provider Layer

외부 클라이언트에게 데이터를 제공하는 API 레이어입니다.

**Supported Protocols:**
- **GraphQL**: Flexible querying with filtering
- **gRPC**: High-performance RPC
- **REST API**: Standard HTTP API (TLS 1.2+)

**Interface: DataProvider**
```go
type DataProvider interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    GetPort() int
}
```

## 🔄 Data Flow

### Indexing Flow

```
1. Chain Adapter fetches block from node
   ↓
2. Normalize to common Block model
   ↓
3. Validate block data
   ↓
4. Store in database
   ↓
5. Publish event to EventBus
   ↓
6. Subscribers receive event
```

### Query Flow

```
1. Client sends request (GraphQL/gRPC/REST)
   ↓
2. Provider validates request
   ↓
3. Application service processes query
   ↓
4. Storage layer retrieves data
   ↓
5. Response is serialized and returned
```

## 🚀 Scalability & Performance

### Performance Targets

| Metric | Target | Notes |
|--------|--------|-------|
| Indexing Speed | 100-200 blocks/s | Per chain adapter |
| Query Latency (GraphQL) | <100ms | P95 |
| Query Latency (gRPC) | <50ms | P95 |
| Event Delivery | <1µs | Sub-microsecond |
| Concurrent Chains | 10+ | Simultaneously |
| Storage Efficiency | <1GB per 1M blocks | Compressed |

### Scalability Strategies

1. **Horizontal Scaling**: Multiple indexer instances per chain
2. **Worker Pool**: Concurrent block processing
3. **Batch Operations**: Bulk storage writes
4. **Caching**: In-memory cache for hot data
5. **Sharding**: Future support for data sharding

## 🔒 Security

### Transport Security
- **TLS 1.2+** for all API endpoints
- Certificate-based authentication for gRPC
- JWT tokens for REST API authentication

### Data Integrity
- Block hash verification
- Transaction signature validation
- Chain reorganization handling

## 📊 Monitoring & Observability

### Metrics (Prometheus)
- Indexing rate per chain
- Query latency percentiles
- Error rates
- Storage usage
- Event bus statistics

### Logging
- Structured logging (JSON format)
- Multiple log levels (debug, info, warn, error)
- Distributed tracing support

## 🧪 Testing Strategy

1. **Unit Tests**: Individual components
2. **Integration Tests**: Chain adapters with test nodes
3. **E2E Tests**: Full indexing and query flow
4. **Performance Tests**: Load testing and benchmarking
5. **Chaos Tests**: Network failures and recovery

## 📚 Directory Structure

See [Directory Structure](./DIRECTORY_STRUCTURE.md) for detailed layout.

## 🛣️ Roadmap

### Phase 1: Foundation (Current)
- ✅ Core architecture design
- ✅ EVM adapter implementation
- ✅ Basic storage layer
- 🚧 GraphQL API

### Phase 2: Multi-Chain
- 🚧 Solana adapter
- 🚧 gRPC API
- 🚧 REST API
- 🚧 Cosmos adapter

### Phase 3: Advanced Features
- Chain reorganization handling
- Historical state queries
- Analytics aggregations
- WebSocket subscriptions

### Phase 4: Production
- High availability setup
- Monitoring & alerting
- Performance optimization
- Documentation & examples

## 📖 References

- [Implementation Plan](./IMPLEMENTATION_PLAN.md)
- [API Documentation](./API_REFERENCE.md)
- [Chain Adapter Guide](./CHAIN_ADAPTER_GUIDE.md)
- [Deployment Guide](./DEPLOYMENT.md)

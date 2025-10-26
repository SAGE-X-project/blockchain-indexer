# Directory Structure

> Detailed directory structure for blockchain-indexer project

## 📁 Project Layout

```
blockchain-indexer/
├── cmd/                           # Application entry points
│   ├── indexer/                  # Main indexer application
│   │   └── main.go
│   └── cli/                      # CLI tools
│       └── main.go
│
├── internal/                     # Private application code
│   ├── config/                   # Configuration management
│   │   ├── config.go
│   │   ├── loader.go
│   │   └── validator.go
│   │
│   ├── logger/                   # Logging utilities
│   │   ├── logger.go
│   │   └── logger_test.go
│   │
│   └── metrics/                  # Metrics and monitoring
│       ├── prometheus.go
│       └── collector.go
│
├── pkg/                          # Public library code
│   ├── domain/                   # Domain layer (Clean Architecture)
│   │   ├── models/              # Domain entities
│   │   │   ├── block.go
│   │   │   ├── transaction.go
│   │   │   ├── chain.go
│   │   │   └── types.go
│   │   │
│   │   ├── repository/          # Repository interfaces
│   │   │   ├── block_repository.go
│   │   │   ├── transaction_repository.go
│   │   │   └── storage.go
│   │   │
│   │   └── service/             # Domain services interfaces
│   │       ├── chain_adapter.go
│   │       ├── indexer.go
│   │       └── provider.go
│   │
│   ├── application/              # Application layer (Use Cases)
│   │   ├── indexer/             # Indexing use cases
│   │   │   ├── block_indexer.go
│   │   │   ├── transaction_indexer.go
│   │   │   └── chain_manager.go
│   │   │
│   │   ├── query/               # Query use cases
│   │   │   ├── block_query.go
│   │   │   ├── transaction_query.go
│   │   │   └── chain_query.go
│   │   │
│   │   └── processor/           # Block/TX processors
│   │       ├── block_processor.go
│   │       ├── transaction_processor.go
│   │       └── event_publisher.go
│   │
│   ├── infrastructure/           # Infrastructure layer
│   │   ├── adapter/             # Chain adapters
│   │   │   ├── registry.go      # Adapter registry
│   │   │   │
│   │   │   ├── evm/            # EVM adapter
│   │   │   │   ├── adapter.go
│   │   │   │   ├── client.go
│   │   │   │   ├── normalizer.go
│   │   │   │   └── adapter_test.go
│   │   │   │
│   │   │   ├── solana/         # Solana adapter
│   │   │   │   ├── adapter.go
│   │   │   │   ├── client.go
│   │   │   │   ├── normalizer.go
│   │   │   │   └── adapter_test.go
│   │   │   │
│   │   │   ├── cosmos/         # Cosmos adapter
│   │   │   │   ├── adapter.go
│   │   │   │   ├── client.go
│   │   │   │   ├── normalizer.go
│   │   │   │   └── adapter_test.go
│   │   │   │
│   │   │   ├── polkadot/       # Polkadot adapter
│   │   │   │   ├── adapter.go
│   │   │   │   ├── client.go
│   │   │   │   ├── normalizer.go
│   │   │   │   └── adapter_test.go
│   │   │   │
│   │   │   ├── avalanche/      # Avalanche adapter
│   │   │   │   ├── adapter.go
│   │   │   │   ├── client.go
│   │   │   │   ├── normalizer.go
│   │   │   │   └── adapter_test.go
│   │   │   │
│   │   │   └── ripple/         # Ripple adapter
│   │   │       ├── adapter.go
│   │   │       ├── client.go
│   │   │       ├── normalizer.go
│   │   │       └── adapter_test.go
│   │   │
│   │   ├── storage/             # Storage implementations
│   │   │   ├── pebble/         # PebbleDB implementation
│   │   │   │   ├── storage.go
│   │   │   │   ├── block_repository.go
│   │   │   │   ├── transaction_repository.go
│   │   │   │   ├── encoder.go
│   │   │   │   ├── schema.go
│   │   │   │   └── storage_test.go
│   │   │   │
│   │   │   ├── postgres/       # PostgreSQL (future)
│   │   │   │   └── storage.go
│   │   │   │
│   │   │   └── cache/          # In-memory cache
│   │   │       ├── cache.go
│   │   │       └── lru.go
│   │   │
│   │   └── event/              # Event bus
│   │       ├── bus.go
│   │       ├── subscriber.go
│   │       ├── filter.go
│   │       ├── types.go
│   │       └── bus_test.go
│   │
│   ├── presentation/             # Presentation layer (APIs)
│   │   ├── graphql/             # GraphQL API
│   │   │   ├── handler.go
│   │   │   ├── resolver.go
│   │   │   ├── schema.go
│   │   │   ├── types.go
│   │   │   └── handler_test.go
│   │   │
│   │   ├── grpc/                # gRPC API
│   │   │   ├── server.go
│   │   │   ├── handler.go
│   │   │   ├── interceptor.go
│   │   │   └── server_test.go
│   │   │
│   │   ├── rest/                # REST API
│   │   │   ├── server.go
│   │   │   ├── handler.go
│   │   │   ├── middleware.go
│   │   │   ├── routes.go
│   │   │   └── server_test.go
│   │   │
│   │   └── common/              # Common API utilities
│   │       ├── response.go
│   │       ├── error.go
│   │       └── pagination.go
│   │
│   └── shared/                   # Shared utilities
│       ├── errors/              # Custom errors
│       │   ├── errors.go
│       │   └── codes.go
│       │
│       ├── validator/           # Data validation
│       │   └── validator.go
│       │
│       └── utils/               # Helper functions
│           ├── hash.go
│           ├── encoding.go
│           └── time.go
│
├── api/                          # API definitions
│   ├── proto/                   # Protocol Buffers (gRPC)
│   │   ├── indexer.proto
│   │   ├── block.proto
│   │   ├── transaction.proto
│   │   └── generated/           # Generated code
│   │
│   └── graphql/                 # GraphQL schemas
│       ├── schema.graphql
│       └── generated/           # Generated code
│
├── config/                       # Configuration files
│   ├── config.yaml              # Default configuration
│   ├── config.example.yaml      # Example configuration
│   └── chains/                  # Chain-specific configs
│       ├── ethereum.yaml
│       ├── solana.yaml
│       └── cosmos.yaml
│
├── deployments/                  # Deployment configurations
│   ├── docker/                  # Docker files
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   │
│   ├── kubernetes/              # Kubernetes manifests
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── configmap.yaml
│   │
│   └── systemd/                 # Systemd service files
│       └── blockchain-indexer.service
│
├── scripts/                      # Utility scripts
│   ├── build.sh                 # Build script
│   ├── test.sh                  # Test script
│   ├── deploy.sh                # Deployment script
│   └── generate.sh              # Code generation script
│
├── docs/                         # Documentation
│   ├── ARCHITECTURE.md          # Architecture documentation
│   ├── DIRECTORY_STRUCTURE.md   # This file
│   ├── IMPLEMENTATION_PLAN.md   # Implementation plan
│   ├── API_REFERENCE.md         # API reference
│   ├── CHAIN_ADAPTER_GUIDE.md   # Chain adapter guide
│   └── DEPLOYMENT.md            # Deployment guide
│
├── test/                         # Test files
│   ├── integration/             # Integration tests
│   │   ├── evm_test.go
│   │   ├── solana_test.go
│   │   └── storage_test.go
│   │
│   ├── e2e/                     # End-to-end tests
│   │   ├── indexing_test.go
│   │   └── api_test.go
│   │
│   └── testdata/                # Test data
│       ├── blocks/
│       └── transactions/
│
├── tools/                        # Development tools
│   └── tools.go                 # Tool dependencies
│
├── go.mod                        # Go module file
├── go.sum                        # Go dependencies
├── Makefile                      # Build automation
├── README.md                     # Project README
└── LICENSE                       # License file
```

## 📂 Directory Descriptions

### `cmd/`
Application entry points. Each subdirectory contains a `main.go` file for a specific executable.

- **`indexer/`**: Main blockchain indexer application
- **`cli/`**: Command-line interface tools for administration

### `internal/`
Private application code that cannot be imported by other projects.

- **`config/`**: Configuration loading, validation, and management
- **`logger/`**: Logging setup and utilities
- **`metrics/`**: Prometheus metrics and monitoring

### `pkg/`
Public library code that can be imported by other projects.

#### `pkg/domain/` - Domain Layer
Core business entities and interfaces (Clean Architecture).

- **`models/`**: Domain entities (Block, Transaction, Chain)
- **`repository/`**: Repository interfaces for data access
- **`service/`**: Domain service interfaces

#### `pkg/application/` - Application Layer
Use cases and business logic orchestration.

- **`indexer/`**: Block and transaction indexing logic
- **`query/`**: Query handling and data retrieval
- **`processor/`**: Block and transaction processing

#### `pkg/infrastructure/` - Infrastructure Layer
External concerns and technical implementations.

- **`adapter/`**: Chain-specific adapters (EVM, Solana, Cosmos, etc.)
  - Each adapter is self-contained with its own client, normalizer, and tests
- **`storage/`**: Storage implementations (PebbleDB, PostgreSQL, Cache)
- **`event/`**: Event bus for pub/sub functionality

#### `pkg/presentation/` - Presentation Layer
API implementations and request handlers.

- **`graphql/`**: GraphQL API server
- **`grpc/`**: gRPC API server
- **`rest/`**: REST API server
- **`common/`**: Shared API utilities

#### `pkg/shared/` - Shared Utilities
Common utilities used across layers.

- **`errors/`**: Custom error types
- **`validator/`**: Data validation utilities
- **`utils/`**: Helper functions

### `api/`
API definitions and generated code.

- **`proto/`**: Protocol Buffer definitions for gRPC
- **`graphql/`**: GraphQL schema definitions

### `config/`
Configuration files for different environments and chains.

- **`chains/`**: Chain-specific configuration files

### `deployments/`
Deployment configurations for different platforms.

- **`docker/`**: Docker and Docker Compose files
- **`kubernetes/`**: Kubernetes manifests
- **`systemd/`**: Systemd service files

### `scripts/`
Build, test, and deployment automation scripts.

### `docs/`
Project documentation.

### `test/`
Test files and test data.

- **`integration/`**: Integration tests
- **`e2e/`**: End-to-end tests
- **`testdata/`**: Test fixtures and data

### `tools/`
Development tools and dependencies.

## 🎯 Design Patterns

### Dependency Injection
All components use constructor injection for dependencies.

```go
// Good: Constructor injection
func NewBlockIndexer(
    adapter service.ChainAdapter,
    storage repository.BlockRepository,
    eventBus *event.Bus,
) *BlockIndexer {
    return &BlockIndexer{
        adapter:  adapter,
        storage:  storage,
        eventBus: eventBus,
    }
}
```

### Repository Pattern
Data access is abstracted through repository interfaces.

```go
// Interface in domain layer
type BlockRepository interface {
    GetBlock(ctx context.Context, chainID string, height uint64) (*models.Block, error)
    SaveBlock(ctx context.Context, block *models.Block) error
}

// Implementation in infrastructure layer
type PebbleBlockRepository struct {
    db *pebble.DB
}
```

### Adapter Pattern
Chain-specific logic is encapsulated in adapters.

```go
// All adapters implement the same interface
type ChainAdapter interface {
    GetLatestBlockNumber(ctx context.Context) (uint64, error)
    GetBlock(ctx context.Context, number uint64) (*models.Block, error)
}
```

### Factory Pattern
Adapters are created through a registry/factory.

```go
registry := adapter.NewRegistry()
registry.Register("ethereum", evm.NewAdapter)
registry.Register("solana", solana.NewAdapter)

adapter, err := registry.Create("ethereum", config)
```

## 📝 Naming Conventions

### Files
- **Lowercase with underscores**: `block_repository.go`
- **Test files**: `block_repository_test.go`
- **Interface files**: Named after the main interface (`storage.go` for `Storage` interface)

### Packages
- **Lowercase, single word**: `domain`, `storage`, `adapter`
- **Descriptive names**: Avoid generic names like `utils` or `common` when possible

### Interfaces
- **Descriptive names**: `BlockRepository`, `ChainAdapter`
- **Suffix with -er for actions**: `Indexer`, `Processor`, `Handler`

### Structs
- **PascalCase**: `BlockIndexer`, `EVMAdapter`
- **Avoid stuttering**: `evm.Adapter` instead of `evm.EVMAdapter`

## 🔄 Import Paths

```go
import (
    // Domain layer
    "github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
    "github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"

    // Application layer
    "github.com/sage-x-project/blockchain-indexer/pkg/application/indexer"

    // Infrastructure layer
    "github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/evm"
    "github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/storage/pebble"

    // Presentation layer
    "github.com/sage-x-project/blockchain-indexer/pkg/presentation/graphql"

    // Internal packages
    "github.com/sage-x-project/blockchain-indexer/internal/config"
    "github.com/sage-x-project/blockchain-indexer/internal/logger"
)
```

## 📖 References

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [SOLID Principles in Go](https://dave.cheney.net/2016/08/20/solid-go-design)

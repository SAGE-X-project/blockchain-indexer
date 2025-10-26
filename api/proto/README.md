# Protocol Buffers Definitions

This directory contains Protocol Buffer definitions for the blockchain indexer gRPC API.

## Overview

The gRPC API provides high-performance access to indexed blockchain data with:
- Unary RPCs for single request/response operations
- Server streaming for real-time updates
- Type-safe contract definitions
- Efficient binary serialization

## Service Definition

See `indexer/v1/indexer.proto` for the complete service definition.

### Main Services

- **IndexerService**: Primary service for accessing indexed blockchain data

### Key Operations

#### Chain Operations
- `GetChain`: Retrieve chain information
- `ListChains`: List all indexed chains

#### Block Operations
- `GetBlock`: Get block by number
- `GetBlockByHash`: Get block by hash
- `ListBlocks`: List blocks with pagination
- `GetLatestBlock`: Get the latest indexed block

#### Transaction Operations
- `GetTransaction`: Get transaction by hash
- `ListTransactionsByBlock`: Get all transactions in a block
- `ListTransactionsByAddress`: Get transactions for an address

#### Progress Operations
- `GetProgress`: Get indexing progress for a chain

#### Streaming Operations
- `StreamBlocks`: Real-time stream of newly indexed blocks
- `StreamTransactions`: Real-time stream of transactions
- `StreamProgress`: Live progress updates

## Code Generation

To generate gRPC code from these proto files:

### Prerequisites

```bash
# Install protoc compiler
# macOS
brew install protobuf

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Generate Code

```bash
# From project root
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/indexer/v1/indexer.proto
```

This will generate:
- `indexer.pb.go`: Message type definitions
- `indexer_grpc.pb.go`: Service definitions and client/server stubs

## Usage Example

### Server

```go
import (
    indexerv1 "github.com/sage-x-project/blockchain-indexer/api/proto/indexer/v1"
    "google.golang.org/grpc"
)

// Implement the service
type server struct {
    indexerv1.UnimplementedIndexerServiceServer
    // ... dependencies
}

func (s *server) GetBlock(ctx context.Context, req *indexerv1.GetBlockRequest) (*indexerv1.GetBlockResponse, error) {
    // Implementation
}

// Start server
lis, _ := net.Listen("tcp", ":50051")
grpcServer := grpc.NewServer()
indexerv1.RegisterIndexerServiceServer(grpcServer, &server{})
grpcServer.Serve(lis)
```

### Client

```go
import (
    indexerv1 "github.com/sage-x-project/blockchain-indexer/api/proto/indexer/v1"
    "google.golang.org/grpc"
)

// Connect to server
conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
defer conn.Close()

client := indexerv1.NewIndexerServiceClient(conn)

// Call methods
resp, err := client.GetBlock(ctx, &indexerv1.GetBlockRequest{
    ChainId: "ethereum",
    Number: 1000,
})
```

### Streaming

```go
// Stream blocks
stream, err := client.StreamBlocks(ctx, &indexerv1.StreamBlocksRequest{
    ChainId: "ethereum",
})

for {
    block, err := stream.Recv()
    if err != nil {
        break
    }
    // Process block
}
```

## API Design

### Naming Conventions
- Services: `ServiceName` + `Service` (e.g., `IndexerService`)
- Methods: Verb + Noun (e.g., `GetBlock`, `ListTransactions`)
- Messages: Method name + `Request`/`Response`

### Pagination
- Use `page_size` and `page_token` for cursor-based pagination
- Return `next_page_token` for fetching next page

### Streaming
- Prefix with `Stream` for server streaming RPCs
- Client specifies filters in request
- Server sends updates as they occur

## Error Handling

gRPC uses standard status codes:
- `OK`: Success
- `NOT_FOUND`: Resource not found
- `INVALID_ARGUMENT`: Invalid request parameters
- `INTERNAL`: Internal server error
- `UNAVAILABLE`: Service temporarily unavailable

## Performance

Benefits of gRPC:
- Binary protocol (faster than JSON)
- HTTP/2 multiplexing
- Streaming support
- Strong typing
- Code generation for multiple languages

## Next Steps

1. Generate code from proto definitions
2. Implement server with repository integration
3. Add authentication/authorization
4. Deploy with proper load balancing
5. Add monitoring and tracing

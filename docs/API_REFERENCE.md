# API Reference

> Comprehensive API documentation for blockchain-indexer

blockchain-indexer provides three API interfaces for querying indexed blockchain data:
- **GraphQL API** - Flexible queries with real-time subscriptions
- **gRPC API** - High-performance RPC with streaming
- **REST API** - Standard HTTP/JSON interface

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Authentication](#authentication)
3. [GraphQL API](#graphql-api)
4. [gRPC API](#grpc-api)
5. [REST API](#rest-api)
6. [Common Patterns](#common-patterns)
7. [Error Handling](#error-handling)
8. [Rate Limiting](#rate-limiting)

---

## Quick Start

### Starting the Server

```bash
# Start with default configuration
./bin/indexer server --config config/config.yaml

# The server will start with:
# - REST API:    http://localhost:8080/api
# - GraphQL:     http://localhost:8080/graphql
# - gRPC:        localhost:50051
# - Health:      http://localhost:8080/health
# - Metrics:     http://localhost:9091/metrics
```

### Testing Connectivity

```bash
# Test health endpoint
curl http://localhost:8080/health

# Test REST API
curl http://localhost:8080/api/

# Access GraphQL Playground
open http://localhost:8080/graphql
```

---

## Authentication

Currently, the API supports basic authentication. TLS 1.2+ is supported for secure communication.

### TLS Configuration

```yaml
server:
  tls:
    enabled: true
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
```

### Future Authentication

JWT-based authentication is planned for future releases.

---

## GraphQL API

### Endpoint

```
http://localhost:8080/graphql
```

### Schema Overview

The GraphQL schema provides:
- **Queries** - Fetch blocks, transactions, chain info, statistics
- **Subscriptions** - Real-time updates for blocks, transactions, sync progress
- **Pagination** - Cursor-based pagination for large datasets

### Core Types

#### Chain
```graphql
type Chain {
  chainID: String!
  chainType: ChainType!
  name: String!
  network: String!
  status: ChainStatus!
  startBlock: BigInt!
  latestIndexedBlock: BigInt!
  latestChainBlock: BigInt!
  lastUpdated: Time!
}
```

#### Block
```graphql
type Block {
  chainID: String!
  chainType: ChainType!
  number: BigInt!
  hash: String!
  parentHash: String!
  timestamp: Time!
  gasUsed: BigInt
  gasLimit: BigInt
  miner: String
  txCount: Int!
  transactions: [Transaction!]!
  createdAt: Time!
}
```

#### Transaction
```graphql
type Transaction {
  chainID: String!
  hash: String!
  blockNumber: BigInt!
  blockHash: String!
  blockTimestamp: Time!
  txIndex: Int!
  from: String!
  to: String
  value: BigInt!
  gasPrice: BigInt
  gasLimit: BigInt!
  gasUsed: BigInt
  nonce: BigInt!
  input: String
  status: TransactionStatus!
  contractAddress: String
  logs: [Log!]!
  createdAt: Time!
}
```

### Query Examples

#### Get Chain Information

```graphql
query {
  chain(chainID: "eth-mainnet") {
    chainID
    chainType
    name
    network
    status
    latestIndexedBlock
    latestChainBlock
  }
}
```

#### Get Block by Number

```graphql
query {
  block(chainID: "eth-mainnet", number: "1000000") {
    number
    hash
    timestamp
    miner
    txCount
    transactions {
      hash
      from
      to
      value
      status
    }
  }
}
```

#### Get Transaction by Hash

```graphql
query {
  transaction(chainID: "eth-mainnet", hash: "0x123...") {
    hash
    blockNumber
    from
    to
    value
    gasUsed
    status
    logs {
      address
      topics
      data
    }
  }
}
```

#### Paginated Blocks Query

```graphql
query {
  blocks(
    chainID: "eth-mainnet"
    first: 10
    orderBy: "number_desc"
  ) {
    edges {
      node {
        number
        hash
        timestamp
        txCount
      }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
      totalCount
    }
  }
}
```

#### Transactions by Address

```graphql
query {
  transactionsByAddress(
    chainID: "eth-mainnet"
    address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
    first: 20
  ) {
    edges {
      node {
        hash
        blockNumber
        from
        to
        value
        timestamp: blockTimestamp
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

#### Get Indexing Progress

```graphql
query {
  progress(chainID: "eth-mainnet") {
    chainID
    latestIndexedBlock
    latestChainBlock
    blocksBehind
    progressPercentage
    blocksPerSecond
    estimatedTimeLeft
    status
  }
}
```

#### Get Statistics

```graphql
query {
  stats(chainID: "eth-mainnet") {
    totalBlocks
    totalTransactions
    averageBlockTime
    averageTxPerBlock
  }

  globalStats {
    totalBlocks
    totalTransactions
    chainsIndexed
  }
}
```

### Subscription Examples

#### Subscribe to New Blocks

```graphql
subscription {
  blockIndexed(chainID: "eth-mainnet") {
    number
    hash
    timestamp
    txCount
  }
}
```

#### Subscribe to New Transactions

```graphql
subscription {
  transactionIndexed(chainID: "eth-mainnet") {
    hash
    from
    to
    value
    blockNumber
  }
}
```

#### Subscribe to Sync Progress

```graphql
subscription {
  syncProgress(chainID: "eth-mainnet") {
    latestIndexedBlock
    latestChainBlock
    blocksBehind
    progressPercentage
    blocksPerSecond
  }
}
```

### GraphQL Playground

Access the interactive GraphQL Playground at:
```
http://localhost:8080/graphql
```

The playground provides:
- Schema exploration
- Query autocomplete
- Documentation sidebar
- Query history
- Real-time subscription testing

---

## gRPC API

### Endpoint

```
localhost:50051
```

### Service Definition

The gRPC service is defined in `api/proto/indexer/v1/indexer.proto`:

```protobuf
service IndexerService {
  // Chain methods
  rpc GetChain(GetChainRequest) returns (Chain);
  rpc ListChains(ListChainsRequest) returns (ListChainsResponse);

  // Block methods
  rpc GetBlock(GetBlockRequest) returns (Block);
  rpc GetBlockByHash(GetBlockByHashRequest) returns (Block);
  rpc ListBlocks(ListBlocksRequest) returns (ListBlocksResponse);
  rpc GetBlockRange(GetBlockRangeRequest) returns (stream Block);
  rpc GetLatestBlock(GetLatestBlockRequest) returns (Block);

  // Transaction methods
  rpc GetTransaction(GetTransactionRequest) returns (Transaction);
  rpc ListTransactions(ListTransactionsRequest) returns (ListTransactionsResponse);
  rpc GetTransactionsByBlock(GetTransactionsByBlockRequest) returns (GetTransactionsByBlockResponse);
  rpc GetTransactionsByAddress(GetTransactionsByAddressRequest) returns (stream Transaction);

  // Progress methods
  rpc GetProgress(GetProgressRequest) returns (Progress);
  rpc ListProgress(ListProgressRequest) returns (ListProgressResponse);

  // Statistics
  rpc GetStats(GetStatsRequest) returns (Stats);
  rpc GetGlobalStats(GetGlobalStatsRequest) returns (Stats);

  // Streaming subscriptions
  rpc StreamBlocks(StreamBlocksRequest) returns (stream Block);
  rpc StreamTransactions(StreamTransactionsRequest) returns (stream Transaction);
  rpc StreamProgress(StreamProgressRequest) returns (stream Progress);
}
```

### Client Examples

#### Go Client

```go
package main

import (
    "context"
    "log"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    indexerv1 "github.com/sage-x-project/blockchain-indexer/api/proto/indexer/v1"
)

func main() {
    // Connect to server
    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := indexerv1.NewIndexerServiceClient(conn)

    // Get block
    block, err := client.GetBlock(context.Background(), &indexerv1.GetBlockRequest{
        ChainId: "eth-mainnet",
        Number:  1000000,
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Block %d: %s", block.Number, block.Hash)
}
```

#### Stream Blocks

```go
stream, err := client.StreamBlocks(context.Background(), &indexerv1.StreamBlocksRequest{
    ChainId: "eth-mainnet",
})
if err != nil {
    log.Fatal(err)
}

for {
    block, err := stream.Recv()
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("New block: %d", block.Number)
}
```

#### Python Client

```python
import grpc
from api.proto.indexer.v1 import indexer_pb2, indexer_pb2_grpc

# Connect to server
channel = grpc.insecure_channel('localhost:50051')
client = indexer_pb2_grpc.IndexerServiceStub(channel)

# Get block
request = indexer_pb2.GetBlockRequest(
    chain_id='eth-mainnet',
    number=1000000
)
block = client.GetBlock(request)

print(f"Block {block.number}: {block.hash}")
```

### TLS Configuration

For production use with TLS:

```go
creds, err := credentials.NewClientTLSFromFile("cert.pem", "")
if err != nil {
    log.Fatal(err)
}

conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
```

---

## REST API

### Base URL

```
http://localhost:8080/api
```

### Endpoints

#### Health Check

```
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-10-30T12:00:00Z"
}
```

#### List Chains

```
GET /api/chains
```

**Response:**
```json
{
  "chains": [
    {
      "chain_id": "eth-mainnet",
      "chain_type": "EVM",
      "name": "Ethereum Mainnet",
      "network": "mainnet",
      "status": "ACTIVE",
      "latest_indexed_block": 18500000,
      "latest_chain_block": 18500100
    }
  ]
}
```

#### Get Chain

```
GET /api/chains/{chainID}
```

**Example:**
```bash
curl http://localhost:8080/api/chains/eth-mainnet
```

**Response:**
```json
{
  "chain_id": "eth-mainnet",
  "chain_type": "EVM",
  "name": "Ethereum Mainnet",
  "network": "mainnet",
  "status": "ACTIVE",
  "start_block": 0,
  "latest_indexed_block": 18500000,
  "latest_chain_block": 18500100,
  "last_updated": "2025-10-30T12:00:00Z"
}
```

#### Get Block by Number

```
GET /api/chains/{chainID}/blocks/{blockNumber}
```

**Example:**
```bash
curl http://localhost:8080/api/chains/eth-mainnet/blocks/1000000
```

**Response:**
```json
{
  "chain_id": "eth-mainnet",
  "number": 1000000,
  "hash": "0x123...",
  "parent_hash": "0x456...",
  "timestamp": "2016-02-13T22:54:13Z",
  "gas_used": 21000,
  "gas_limit": 5000,
  "miner": "0x789...",
  "tx_count": 0,
  "transactions": []
}
```

#### Get Block by Hash

```
GET /api/chains/{chainID}/blocks/hash/{blockHash}
```

**Example:**
```bash
curl http://localhost:8080/api/chains/eth-mainnet/blocks/hash/0x123...
```

#### List Blocks

```
GET /api/chains/{chainID}/blocks?limit=10&offset=0&order=desc
```

**Query Parameters:**
- `limit` - Number of blocks to return (default: 20, max: 100)
- `offset` - Offset for pagination (default: 0)
- `order` - Sort order: `asc` or `desc` (default: desc)

**Example:**
```bash
curl "http://localhost:8080/api/chains/eth-mainnet/blocks?limit=5&order=desc"
```

**Response:**
```json
{
  "blocks": [
    {
      "number": 18500000,
      "hash": "0x...",
      "timestamp": "2025-10-30T12:00:00Z",
      "tx_count": 150
    }
  ],
  "total": 18500000,
  "limit": 5,
  "offset": 0
}
```

#### Get Transaction

```
GET /api/chains/{chainID}/transactions/{txHash}
```

**Example:**
```bash
curl http://localhost:8080/api/chains/eth-mainnet/transactions/0xabc...
```

**Response:**
```json
{
  "chain_id": "eth-mainnet",
  "hash": "0xabc...",
  "block_number": 1000000,
  "block_hash": "0x123...",
  "block_timestamp": "2016-02-13T22:54:13Z",
  "tx_index": 0,
  "from": "0x111...",
  "to": "0x222...",
  "value": "1000000000000000000",
  "gas_price": "20000000000",
  "gas_used": 21000,
  "nonce": 5,
  "status": "SUCCESS",
  "logs": []
}
```

#### List Transactions

```
GET /api/chains/{chainID}/transactions?limit=20&offset=0
```

**Query Parameters:**
- `limit` - Number of transactions to return (default: 20, max: 100)
- `offset` - Offset for pagination
- `block` - Filter by block number
- `from` - Filter by sender address
- `to` - Filter by recipient address

**Example:**
```bash
curl "http://localhost:8080/api/chains/eth-mainnet/transactions?from=0x123...&limit=10"
```

#### Get Transactions by Block

```
GET /api/chains/{chainID}/blocks/{blockNumber}/transactions
```

**Example:**
```bash
curl http://localhost:8080/api/chains/eth-mainnet/blocks/1000000/transactions
```

#### Get Progress

```
GET /api/chains/{chainID}/progress
```

**Example:**
```bash
curl http://localhost:8080/api/chains/eth-mainnet/progress
```

**Response:**
```json
{
  "chain_id": "eth-mainnet",
  "chain_type": "EVM",
  "latest_indexed_block": 18500000,
  "latest_chain_block": 18500100,
  "blocks_behind": 100,
  "progress_percentage": 99.9995,
  "blocks_per_second": 10.5,
  "estimated_time_left": "9s",
  "status": "SYNCING"
}
```

#### Get Statistics

```
GET /api/chains/{chainID}/stats
```

**Example:**
```bash
curl http://localhost:8080/api/chains/eth-mainnet/stats
```

**Response:**
```json
{
  "total_blocks": 18500000,
  "total_transactions": 2000000000,
  "average_block_time": 12.5,
  "average_tx_per_block": 150.2
}
```

#### Global Statistics

```
GET /api/stats
```

**Response:**
```json
{
  "total_blocks": 50000000,
  "total_transactions": 5000000000,
  "chains_indexed": 3,
  "average_block_time": 10.2,
  "average_tx_per_block": 100.5
}
```

---

## Common Patterns

### Pagination

#### GraphQL Cursor-Based Pagination

```graphql
query {
  blocks(chainID: "eth-mainnet", first: 10) {
    edges {
      node { number, hash }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}

# Next page
query {
  blocks(chainID: "eth-mainnet", first: 10, after: "cursor_value") {
    edges {
      node { number, hash }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

#### REST Offset-Based Pagination

```bash
# First page
curl "http://localhost:8080/api/chains/eth-mainnet/blocks?limit=10&offset=0"

# Second page
curl "http://localhost:8080/api/chains/eth-mainnet/blocks?limit=10&offset=10"
```

### Filtering

#### GraphQL

```graphql
query {
  transactions(
    chainID: "eth-mainnet"
    from: "0x123..."
    blockNumber: "1000000"
    first: 20
  ) {
    edges {
      node {
        hash
        value
      }
    }
  }
}
```

#### REST

```bash
curl "http://localhost:8080/api/chains/eth-mainnet/transactions?from=0x123...&block=1000000&limit=20"
```

### Real-Time Updates

#### GraphQL Subscriptions

```graphql
subscription {
  blockIndexed(chainID: "eth-mainnet") {
    number
    hash
    txCount
  }
}
```

#### gRPC Streaming

```go
stream, err := client.StreamBlocks(ctx, &indexerv1.StreamBlocksRequest{
    ChainId: "eth-mainnet",
})

for {
    block, err := stream.Recv()
    if err != nil {
        break
    }
    // Process block
}
```

---

## Error Handling

### GraphQL Errors

```json
{
  "errors": [
    {
      "message": "Block not found",
      "path": ["block"],
      "extensions": {
        "code": "NOT_FOUND",
        "chainID": "eth-mainnet",
        "blockNumber": "999999999"
      }
    }
  ],
  "data": null
}
```

### gRPC Status Codes

- `OK` - Success
- `NOT_FOUND` - Resource not found
- `INVALID_ARGUMENT` - Invalid request parameters
- `INTERNAL` - Internal server error
- `UNAVAILABLE` - Service temporarily unavailable

### REST Error Responses

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Block not found",
    "details": {
      "chain_id": "eth-mainnet",
      "block_number": 999999999
    }
  }
}
```

**HTTP Status Codes:**
- `200` - Success
- `400` - Bad Request
- `404` - Not Found
- `500` - Internal Server Error
- `503` - Service Unavailable

---

## Rate Limiting

Currently, no rate limiting is enforced. For production deployments, consider implementing rate limiting at the reverse proxy level (e.g., nginx, Envoy).

Recommended limits:
- GraphQL: 100 requests/minute per IP
- gRPC: 1000 requests/minute per connection
- REST: 100 requests/minute per IP

---

## Best Practices

1. **Use GraphQL for flexible queries** - Best for frontend applications needing specific data fields

2. **Use gRPC for high-performance backends** - Best for service-to-service communication with low latency requirements

3. **Use REST for simple integrations** - Best for quick prototypes and simple HTTP clients

4. **Implement pagination** - Always paginate large result sets to avoid memory issues

5. **Use subscriptions for real-time data** - More efficient than polling for live updates

6. **Cache responses** - Implement client-side caching for frequently accessed data

7. **Monitor API usage** - Track query patterns and optimize based on actual usage

---

## Support

For questions or issues:
- GitHub Issues: https://github.com/sage-x-project/blockchain-indexer/issues
- Documentation: See `/docs` directory

---

**Last Updated:** 2025-10-30

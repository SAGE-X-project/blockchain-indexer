# Chain Adapter Guide

> Complete guide for using and extending blockchain chain adapters

This guide covers how to use existing chain adapters and how to add support for new blockchains.

---

## Table of Contents

1. [Overview](#overview)
2. [Supported Chains](#supported-chains)
3. [Using Existing Adapters](#using-existing-adapters)
4. [Configuration](#configuration)
5. [Adding New Chain Support](#adding-new-chain-support)
6. [Adapter Registry](#adapter-registry)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

---

## Overview

Chain adapters provide a unified interface for interacting with different blockchain networks. Each adapter implements the `ChainAdapter` interface, allowing the indexer to work with any supported blockchain in a consistent way.

### Architecture

```
┌─────────────────────────────────────────────┐
│         Application Layer                   │
│    (Chain-agnostic business logic)          │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│         ChainAdapter Interface              │
│  (Unified blockchain interaction)           │
└──────────────────┬──────────────────────────┘
                   │
      ┌────────────┼────────────┐
      │            │            │
┌─────▼────┐ ┌────▼────┐ ┌────▼────┐
│   EVM    │ │ Solana  │ │ Cosmos  │
│ Adapter  │ │ Adapter │ │ Adapter │
└──────────┘ └─────────┘ └─────────┘
      │            │            │
┌─────▼────┐ ┌────▼────┐ ┌────▼────┐
│   ETH    │ │ Solana  │ │ Cosmos  │
│   RPC    │ │   RPC   │ │   RPC   │
└──────────┘ └─────────┘ └─────────┘
```

### Key Concepts

- **Chain Adapter** - Implements blockchain-specific logic
- **Normalizer** - Converts chain-specific data to domain models
- **Client** - Handles RPC communication with blockchain nodes
- **Registry** - Manages adapter creation and lifecycle

---

## Supported Chains

### EVM Chains

Supports all EVM-compatible blockchains:

- **Ethereum** - Mainnet, Sepolia, Holesky
- **BSC** - Binance Smart Chain
- **Polygon** - Polygon PoS
- **Arbitrum** - Layer 2 rollup
- **Optimism** - Layer 2 rollup
- **Avalanche C-Chain** - EVM-compatible chain

**Package:** `pkg/infrastructure/adapter/evm`

### Solana

Supports Solana blockchain:

- **Mainnet Beta**
- **Testnet**
- **Devnet**

**Package:** `pkg/infrastructure/adapter/solana`

### Cosmos

Supports Cosmos SDK and Tendermint-based chains:

- **Cosmos Hub**
- **Osmosis**
- **Juno**
- **Akash**
- **Any Cosmos SDK chain**

**Package:** `pkg/infrastructure/adapter/cosmos`

### Polkadot

Supports Substrate-based chains:

- **Polkadot**
- **Kusama**
- **Westend**
- **Moonbeam**
- **Astar**
- **Acala**

**Package:** `pkg/infrastructure/adapter/polkadot`

### Avalanche

Supports Avalanche chains:

- **C-Chain** (Contract Chain - EVM compatible)
- **X-Chain** (Exchange Chain - planned)
- **P-Chain** (Platform Chain - planned)

**Package:** `pkg/infrastructure/adapter/avalanche`

### Ripple

Supports XRP Ledger:

- **Mainnet**
- **Testnet**
- **Devnet**

**Package:** `pkg/infrastructure/adapter/ripple`

---

## Using Existing Adapters

### Basic Usage

#### 1. Using the Adapter Registry

```go
package main

import (
    "context"
    "log"

    "github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
    "github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter"
    "github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/evm"
)

func main() {
    // Create adapter config
    config := evm.DefaultConfig()
    config.ChainID = "eth-mainnet"
    config.ChainName = "Ethereum Mainnet"
    config.Network = "mainnet"
    config.RPCEndpoints = []string{"https://eth.llamarpc.com"}

    // Create adapter using registry
    adapter, err := adapter.CreateAdapter(models.ChainTypeEVM, config)
    if err != nil {
        log.Fatal(err)
    }

    // Connect to blockchain
    ctx := context.Background()
    if err := adapter.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer adapter.Disconnect()

    // Get latest block number
    latestBlock, err := adapter.GetLatestBlockNumber(ctx)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Latest block: %d", latestBlock)
}
```

#### 2. Direct Adapter Creation

```go
package main

import (
    "context"
    "log"

    "github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/solana"
)

func main() {
    // Create Solana config
    config := solana.DefaultConfig(
        "solana-mainnet",
        "mainnet-beta",
        "https://api.mainnet-beta.solana.com",
    )

    // Create adapter directly
    adapter, err := solana.NewAdapter(config)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := adapter.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer adapter.Disconnect()

    // Get latest slot
    latestSlot, err := adapter.GetLatestBlockNumber(ctx)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Latest slot: %d", latestSlot)
}
```

### Common Operations

#### Get Block by Number

```go
block, err := adapter.GetBlockByNumber(ctx, 1000000)
if err != nil {
    log.Fatal(err)
}

log.Printf("Block %d: %s", block.Number, block.Hash)
log.Printf("Transactions: %d", len(block.Transactions))
```

#### Get Block by Hash

```go
block, err := adapter.GetBlockByHash(ctx, "0x123...")
if err != nil {
    log.Fatal(err)
}

log.Printf("Block number: %d", block.Number)
```

#### Get Transaction

```go
tx, err := adapter.GetTransaction(ctx, "0xabc...")
if err != nil {
    log.Fatal(err)
}

log.Printf("From: %s, To: %s, Value: %s", tx.From, tx.To, tx.Value)
```

#### Get Block Range

```go
blocks, err := adapter.GetBlocks(ctx, 1000000, 1000010)
if err != nil {
    log.Fatal(err)
}

for _, block := range blocks {
    log.Printf("Block %d: %d txs", block.Number, len(block.Transactions))
}
```

#### Subscribe to New Blocks

```go
subscription, err := adapter.SubscribeNewBlocks(ctx)
if err != nil {
    log.Fatal(err)
}
defer subscription.Unsubscribe()

for {
    select {
    case block := <-subscription.Blocks():
        log.Printf("New block: %d", block.Number)
    case err := <-subscription.Err():
        log.Printf("Subscription error: %v", err)
        return
    }
}
```

#### Health Check

```go
if adapter.IsHealthy(ctx) {
    log.Println("Adapter is healthy")
} else {
    log.Println("Adapter is unhealthy")
}
```

---

## Configuration

### EVM Configuration

```go
config := &evm.Config{
    ChainID:            "eth-mainnet",
    ChainName:          "Ethereum Mainnet",
    Network:            "mainnet",
    RPCEndpoints:       []string{
        "https://eth.llamarpc.com",
        "https://rpc.ankr.com/eth",
    },
    WSEndpoints:        []string{"wss://eth.llamarpc.com"},

    // Connection settings
    MaxConnections:     10,
    ConnectionTimeout:  30 * time.Second,
    RequestTimeout:     10 * time.Second,

    // Retry settings
    MaxRetries:         3,
    RetryDelay:         1 * time.Second,
    RetryBackoff:       2.0,

    // Block fetching
    BatchSize:          100,
    ConcurrentFetches:  10,
    BlockConfirmations: 12,
    EnableReceiptFetch: true,

    // WebSocket
    EnableWebSocket:    true,
}
```

### Solana Configuration

```go
config := &solana.Config{
    ChainID:            "solana-mainnet",
    ChainName:          "Solana",
    Network:            "mainnet-beta",
    RPCEndpoint:        "https://api.mainnet-beta.solana.com",
    WSEndpoint:         "wss://api.mainnet-beta.solana.com",

    RequestTimeout:     30 * time.Second,
    MaxRetries:         3,
    RetryDelay:         2 * time.Second,
    MaxConcurrentReqs:  10,

    ConcurrentFetches:  5,
    MaxBlockRange:      100,
    EnableWebSocket:    true,
    EnableVotes:        false,
    EnableRewards:      true,
    TransactionDetails: "full",
}
```

### Cosmos Configuration

```go
config := &cosmos.Config{
    ChainID:                    "cosmoshub-4",
    ChainName:                  "Cosmos Hub",
    Network:                    "mainnet",
    RPCURL:                     "https://rpc.cosmos.network:443",
    RESTURL:                    "https://api.cosmos.network",

    Timeout:                    30 * time.Second,
    RetryAttempts:              3,
    RetryDelay:                 1 * time.Second,
    MaxConnections:             10,
    BatchSize:                  100,

    Bech32Prefix:               "cosmos",
    CoinDenom:                  "uatom",
    IncludeTxEvents:            true,
    IncludeBeginBlockEvents:    false,
    IncludeEndBlockEvents:      false,
}
```

### Configuration from YAML

```yaml
# config/config.yaml
chains:
  - chain_id: "eth-mainnet"
    chain_name: "Ethereum Mainnet"
    network: "mainnet"
    chain_type: "evm"

    rpc_endpoints:
      - "https://eth.llamarpc.com"

    timeout: 30s
    retry_attempts: 3
    batch_size: 100
```

Load configuration:

```go
import (
    "github.com/sage-x-project/blockchain-indexer/internal/config"
)

cfg, err := config.Load("config/config.yaml")
if err != nil {
    log.Fatal(err)
}

// Access chain configs
for _, chainCfg := range cfg.Chains {
    log.Printf("Chain: %s (%s)", chainCfg.ChainName, chainCfg.ChainType)
}
```

---

## Adding New Chain Support

### Step 1: Implement the ChainAdapter Interface

```go
package mychain

import (
    "context"

    "github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
    "github.com/sage-x-project/blockchain-indexer/pkg/domain/service"
)

type Adapter struct {
    config    *Config
    client    *Client
    normalizer *Normalizer
    connected bool
}

func NewAdapter(config *Config) (*Adapter, error) {
    if err := config.Validate(); err != nil {
        return nil, err
    }

    client := NewClient(config)
    normalizer := NewNormalizer(config)

    return &Adapter{
        config:     config,
        client:     client,
        normalizer: normalizer,
    }, nil
}

// Implement all ChainAdapter interface methods
func (a *Adapter) GetChainType() models.ChainType {
    return models.ChainType("mychain")
}

func (a *Adapter) GetChainID() string {
    return a.config.ChainID
}

func (a *Adapter) GetChainInfo() *models.ChainInfo {
    return &models.ChainInfo{
        ChainType: a.GetChainType(),
        ChainID:   a.config.ChainID,
        Name:      a.config.ChainName,
        Network:   a.config.Network,
    }
}

func (a *Adapter) Connect(ctx context.Context) error {
    // Implement connection logic
    a.connected = true
    return nil
}

func (a *Adapter) Disconnect() error {
    // Implement disconnection logic
    a.connected = false
    return nil
}

func (a *Adapter) IsConnected() bool {
    return a.connected
}

func (a *Adapter) IsHealthy(ctx context.Context) bool {
    // Implement health check
    return a.connected
}

func (a *Adapter) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
    // Implement latest block fetching
    return 0, nil
}

func (a *Adapter) GetBlockByNumber(ctx context.Context, number uint64) (*models.Block, error) {
    // Implement block fetching
    return nil, nil
}

func (a *Adapter) GetBlockByHash(ctx context.Context, hash string) (*models.Block, error) {
    // Implement block by hash fetching
    return nil, nil
}

func (a *Adapter) GetBlocks(ctx context.Context, start, end uint64) ([]*models.Block, error) {
    // Implement batch block fetching
    return nil, nil
}

func (a *Adapter) GetTransaction(ctx context.Context, hash string) (*models.Transaction, error) {
    // Implement transaction fetching
    return nil, nil
}

func (a *Adapter) GetTransactionsByBlock(ctx context.Context, blockNumber uint64) ([]*models.Transaction, error) {
    // Implement transactions by block
    return nil, nil
}

func (a *Adapter) SubscribeNewBlocks(ctx context.Context) (service.BlockSubscription, error) {
    // Implement block subscription
    return nil, nil
}

func (a *Adapter) SubscribeNewTransactions(ctx context.Context) (service.TransactionSubscription, error) {
    // Implement transaction subscription
    return nil, nil
}
```

### Step 2: Create Configuration

```go
package mychain

import (
    "fmt"
    "time"
)

type Config struct {
    ChainID        string
    ChainName      string
    Network        string
    RPCURL         string
    Timeout        time.Duration
    RetryAttempts  int
    MaxConnections int
    BatchSize      int
}

func DefaultConfig() *Config {
    return &Config{
        ChainID:        "mychain-mainnet",
        ChainName:      "My Chain",
        Network:        "mainnet",
        RPCURL:         "https://rpc.mychain.com",
        Timeout:        30 * time.Second,
        RetryAttempts:  3,
        MaxConnections: 10,
        BatchSize:      100,
    }
}

func (c *Config) Validate() error {
    if c.ChainID == "" {
        return fmt.Errorf("chain_id is required")
    }
    if c.RPCURL == "" {
        return fmt.Errorf("rpc_url is required")
    }
    if c.Timeout <= 0 {
        return fmt.Errorf("timeout must be positive")
    }
    return nil
}
```

### Step 3: Implement RPC Client

```go
package mychain

import (
    "context"
    "net/http"
)

type Client struct {
    config     *Config
    httpClient *http.Client
}

func NewClient(config *Config) *Client {
    return &Client{
        config: config,
        httpClient: &http.Client{
            Timeout: config.Timeout,
        },
    }
}

func (c *Client) GetLatestBlock(ctx context.Context) (*Block, error) {
    // Implement RPC call
    return nil, nil
}

func (c *Client) GetBlockByNumber(ctx context.Context, number uint64) (*Block, error) {
    // Implement RPC call
    return nil, nil
}

// Add more RPC methods as needed
```

### Step 4: Implement Data Normalizer

```go
package mychain

import (
    "github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

type Normalizer struct {
    config *Config
}

func NewNormalizer(config *Config) *Normalizer {
    return &Normalizer{config: config}
}

func (n *Normalizer) NormalizeBlock(block *Block) (*models.Block, error) {
    // Convert chain-specific block to domain model
    domainBlock := &models.Block{
        ChainID:    n.config.ChainID,
        ChainType:  models.ChainType("mychain"),
        Number:     block.Number,
        Hash:       block.Hash,
        ParentHash: block.ParentHash,
        Timestamp:  block.Timestamp,
        // ... map other fields
    }

    // Normalize transactions
    for _, tx := range block.Transactions {
        domainTx, err := n.NormalizeTransaction(tx, block)
        if err != nil {
            return nil, err
        }
        domainBlock.Transactions = append(domainBlock.Transactions, domainTx)
    }

    return domainBlock, nil
}

func (n *Normalizer) NormalizeTransaction(tx *Transaction, block *Block) (*models.Transaction, error) {
    // Convert chain-specific transaction to domain model
    return &models.Transaction{
        ChainID:        n.config.ChainID,
        Hash:           tx.Hash,
        BlockNumber:    block.Number,
        BlockHash:      block.Hash,
        BlockTimestamp: block.Timestamp,
        // ... map other fields
    }, nil
}
```

### Step 5: Register with Registry

```go
package main

import (
    "github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
    "github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter"
    "your-module/pkg/infrastructure/adapter/mychain"
)

func init() {
    // Register custom chain adapter
    adapter.GlobalRegistry().Register(
        models.ChainType("mychain"),
        func(config interface{}) (service.ChainAdapter, error) {
            mychainConfig, ok := config.(*mychain.Config)
            if !ok {
                return nil, fmt.Errorf("invalid config type for mychain adapter")
            }
            return mychain.NewAdapter(mychainConfig)
        },
    )
}
```

### Step 6: Add Tests

```go
package mychain

import (
    "testing"
)

func TestDefaultConfig(t *testing.T) {
    config := DefaultConfig()

    if config.ChainID != "mychain-mainnet" {
        t.Errorf("expected chain_id to be 'mychain-mainnet', got '%s'", config.ChainID)
    }

    if err := config.Validate(); err != nil {
        t.Errorf("default config should be valid: %v", err)
    }
}

func TestNewAdapter(t *testing.T) {
    config := DefaultConfig()

    adapter, err := NewAdapter(config)
    if err != nil {
        t.Fatalf("NewAdapter() error = %v", err)
    }

    if adapter.GetChainType() != models.ChainType("mychain") {
        t.Errorf("GetChainType() = %v, want mychain", adapter.GetChainType())
    }
}
```

---

## Adapter Registry

The adapter registry provides centralized management of all chain adapters.

### Using the Registry

```go
import (
    "github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter"
)

// Get global registry instance
registry := adapter.GlobalRegistry()

// Check if chain is supported
if adapter.IsChainSupported(models.ChainTypeEVM) {
    log.Println("EVM is supported")
}

// List all supported chains
chains := adapter.SupportedChains()
for _, chainType := range chains {
    log.Printf("Supported: %s", chainType)
}

// Create adapter
config := evm.DefaultConfig()
adapter, err := registry.Create(models.ChainTypeEVM, config)
```

### Custom Registry

```go
// Create custom registry
registry := adapter.NewRegistry()

// Register custom adapter
registry.Register(models.ChainType("custom"), func(config interface{}) (service.ChainAdapter, error) {
    // Custom factory logic
    return customAdapter, nil
})

// Unregister adapter
registry.Unregister(models.ChainType("custom"))
```

---

## Best Practices

### 1. Configuration Validation

Always validate configuration before creating adapters:

```go
if err := config.Validate(); err != nil {
    return nil, fmt.Errorf("invalid config: %w", err)
}
```

### 2. Error Handling

Implement comprehensive error handling:

```go
block, err := adapter.GetBlockByNumber(ctx, number)
if err != nil {
    if errors.Is(err, ErrBlockNotFound) {
        // Handle not found
        return nil, nil
    }
    return nil, fmt.Errorf("failed to get block: %w", err)
}
```

### 3. Context Usage

Always respect context cancellation:

```go
func (a *Adapter) GetBlockByNumber(ctx context.Context, number uint64) (*models.Block, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Proceed with operation
}
```

### 4. Connection Management

Properly manage connections:

```go
if err := adapter.Connect(ctx); err != nil {
    return err
}
defer adapter.Disconnect()

// Use adapter
```

### 5. Retry Logic

Implement exponential backoff for retries:

```go
for attempt := 0; attempt < maxRetries; attempt++ {
    if attempt > 0 {
        time.Sleep(retryDelay * time.Duration(1<<uint(attempt)))
    }

    result, err := operation()
    if err == nil {
        return result, nil
    }
}
```

### 6. Resource Cleanup

Always clean up resources:

```go
subscription, err := adapter.SubscribeNewBlocks(ctx)
if err != nil {
    return err
}
defer subscription.Unsubscribe()
```

### 7. Health Monitoring

Regularly check adapter health:

```go
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        if !adapter.IsHealthy(ctx) {
            log.Println("Adapter unhealthy, reconnecting...")
            adapter.Disconnect()
            adapter.Connect(ctx)
        }
    }
}
```

---

## Troubleshooting

### Connection Issues

**Problem:** Adapter fails to connect

**Solutions:**
- Check RPC endpoint URL
- Verify network connectivity
- Check firewall rules
- Increase connection timeout

```go
config.ConnectionTimeout = 60 * time.Second
```

### Rate Limiting

**Problem:** RPC requests are rate limited

**Solutions:**
- Add multiple RPC endpoints
- Reduce concurrent fetches
- Implement request queuing

```go
config.RPCEndpoints = []string{
    "https://primary-rpc.com",
    "https://backup-rpc.com",
}
config.ConcurrentFetches = 5
```

### Block Not Found

**Problem:** Blocks return not found errors

**Solutions:**
- Check block confirmations setting
- Verify block number is valid
- Check if chain is fully synced

```go
config.BlockConfirmations = 12
```

### Memory Issues

**Problem:** High memory usage

**Solutions:**
- Reduce batch size
- Reduce concurrent fetches
- Enable caching with limits

```go
config.BatchSize = 50
config.ConcurrentFetches = 5
config.CacheMaxSize = 1000
```

### Subscription Drops

**Problem:** WebSocket subscriptions disconnect

**Solutions:**
- Implement reconnection logic
- Increase ping interval
- Check WebSocket endpoint stability

```go
config.WSReconnectDelay = 5 * time.Second
config.WSPingInterval = 30 * time.Second
```

---

## Support

For additional help:
- Check example configs in `config/` directory
- Review existing adapter implementations
- Submit issues on GitHub
- See main documentation in `/docs`

---

**Last Updated:** 2025-10-30

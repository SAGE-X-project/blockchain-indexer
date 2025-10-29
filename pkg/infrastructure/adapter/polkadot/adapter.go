package polkadot

import (
	"context"
	"fmt"
	"sync"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/service"
)

// Ensure Adapter implements ChainAdapter interface
var _ service.ChainAdapter = (*Adapter)(nil)

// Adapter implements the ChainAdapter interface for Polkadot/Substrate chains
type Adapter struct {
	config     *Config
	client     *Client
	normalizer *Normalizer
	chainInfo  *models.ChainInfo
	mu         sync.RWMutex
	connected  bool
}

// NewAdapter creates a new Polkadot chain adapter
func NewAdapter(config *Config) (*Adapter, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client, err := NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	normalizer := NewNormalizer(config.ChainID, config.Network)

	adapter := &Adapter{
		config:     config,
		client:     client,
		normalizer: normalizer,
		chainInfo:  normalizer.NormalizeChainInfo(),
		connected:  false,
	}

	// Verify connection
	if err := adapter.verifyConnection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to verify connection: %w", err)
	}

	adapter.connected = true

	return adapter, nil
}

// verifyConnection verifies the connection to the Polkadot node
func (a *Adapter) verifyConnection(ctx context.Context) error {
	// Try to get system health to verify connection
	_, err := a.client.GetSystemHealth(ctx)
	if err != nil {
		return fmt.Errorf("failed to get system health: %w", err)
	}

	return nil
}

// GetChainType returns the chain type
func (a *Adapter) GetChainType() models.ChainType {
	return models.ChainTypePolkadot
}

// GetChainID returns the chain ID
func (a *Adapter) GetChainID() string {
	return a.config.ChainID
}

// GetChainInfo returns chain information
func (a *Adapter) GetChainInfo() *models.ChainInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.chainInfo
}

// GetLatestBlockNumber returns the latest finalized block number
func (a *Adapter) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	blockNumber, err := a.client.GetLatestBlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block number: %w", err)
	}

	return blockNumber, nil
}

// GetBlockByNumber fetches a block by number
func (a *Adapter) GetBlockByNumber(ctx context.Context, number uint64) (*models.Block, error) {
	// Fetch block
	signedBlock, err := a.client.GetBlockByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %d: %w", number, err)
	}

	// Get block hash
	blockHash, err := a.client.GetBlockHash(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get block hash for block %d: %w", number, err)
	}

	// Normalize block
	domainBlock, err := a.normalizer.NormalizeBlock(signedBlock, blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize block %d: %w", number, err)
	}

	return domainBlock, nil
}

// GetBlockByHash fetches a block by hash
func (a *Adapter) GetBlockByHash(ctx context.Context, hash string) (*models.Block, error) {
	// Parse hash
	blockHash, err := parseHash(hash)
	if err != nil {
		return nil, fmt.Errorf("invalid hash: %w", err)
	}

	// Fetch block
	signedBlock, err := a.client.GetBlock(ctx, blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash %s: %w", hash, err)
	}

	// Normalize block
	domainBlock, err := a.normalizer.NormalizeBlock(signedBlock, blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize block %s: %w", hash, err)
	}

	return domainBlock, nil
}

// GetBlocks fetches multiple blocks in a range
func (a *Adapter) GetBlocks(ctx context.Context, start, end uint64) ([]*models.Block, error) {
	blocks := make([]*models.Block, 0, end-start+1)

	for i := start; i <= end; i++ {
		block, err := a.GetBlockByNumber(ctx, i)
		if err != nil {
			return nil, fmt.Errorf("failed to get block %d: %w", i, err)
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

// GetTransaction fetches a transaction by hash
func (a *Adapter) GetTransaction(ctx context.Context, hash string) (*models.Transaction, error) {
	// Polkadot doesn't have a direct transaction-by-hash lookup
	// This would require maintaining an index or scanning blocks
	return nil, fmt.Errorf("GetTransaction not implemented for Polkadot - extrinsics must be retrieved from blocks")
}

// GetTransactionsByBlock fetches all transactions in a block
func (a *Adapter) GetTransactionsByBlock(ctx context.Context, blockNumber uint64) ([]*models.Transaction, error) {
	block, err := a.GetBlockByNumber(ctx, blockNumber)
	if err != nil {
		return nil, err
	}

	return block.Transactions, nil
}

// IsHealthy checks if the adapter is healthy
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	a.mu.RLock()
	connected := a.connected
	a.mu.RUnlock()

	if !connected {
		return false
	}

	// Check connection health
	health, err := a.client.GetSystemHealth(ctx)
	if err != nil {
		a.mu.Lock()
		a.connected = false
		a.mu.Unlock()
		return false
	}

	// Consider healthy if we have peers and are not syncing
	return health.Peers > 0 && !health.IsSyncing
}

// Connect connects to the Polkadot node
func (a *Adapter) Connect(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.connected {
		return nil
	}

	// Verify connection
	if err := a.verifyConnection(ctx); err != nil {
		return fmt.Errorf("failed to verify connection: %w", err)
	}

	a.connected = true
	return nil
}

// Disconnect closes the connection
func (a *Adapter) Disconnect() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.connected {
		return nil
	}

	if err := a.client.Close(); err != nil {
		return fmt.Errorf("failed to close client: %w", err)
	}

	a.connected = false
	return nil
}

// SubscribeNewBlocks subscribes to new blocks
func (a *Adapter) SubscribeNewBlocks(ctx context.Context) (service.BlockSubscription, error) {
	// WebSocket subscriptions not yet fully implemented
	return nil, fmt.Errorf("block subscriptions not yet implemented for Polkadot")
}

// SubscribeNewTransactions subscribes to new transactions
func (a *Adapter) SubscribeNewTransactions(ctx context.Context) (service.TransactionSubscription, error) {
	// Polkadot doesn't have direct transaction subscriptions
	return nil, fmt.Errorf("transaction subscriptions not supported for Polkadot")
}

// GetConfig returns the adapter configuration
func (a *Adapter) GetConfig() *Config {
	return a.config
}

// IsConnected returns whether the adapter is connected
func (a *Adapter) IsConnected() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.connected
}

// parseHash parses a hex hash string to types.Hash
func parseHash(hash string) (types.Hash, error) {
	var result types.Hash

	// Remove "0x" prefix if present
	if len(hash) >= 2 && hash[:2] == "0x" {
		hash = hash[2:]
	}

	// Decode hex string
	if len(hash) != 64 {
		return result, fmt.Errorf("invalid hash length: expected 64 characters, got %d", len(hash))
	}

	bytes := make([]byte, 32)
	for i := 0; i < 32; i++ {
		var b byte
		_, err := fmt.Sscanf(hash[i*2:i*2+2], "%02x", &b)
		if err != nil {
			return result, fmt.Errorf("failed to parse hash: %w", err)
		}
		bytes[i] = b
	}

	copy(result[:], bytes)
	return result, nil
}

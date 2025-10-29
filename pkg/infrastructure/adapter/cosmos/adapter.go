package cosmos

import (
	"context"
	"fmt"
	"sync"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/service"
)

// Ensure Adapter implements ChainAdapter interface
var _ service.ChainAdapter = (*Adapter)(nil)

// Adapter implements the ChainAdapter interface for Cosmos/Tendermint chains
type Adapter struct {
	config     *Config
	client     *Client
	normalizer *Normalizer
	chainInfo  *models.ChainInfo
	mu         sync.RWMutex
	connected  bool
}

// NewAdapter creates a new Cosmos chain adapter
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

	// Start the client
	if err := client.Start(); err != nil {
		return nil, fmt.Errorf("failed to start client: %w", err)
	}

	// Verify connection
	if err := adapter.verifyConnection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to verify connection: %w", err)
	}

	adapter.connected = true

	return adapter, nil
}

// verifyConnection verifies the connection to the Cosmos node
func (a *Adapter) verifyConnection(ctx context.Context) error {
	// Try to get status to verify connection
	_, err := a.client.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	return nil
}

// GetChainType returns the chain type
func (a *Adapter) GetChainType() models.ChainType {
	return models.ChainTypeCosmos
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

// GetLatestBlockNumber returns the latest block number (height)
func (a *Adapter) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	height, err := a.client.GetLatestBlockHeight(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block height: %w", err)
	}

	return uint64(height), nil
}

// GetBlockByNumber fetches a block by number (height)
func (a *Adapter) GetBlockByNumber(ctx context.Context, number uint64) (*models.Block, error) {
	height := int64(number)

	// Fetch block
	block, err := a.client.GetBlockByHeight(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %d: %w", number, err)
	}

	// Fetch block results for transaction details
	blockResults, err := a.client.BlockResults(ctx, &height)
	if err != nil {
		// Continue without block results
		blockResults = nil
	}

	// Normalize block
	domainBlock, err := a.normalizer.NormalizeBlock(block, blockResults)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize block %d: %w", number, err)
	}

	return domainBlock, nil
}

// GetBlockByHash fetches a block by hash
func (a *Adapter) GetBlockByHash(ctx context.Context, hash string) (*models.Block, error) {
	// Decode hash
	hashBytes, err := decodeHash(hash)
	if err != nil {
		return nil, fmt.Errorf("invalid hash: %w", err)
	}

	// Fetch block by hash
	result, err := a.client.BlockByHash(ctx, hashBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash %s: %w", hash, err)
	}

	block := result.Block
	height := block.Height

	// Fetch block results
	blockResults, err := a.client.BlockResults(ctx, &height)
	if err != nil {
		blockResults = nil
	}

	// Normalize block
	domainBlock, err := a.normalizer.NormalizeBlock(block, blockResults)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize block %s: %w", hash, err)
	}

	return domainBlock, nil
}

// GetTransaction fetches a transaction by hash
func (a *Adapter) GetTransaction(ctx context.Context, hash string) (*models.Transaction, error) {
	// Decode hash
	hashBytes, err := decodeHash(hash)
	if err != nil {
		return nil, fmt.Errorf("invalid hash: %w", err)
	}

	// Fetch transaction
	result, err := a.client.Tx(ctx, hashBytes, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction %s: %w", hash, err)
	}

	// Fetch block results for transaction details
	height := result.Height
	blockResults, err := a.client.BlockResults(ctx, &height)
	if err != nil {
		blockResults = nil
	}

	// Normalize transaction
	tx, err := a.normalizer.NormalizeTransaction(result.Tx, result.Height, result.Index, blockResults)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize transaction %s: %w", hash, err)
	}

	return tx, nil
}

// GetTransactionsByBlock fetches all transactions in a block
func (a *Adapter) GetTransactionsByBlock(ctx context.Context, number uint64) ([]*models.Transaction, error) {
	block, err := a.GetBlockByNumber(ctx, number)
	if err != nil {
		return nil, err
	}

	return block.Transactions, nil
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

// IsHealthy checks if the adapter is healthy
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	err := a.HealthCheck(ctx)
	return err == nil
}

// Connect connects to the Cosmos node
func (a *Adapter) Connect(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.connected {
		return nil
	}

	// Start the client if not already started
	if err := a.client.Start(); err != nil {
		return fmt.Errorf("failed to start client: %w", err)
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
	return a.Close()
}

// SubscribeNewBlocks subscribes to new blocks
func (a *Adapter) SubscribeNewBlocks(ctx context.Context) (service.BlockSubscription, error) {
	if !a.config.EnableWebSocket {
		return nil, fmt.Errorf("websocket is not enabled")
	}

	return &blockSubscription{
		adapter: a,
		ctx:     ctx,
		blockCh: make(chan *models.Block, 100),
		errCh:   make(chan error, 1),
	}, nil
}

// SubscribeNewTransactions subscribes to new transactions
func (a *Adapter) SubscribeNewTransactions(ctx context.Context) (service.TransactionSubscription, error) {
	// Cosmos doesn't have direct transaction subscriptions, return nil
	return nil, fmt.Errorf("transaction subscriptions not supported for Cosmos")
}

// HealthCheck checks if the adapter is healthy
func (a *Adapter) HealthCheck(ctx context.Context) error {
	a.mu.RLock()
	connected := a.connected
	a.mu.RUnlock()

	if !connected {
		return fmt.Errorf("adapter is not connected")
	}

	// Check connection health
	_, err := a.client.Health(ctx)
	if err != nil {
		a.mu.Lock()
		a.connected = false
		a.mu.Unlock()
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// Close closes the adapter and releases resources
func (a *Adapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.connected {
		return nil
	}

	if err := a.client.Stop(); err != nil {
		return fmt.Errorf("failed to stop client: %w", err)
	}

	a.connected = false
	return nil
}

// Subscribe subscribes to new blocks (if WebSocket is enabled)
func (a *Adapter) Subscribe(ctx context.Context, subscriber string) (<-chan *models.Block, error) {
	if !a.config.EnableWebSocket {
		return nil, fmt.Errorf("websocket is not enabled")
	}

	// Subscribe to new blocks
	eventCh, err := a.client.Subscribe(ctx, subscriber, "tm.event='NewBlock'")
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	// Create output channel
	blockCh := make(chan *models.Block, 100)

	// Start goroutine to process events
	go func() {
		defer close(blockCh)
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-eventCh:
				if !ok {
					return
				}

				// Extract block from event
				if event.Data != nil {
					// Try to process the event data
					// This is a simplified version - actual implementation may vary
					// based on the event structure
					continue
				}
			}
		}
	}()

	return blockCh, nil
}

// Unsubscribe unsubscribes from block events
func (a *Adapter) Unsubscribe(ctx context.Context, subscriber string) error {
	if !a.config.EnableWebSocket {
		return nil
	}

	return a.client.UnsubscribeAll(ctx, subscriber)
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

// blockSubscription implements service.BlockSubscription
type blockSubscription struct {
	adapter *Adapter
	ctx     context.Context
	blockCh chan *models.Block
	errCh   chan error
}

// Channel returns the channel that receives new blocks
func (s *blockSubscription) Channel() <-chan *models.Block {
	return s.blockCh
}

// Unsubscribe cancels the subscription
func (s *blockSubscription) Unsubscribe() {
	close(s.blockCh)
	close(s.errCh)
}

// Err returns any subscription error
func (s *blockSubscription) Err() <-chan error {
	return s.errCh
}

// decodeHash decodes a hex hash string to bytes
func decodeHash(hash string) ([]byte, error) {
	// Remove "0x" prefix if present
	if len(hash) >= 2 && hash[:2] == "0x" {
		hash = hash[2:]
	}

	// Decode hex string
	hashBytes := make([]byte, len(hash)/2)
	for i := 0; i < len(hashBytes); i++ {
		var b byte
		_, err := fmt.Sscanf(hash[i*2:i*2+2], "%02x", &b)
		if err != nil {
			return nil, err
		}
		hashBytes[i] = b
	}

	return hashBytes, nil
}

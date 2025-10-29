package solana

import (
	"context"
	"fmt"
	"sync"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/service"
)

// Adapter implements the ChainAdapter interface for Solana
type Adapter struct {
	config     *Config
	client     *Client
	normalizer *Normalizer
	chainInfo  *models.ChainInfo
	mu         sync.RWMutex
	connected  bool
}

// NewAdapter creates a new Solana chain adapter
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
		connected:  true,
	}

	// Verify connection
	if err := adapter.verifyConnection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to verify connection: %w", err)
	}

	return adapter, nil
}

// verifyConnection verifies the connection to the Solana node
func (a *Adapter) verifyConnection(ctx context.Context) error {
	// Try to get version to verify connection
	_, err := a.client.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	return nil
}

// GetChainType returns the chain type
func (a *Adapter) GetChainType() models.ChainType {
	return models.ChainTypeSolana
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

// GetLatestBlockNumber returns the latest slot (block) number
func (a *Adapter) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	slot, err := a.client.GetSlot(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest slot: %w", err)
	}

	return slot, nil
}

// GetBlockByNumber fetches a block by slot number
func (a *Adapter) GetBlockByNumber(ctx context.Context, number uint64) (*models.Block, error) {
	// Fetch block from Solana
	block, err := a.client.GetBlock(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %d: %w", number, err)
	}

	// Normalize block
	domainBlock, err := a.normalizer.NormalizeBlock(number, block)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize block %d: %w", number, err)
	}

	return domainBlock, nil
}

// GetBlockByHash fetches a block by hash (blockhash in Solana)
// Note: Solana doesn't support direct blockhash lookup, this is less efficient
func (a *Adapter) GetBlockByHash(ctx context.Context, hash string) (*models.Block, error) {
	// Solana doesn't have a direct blockhash lookup API
	// We would need to scan blocks or maintain an index
	// For now, return an error
	return nil, fmt.Errorf("GetBlockByHash not efficiently supported in Solana - use GetBlockByNumber instead")
}

// GetBlocks fetches multiple blocks in a range
func (a *Adapter) GetBlocks(ctx context.Context, start, end uint64) ([]*models.Block, error) {
	if start > end {
		return nil, fmt.Errorf("invalid range: start %d > end %d", start, end)
	}

	// Limit range to MaxBlockRange
	if end-start > a.config.MaxBlockRange {
		end = start + a.config.MaxBlockRange
	}

	// First, get the list of available blocks in the range
	// Solana may have skipped slots, so we need to query which slots have blocks
	availableSlots, err := a.client.GetBlocksInRange(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks in range: %w", err)
	}

	blocks := make([]*models.Block, 0, len(availableSlots))

	// Fetch blocks concurrently
	blocksChan := make(chan *models.Block, a.config.ConcurrentFetches)
	errorsChan := make(chan error, a.config.ConcurrentFetches)
	semaphore := make(chan struct{}, a.config.ConcurrentFetches)
	var wg sync.WaitGroup

	for _, slot := range availableSlots {
		wg.Add(1)
		go func(slotNum uint64) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			block, err := a.GetBlockByNumber(ctx, slotNum)
			if err != nil {
				errorsChan <- fmt.Errorf("failed to fetch block %d: %w", slotNum, err)
				return
			}

			blocksChan <- block
		}(slot)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(blocksChan)
		close(errorsChan)
	}()

	// Collect blocks
	blockMap := make(map[uint64]*models.Block)
	for block := range blocksChan {
		blockMap[block.Number] = block
	}

	// Check for errors
	for err := range errorsChan {
		if err != nil {
			return nil, err
		}
	}

	// Sort blocks by slot number
	for _, slot := range availableSlots {
		if block, ok := blockMap[slot]; ok {
			blocks = append(blocks, block)
		}
	}

	return blocks, nil
}

// GetTransaction fetches a transaction by signature (hash)
func (a *Adapter) GetTransaction(ctx context.Context, hash string) (*models.Transaction, error) {
	// Fetch transaction from Solana
	tx, err := a.client.GetTransaction(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction %s: %w", hash, err)
	}

	// Get block to extract blockhash and timestamp
	var blockhash string
	var blockTime int64
	if tx.BlockTime != nil {
		blockTime = *tx.BlockTime
	}

	// Parse transaction
	var txData TransactionWithMeta
	if tx.Transaction != nil {
		txData.Transaction = tx.Transaction
		txData.Meta = tx.Meta
		txData.Version = tx.Version
	}

	// Normalize transaction
	domainTx, err := a.normalizer.NormalizeTransaction(
		tx.Slot,
		blockhash,
		models.TimeFromUnix(blockTime),
		0, // Index not available in single transaction query
		&txData,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize transaction %s: %w", hash, err)
	}

	return domainTx, nil
}

// GetTransactionsByBlock fetches all transactions in a block
func (a *Adapter) GetTransactionsByBlock(ctx context.Context, blockNumber uint64) ([]*models.Transaction, error) {
	block, err := a.GetBlockByNumber(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %d: %w", blockNumber, err)
	}

	return block.Transactions, nil
}

// IsHealthy checks if the adapter is healthy
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	status, err := a.client.GetHealthStatus(ctx)
	if err != nil {
		return false
	}

	// Check if connected and error count is below threshold
	return status.Connected && status.ErrorCount < int32(a.config.MaxErrorCount)
}

// Connect establishes connection to the chain
func (a *Adapter) Connect(ctx context.Context) error {
	if a.connected {
		return nil
	}

	if err := a.verifyConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	a.mu.Lock()
	a.connected = true
	a.mu.Unlock()

	return nil
}

// Disconnect closes the connection to the chain
func (a *Adapter) Disconnect() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.connected {
		return nil
	}

	if err := a.client.Close(); err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	a.connected = false
	return nil
}

// SubscribeNewBlocks subscribes to new blocks
// Note: This requires WebSocket support which is not implemented in this basic version
func (a *Adapter) SubscribeNewBlocks(ctx context.Context) (service.BlockSubscription, error) {
	if !a.config.EnableWebSocket {
		return nil, fmt.Errorf("websocket not enabled")
	}

	// WebSocket subscription would be implemented here
	// For now, return not implemented error
	return nil, fmt.Errorf("websocket subscription not yet implemented for Solana")
}

// SubscribeNewTransactions subscribes to new transactions
// Note: This requires WebSocket support which is not implemented in this basic version
func (a *Adapter) SubscribeNewTransactions(ctx context.Context) (service.TransactionSubscription, error) {
	if !a.config.EnableWebSocket {
		return nil, fmt.Errorf("websocket not enabled")
	}

	// WebSocket subscription would be implemented here
	// For now, return not implemented error
	return nil, fmt.Errorf("websocket subscription not yet implemented for Solana")
}

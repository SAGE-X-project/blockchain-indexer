package evm

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/service"
)

// Adapter implements the ChainAdapter interface for EVM-compatible chains
type Adapter struct {
	config     *Config
	client     *Client
	normalizer *Normalizer
	chainInfo  *models.ChainInfo
	mu         sync.RWMutex
	connected  bool
}

// NewAdapter creates a new EVM chain adapter
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
		connected:  true,
	}

	// Initialize chain info
	if err := adapter.initChainInfo(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize chain info: %w", err)
	}

	return adapter, nil
}

// initChainInfo initializes chain information
func (a *Adapter) initChainInfo(ctx context.Context) error {
	// Verify connection by getting chain ID
	_, err := a.client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	a.chainInfo = &models.ChainInfo{
		ChainType: models.ChainTypeEVM,
		ChainID:   a.config.ChainID,
		Name:      a.config.ChainName,
		Network:   a.config.Network,
	}

	return nil
}

// GetChainType returns the chain type
func (a *Adapter) GetChainType() models.ChainType {
	return models.ChainTypeEVM
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

// GetLatestBlockNumber returns the latest block number
func (a *Adapter) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	blockNumber, err := a.client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block number: %w", err)
	}

	return blockNumber, nil
}

// GetBlockByNumber fetches a block by number
func (a *Adapter) GetBlockByNumber(ctx context.Context, number uint64) (*models.Block, error) {
	// Fetch block
	block, err := a.client.BlockByNumber(ctx, big.NewInt(int64(number)))
	if err != nil {
		return nil, fmt.Errorf("failed to get block %d: %w", number, err)
	}

	// Fetch receipts if enabled
	var receipts []*types.Receipt
	if a.config.EnableReceiptFetch {
		receipts, err = a.fetchReceipts(ctx, block)
		if err != nil {
			// Log warning but don't fail - receipts are optional
			// In production, you'd use a proper logger here
			_ = err
		}
	}

	// Normalize block
	domainBlock, err := a.normalizer.NormalizeBlock(block, receipts)
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
		return nil, fmt.Errorf("invalid block hash: %w", err)
	}

	// Fetch block
	block, err := a.client.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %s: %w", hash, err)
	}

	// Fetch receipts if enabled
	var receipts []*types.Receipt
	if a.config.EnableReceiptFetch {
		receipts, err = a.fetchReceipts(ctx, block)
		if err != nil {
			// Log warning but don't fail
			_ = err
		}
	}

	// Normalize block
	domainBlock, err := a.normalizer.NormalizeBlock(block, receipts)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize block %s: %w", hash, err)
	}

	return domainBlock, nil
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

	blocks := make([]*models.Block, 0, end-start+1)

	// Fetch blocks concurrently
	blocksChan := make(chan *models.Block, a.config.ConcurrentFetches)
	errorsChan := make(chan error, a.config.ConcurrentFetches)
	semaphore := make(chan struct{}, a.config.ConcurrentFetches)
	var wg sync.WaitGroup

	for number := start; number <= end; number++ {
		wg.Add(1)
		go func(num uint64) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			block, err := a.GetBlockByNumber(ctx, num)
			if err != nil {
				errorsChan <- fmt.Errorf("failed to fetch block %d: %w", num, err)
				return
			}

			blocksChan <- block
		}(number)
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

	// Sort blocks by height
	for number := start; number <= end; number++ {
		if block, ok := blockMap[number]; ok {
			blocks = append(blocks, block)
		}
	}

	return blocks, nil
}

// GetTransaction fetches a transaction by hash
func (a *Adapter) GetTransaction(ctx context.Context, hash string) (*models.Transaction, error) {
	// Parse hash
	txHash, err := parseHash(hash)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction hash: %w", err)
	}

	// Fetch transaction
	tx, isPending, err := a.client.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction %s: %w", hash, err)
	}

	if isPending {
		return nil, fmt.Errorf("transaction %s is pending", hash)
	}

	// Fetch receipt
	receipt, err := a.client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt for transaction %s: %w", hash, err)
	}

	// Fetch block to get timestamp
	block, err := a.client.BlockByHash(ctx, receipt.BlockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block for transaction %s: %w", hash, err)
	}

	// Normalize transaction
	domainTx, err := a.normalizer.NormalizeTransaction(tx, block, receipt, uint64(receipt.TransactionIndex))
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
	status, err := a.client.HealthStatus(ctx)
	if err != nil {
		return false
	}

	return status.Connected && status.ErrorCount < a.config.MaxErrorCount
}

// Connect establishes connection to the chain
func (a *Adapter) Connect(ctx context.Context) error {
	if a.connected {
		return nil
	}

	if err := a.initChainInfo(ctx); err != nil {
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
func (a *Adapter) SubscribeNewBlocks(ctx context.Context) (service.BlockSubscription, error) {
	if !a.config.EnableWebSocket {
		return nil, fmt.Errorf("websocket not enabled")
	}

	headerChan := make(chan *types.Header, a.config.SubscriptionBufferSize)
	sub, err := a.client.SubscribeNewHead(ctx, headerChan)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to new blocks: %w", err)
	}

	blockChan := make(chan *models.Block, a.config.SubscriptionBufferSize)
	errChan := make(chan error, 1)

	// Start goroutine to convert headers to blocks
	go func() {
		defer close(blockChan)
		defer close(errChan)

		for {
			select {
			case <-ctx.Done():
				return
			case err := <-sub.Err():
				if err != nil {
					errChan <- err
				}
				return
			case header := <-headerChan:
				if header == nil {
					return
				}

				// Fetch full block
				block, err := a.GetBlockByNumber(ctx, header.Number.Uint64())
				if err != nil {
					errChan <- fmt.Errorf("failed to fetch block %d: %w", header.Number.Uint64(), err)
					continue
				}

				blockChan <- block
			}
		}
	}()

	return &blockSubscription{
		sub:       sub,
		blockChan: blockChan,
		errChan:   errChan,
	}, nil
}

// SubscribeNewTransactions subscribes to new transactions
func (a *Adapter) SubscribeNewTransactions(ctx context.Context) (service.TransactionSubscription, error) {
	// EVM doesn't have a built-in pending transaction subscription that works reliably
	// This would require monitoring the mempool or watching new blocks
	return nil, fmt.Errorf("transaction subscription not implemented")
}

// fetchReceipts fetches receipts for all transactions in a block
func (a *Adapter) fetchReceipts(ctx context.Context, block *types.Block) ([]*types.Receipt, error) {
	receipts := make([]*types.Receipt, 0, len(block.Transactions()))

	for _, tx := range block.Transactions() {
		receipt, err := a.client.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch receipt for tx %s: %w", tx.Hash().Hex(), err)
		}
		receipts = append(receipts, receipt)
	}

	return receipts, nil
}

// Helper functions

func parseHash(hash string) ([32]byte, error) {
	var result [32]byte

	// Remove 0x prefix if present
	if len(hash) >= 2 && hash[0] == '0' && (hash[1] == 'x' || hash[1] == 'X') {
		hash = hash[2:]
	}

	if len(hash) != 64 {
		return result, fmt.Errorf("invalid hash length: expected 64, got %d", len(hash))
	}

	for i := 0; i < 32; i++ {
		high := hexToByte(hash[i*2])
		low := hexToByte(hash[i*2+1])
		if high == 0xff || low == 0xff {
			return result, fmt.Errorf("invalid hex character in hash")
		}
		result[i] = (high << 4) | low
	}

	return result, nil
}

func hexToByte(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	default:
		return 0xff
	}
}

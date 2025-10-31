package avalanche

import (
	"context"
	"fmt"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/service"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/evm"
)

// Ensure Adapter implements ChainAdapter interface
var _ service.ChainAdapter = (*Adapter)(nil)

// Adapter implements the ChainAdapter interface for Avalanche chains
// Currently supports C-Chain (EVM compatible) by wrapping EVM adapter
type Adapter struct {
	config     *Config
	evmAdapter service.ChainAdapter
	chainInfo  *models.ChainInfo
}

// NewAdapter creates a new Avalanche chain adapter
func NewAdapter(config *Config) (*Adapter, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// For now, only C-Chain is supported (EVM compatible)
	if !config.IsCChain() {
		return nil, fmt.Errorf("only C-Chain is currently supported")
	}

	// Create EVM adapter config for C-Chain
	evmConfig := &evm.Config{
		ChainID:            config.ChainID,
		ChainName:          config.ChainName,
		Network:            config.Network,
		RPCEndpoints:       []string{config.RPCURL},
		WSEndpoints:        []string{config.WSURL},
		MaxConnections:     config.MaxConnections,
		ConnectionTimeout:  config.Timeout,
		RequestTimeout:     config.Timeout,
		MaxRetries:         config.RetryAttempts,
		RetryDelay:         config.RetryDelay,
		BatchSize:          config.BatchSize,
		EnableWebSocket:    config.EnableWebSocket,
		RequestsPerSecond:  100,
		BurstSize:          10,
		ConcurrentFetches:  5,
		BlockConfirmations: 0,
	}

	// Create EVM adapter
	evmAdapter, err := evm.NewAdapter(evmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create EVM adapter for C-Chain: %w", err)
	}

	chainInfo := &models.ChainInfo{
		ChainType: models.ChainTypeAvalanche,
		ChainID:   config.ChainID,
		Name:      config.ChainName,
		Network:   config.Network,
	}

	return &Adapter{
		config:     config,
		evmAdapter: evmAdapter,
		chainInfo:  chainInfo,
	}, nil
}

// GetChainType returns the chain type
func (a *Adapter) GetChainType() models.ChainType {
	return models.ChainTypeAvalanche
}

// GetChainID returns the chain ID
func (a *Adapter) GetChainID() string {
	return a.config.ChainID
}

// GetChainInfo returns chain information
func (a *Adapter) GetChainInfo() *models.ChainInfo {
	return a.chainInfo
}

// GetLatestBlockNumber returns the latest block number
func (a *Adapter) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	return a.evmAdapter.GetLatestBlockNumber(ctx)
}

// GetBlockByNumber fetches a block by number
func (a *Adapter) GetBlockByNumber(ctx context.Context, number uint64) (*models.Block, error) {
	block, err := a.evmAdapter.GetBlockByNumber(ctx, number)
	if err != nil {
		return nil, err
	}

	// Update chain ID to Avalanche
	block.ChainID = a.config.ChainID
	return block, nil
}

// GetBlockByHash fetches a block by hash
func (a *Adapter) GetBlockByHash(ctx context.Context, hash string) (*models.Block, error) {
	block, err := a.evmAdapter.GetBlockByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	// Update chain ID to Avalanche
	block.ChainID = a.config.ChainID
	return block, nil
}

// GetBlocks fetches multiple blocks in a range
func (a *Adapter) GetBlocks(ctx context.Context, start, end uint64) ([]*models.Block, error) {
	blocks, err := a.evmAdapter.GetBlocks(ctx, start, end)
	if err != nil {
		return nil, err
	}

	// Update chain ID to Avalanche for all blocks
	for _, block := range blocks {
		block.ChainID = a.config.ChainID
	}

	return blocks, nil
}

// GetTransaction fetches a transaction by hash
func (a *Adapter) GetTransaction(ctx context.Context, hash string) (*models.Transaction, error) {
	tx, err := a.evmAdapter.GetTransaction(ctx, hash)
	if err != nil {
		return nil, err
	}

	// Update chain ID to Avalanche
	tx.ChainID = a.config.ChainID
	return tx, nil
}

// GetTransactionsByBlock fetches all transactions in a block
func (a *Adapter) GetTransactionsByBlock(ctx context.Context, blockNumber uint64) ([]*models.Transaction, error) {
	txs, err := a.evmAdapter.GetTransactionsByBlock(ctx, blockNumber)
	if err != nil {
		return nil, err
	}

	// Update chain ID to Avalanche for all transactions
	for _, tx := range txs {
		tx.ChainID = a.config.ChainID
	}

	return txs, nil
}

// IsHealthy checks if the adapter is healthy
func (a *Adapter) IsHealthy(ctx context.Context) bool {
	return a.evmAdapter.IsHealthy(ctx)
}

// Connect connects to the Avalanche node
func (a *Adapter) Connect(ctx context.Context) error {
	return a.evmAdapter.Connect(ctx)
}

// Disconnect closes the connection
func (a *Adapter) Disconnect() error {
	return a.evmAdapter.Disconnect()
}

// SubscribeNewBlocks subscribes to new blocks
func (a *Adapter) SubscribeNewBlocks(ctx context.Context) (service.BlockSubscription, error) {
	return a.evmAdapter.SubscribeNewBlocks(ctx)
}

// SubscribeNewTransactions subscribes to new transactions
func (a *Adapter) SubscribeNewTransactions(ctx context.Context) (service.TransactionSubscription, error) {
	return a.evmAdapter.SubscribeNewTransactions(ctx)
}

// GetConfig returns the adapter configuration
func (a *Adapter) GetConfig() *Config {
	return a.config
}

package service

import (
	"context"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// ChainAdapter defines the interface for interacting with different blockchain types
// Following the Adapter Pattern and Dependency Inversion Principle
// Each blockchain implementation (EVM, Solana, Cosmos, etc.) must implement this interface
type ChainAdapter interface {
	// Chain information
	GetChainType() models.ChainType
	GetChainID() string
	GetChainInfo() *models.ChainInfo

	// Block operations
	GetLatestBlockNumber(ctx context.Context) (uint64, error)
	GetBlockByNumber(ctx context.Context, number uint64) (*models.Block, error)
	GetBlockByHash(ctx context.Context, hash string) (*models.Block, error)
	GetBlocks(ctx context.Context, start, end uint64) ([]*models.Block, error)

	// Transaction operations
	GetTransaction(ctx context.Context, hash string) (*models.Transaction, error)
	GetTransactionsByBlock(ctx context.Context, blockNumber uint64) ([]*models.Transaction, error)

	// Health and connection
	IsHealthy(ctx context.Context) bool
	Connect(ctx context.Context) error
	Disconnect() error

	// Subscription support (optional, can return nil if not supported)
	SubscribeNewBlocks(ctx context.Context) (BlockSubscription, error)
	SubscribeNewTransactions(ctx context.Context) (TransactionSubscription, error)
}

// BlockSubscription represents a subscription to new blocks
type BlockSubscription interface {
	// Channel returns the channel that receives new blocks
	Channel() <-chan *models.Block

	// Unsubscribe cancels the subscription
	Unsubscribe()

	// Err returns any subscription error
	Err() <-chan error
}

// TransactionSubscription represents a subscription to new transactions
type TransactionSubscription interface {
	// Channel returns the channel that receives new transactions
	Channel() <-chan *models.Transaction

	// Unsubscribe cancels the subscription
	Unsubscribe()

	// Err returns any subscription error
	Err() <-chan error
}

// ChainAdapterConfig holds configuration for a chain adapter
type ChainAdapterConfig struct {
	ChainType   models.ChainType
	ChainID     string
	RPCEndpoint string
	WSEndpoint  string
	Timeout     int // Timeout in seconds
	MaxRetries  int
	Config      map[string]interface{} // Chain-specific config
}

// Validate validates the adapter configuration
func (c *ChainAdapterConfig) Validate() error {
	if !c.ChainType.IsValid() {
		return models.ErrInvalidChainType
	}
	if c.ChainID == "" {
		return models.ErrInvalidChainID
	}
	if c.RPCEndpoint == "" {
		return ErrInvalidRPCEndpoint
	}
	if c.Timeout <= 0 {
		c.Timeout = 30 // Default timeout
	}
	if c.MaxRetries <= 0 {
		c.MaxRetries = 3 // Default retries
	}
	return nil
}

// ChainAdapterFactory is a function that creates a ChainAdapter
type ChainAdapterFactory func(config *ChainAdapterConfig) (ChainAdapter, error)

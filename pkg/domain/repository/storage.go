package repository

import (
	"context"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// Storage provides the main storage interface
// Following the Repository Pattern and Interface Segregation Principle
// Combines BlockRepository and TransactionRepository
type Storage interface {
	BlockRepository
	TransactionRepository
	ChainRepository

	// Lifecycle methods
	Close() error

	// Transaction support
	NewBatch() Batch

	// Statistics
	GetStats(ctx context.Context) (*StorageStats, error)
}

// Batch provides atomic batch write operations
// Following the Interface Segregation Principle
type Batch interface {
	// Block operations
	SetBlock(ctx context.Context, block *models.Block) error
	SetBlocks(ctx context.Context, blocks []*models.Block) error

	// Transaction operations
	SetTransaction(ctx context.Context, tx *models.Transaction) error
	SetTransactions(ctx context.Context, txs []*models.Transaction) error

	// Commit writes all batched operations atomically
	Commit() error

	// Reset clears all operations in the batch
	Reset()

	// Count returns the number of operations in the batch
	Count() int

	// Close releases batch resources without committing
	Close() error
}

// StorageStats holds storage statistics
type StorageStats struct {
	TotalBlocks       uint64            `json:"total_blocks"`
	TotalTransactions uint64            `json:"total_transactions"`
	DiskUsage         uint64            `json:"disk_usage"`        // bytes
	ChainStats        map[string]*ChainStorageStats `json:"chain_stats"` // stats per chain
}

// ChainStorageStats holds storage statistics for a specific chain
type ChainStorageStats struct {
	ChainID          string `json:"chain_id"`
	BlockCount       uint64 `json:"block_count"`
	TransactionCount uint64 `json:"transaction_count"`
	LatestBlock      uint64 `json:"latest_block"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	Type     string                 // Storage type: "pebble", "postgres", etc.
	Path     string                 // Path for embedded databases
	Host     string                 // Host for remote databases
	Port     int                    // Port for remote databases
	Database string                 // Database name
	Username string                 // Username for authentication
	Password string                 // Password for authentication
	Options  map[string]interface{} // Additional storage-specific options
}

// Validate validates the storage configuration
func (c *StorageConfig) Validate() error {
	if c.Type == "" {
		return ErrInvalidStorageType
	}
	// Add more validation based on storage type
	return nil
}

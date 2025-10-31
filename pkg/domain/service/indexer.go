package service

import (
	"context"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// Indexer defines the interface for blockchain indexing service
// Following the Single Responsibility Principle and Dependency Inversion Principle
type Indexer interface {
	// Lifecycle methods
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// Indexing operations
	IndexBlock(ctx context.Context, chainID string, blockNumber uint64) error
	IndexBlockRange(ctx context.Context, chainID string, start, end uint64) error
	ReindexBlock(ctx context.Context, chainID string, blockNumber uint64) error

	// Status and monitoring
	GetStatus(chainID string) (*IndexerStatus, error)
	GetAllStatuses() ([]*IndexerStatus, error)

	// Chain management
	AddChain(ctx context.Context, chain *models.Chain) error
	RemoveChain(ctx context.Context, chainID string) error
	PauseChain(ctx context.Context, chainID string) error
	ResumeChain(ctx context.Context, chainID string) error
}

// IndexerStatus represents the current status of an indexer
type IndexerStatus struct {
	ChainID            string             `json:"chain_id"`
	ChainType          models.ChainType   `json:"chain_type"`
	Status             models.ChainStatus `json:"status"`
	LatestIndexedBlock uint64             `json:"latest_indexed_block"`
	LatestChainBlock   uint64             `json:"latest_chain_block"`
	BlocksBehind       uint64             `json:"blocks_behind"`
	SyncProgress       float64            `json:"sync_progress"` // 0-100%
	BlocksPerSecond    float64            `json:"blocks_per_second"`
	EstimatedTimeLeft  int64              `json:"estimated_time_left"` // seconds
	LastError          string             `json:"last_error,omitempty"`
}

// IndexerConfig holds configuration for the indexer
type IndexerConfig struct {
	// Number of worker goroutines per chain
	WorkersPerChain int

	// Batch size for block fetching
	BatchSize int

	// Maximum number of retry attempts
	MaxRetries int

	// Number of confirmation blocks to wait
	ConfirmationBlocks uint64

	// Enable gap detection and recovery
	EnableGapRecovery bool

	// Enable real-time indexing
	EnableRealtime bool
}

// Validate validates the indexer configuration
func (c *IndexerConfig) Validate() error {
	if c.WorkersPerChain <= 0 {
		c.WorkersPerChain = 10 // Default
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 100 // Default
	}
	if c.MaxRetries <= 0 {
		c.MaxRetries = 3 // Default
	}
	return nil
}

// DefaultIndexerConfig returns a default indexer configuration
func DefaultIndexerConfig() *IndexerConfig {
	return &IndexerConfig{
		WorkersPerChain:    10,
		BatchSize:          100,
		MaxRetries:         3,
		ConfirmationBlocks: 0,
		EnableGapRecovery:  true,
		EnableRealtime:     true,
	}
}

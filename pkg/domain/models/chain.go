package models

import (
	"errors"
	"time"
)

// Chain represents a blockchain configuration
type Chain struct {
	// Chain identification
	ChainType ChainType `json:"chain_type"`
	ChainID   string    `json:"chain_id"`
	Name      string    `json:"name"`
	Network   string    `json:"network"` // mainnet, testnet, devnet, etc.

	// RPC configuration
	RPCEndpoints []string `json:"rpc_endpoints"`
	WSEndpoints  []string `json:"ws_endpoints,omitempty"`

	// Indexing configuration
	StartBlock      uint64 `json:"start_block"`       // Block to start indexing from
	Enabled         bool   `json:"enabled"`           // Whether indexing is enabled
	BatchSize       int    `json:"batch_size"`        // Number of blocks to fetch in parallel
	Workers         int    `json:"workers"`           // Number of worker goroutines
	ConfirmationBlocks uint64 `json:"confirmation_blocks"` // Number of blocks to wait for confirmation

	// Chain-specific configuration
	Config map[string]interface{} `json:"config,omitempty"`

	// Status information
	LatestIndexedBlock uint64    `json:"latest_indexed_block"`
	LatestChainBlock   uint64    `json:"latest_chain_block"`
	LastUpdated        time.Time `json:"last_updated"`
	Status             ChainStatus `json:"status"`
}

// ChainStatus represents the status of chain indexing
type ChainStatus string

const (
	// ChainStatusIdle indicates the chain is not being indexed
	ChainStatusIdle ChainStatus = "idle"

	// ChainStatusSyncing indicates the chain is syncing
	ChainStatusSyncing ChainStatus = "syncing"

	// ChainStatusLive indicates the chain is fully synced and live
	ChainStatusLive ChainStatus = "live"

	// ChainStatusError indicates an error occurred
	ChainStatusError ChainStatus = "error"

	// ChainStatusPaused indicates indexing is paused
	ChainStatusPaused ChainStatus = "paused"
)

// String returns the string representation of ChainStatus
func (s ChainStatus) String() string {
	return string(s)
}

// NewChain creates a new Chain configuration
func NewChain(chainType ChainType, chainID, name string) *Chain {
	return &Chain{
		ChainType:          chainType,
		ChainID:            chainID,
		Name:               name,
		Enabled:            true,
		BatchSize:          100,
		Workers:            10,
		ConfirmationBlocks: 0,
		RPCEndpoints:       make([]string, 0),
		WSEndpoints:        make([]string, 0),
		Config:             make(map[string]interface{}),
		Status:             ChainStatusIdle,
		LastUpdated:        time.Now(),
	}
}

// Validate validates the chain configuration
func (c *Chain) Validate() error {
	if !c.ChainType.IsValid() {
		return ErrInvalidChainType
	}
	if c.ChainID == "" {
		return ErrInvalidChainID
	}
	if c.Name == "" {
		return errors.New("chain name is required")
	}
	if len(c.RPCEndpoints) == 0 {
		return errors.New("at least one RPC endpoint is required")
	}
	if c.BatchSize <= 0 {
		return errors.New("batch size must be positive")
	}
	if c.Workers <= 0 {
		return errors.New("workers must be positive")
	}
	return nil
}

// IsSynced returns true if the chain is fully synced
func (c *Chain) IsSynced() bool {
	return c.Status == ChainStatusLive
}

// GetSyncProgress returns the sync progress as a percentage (0-100)
func (c *Chain) GetSyncProgress() float64 {
	if c.LatestChainBlock == 0 {
		return 0
	}
	return float64(c.LatestIndexedBlock) / float64(c.LatestChainBlock) * 100
}

// GetBlocksBehind returns the number of blocks behind the chain head
func (c *Chain) GetBlocksBehind() uint64 {
	if c.LatestChainBlock > c.LatestIndexedBlock {
		return c.LatestChainBlock - c.LatestIndexedBlock
	}
	return 0
}

// UpdateStatus updates the chain status and last updated time
func (c *Chain) UpdateStatus(status ChainStatus) {
	c.Status = status
	c.LastUpdated = time.Now()
}

// UpdateLatestBlock updates the latest indexed and chain blocks
func (c *Chain) UpdateLatestBlock(indexedBlock, chainBlock uint64) {
	c.LatestIndexedBlock = indexedBlock
	c.LatestChainBlock = chainBlock
	c.LastUpdated = time.Now()
}

// GetConfig retrieves a configuration value by key
func (c *Chain) GetConfig(key string) (interface{}, bool) {
	if c.Config == nil {
		return nil, false
	}
	value, exists := c.Config[key]
	return value, exists
}

// SetConfig sets a configuration value
func (c *Chain) SetConfig(key string, value interface{}) {
	if c.Config == nil {
		c.Config = make(map[string]interface{})
	}
	c.Config[key] = value
}

// ChainStats represents statistics for a chain
type ChainStats struct {
	ChainID            string      `json:"chain_id"`
	ChainType          ChainType   `json:"chain_type"`
	LatestIndexedBlock uint64      `json:"latest_indexed_block"`
	LatestChainBlock   uint64      `json:"latest_chain_block"`
	BlocksBehind       uint64      `json:"blocks_behind"`
	SyncProgress       float64     `json:"sync_progress"`
	Status             ChainStatus `json:"status"`
	TotalBlocks        uint64      `json:"total_blocks"`
	TotalTransactions  uint64      `json:"total_transactions"`
	LastUpdated        time.Time   `json:"last_updated"`
}

// GetStats returns statistics for the chain
func (c *Chain) GetStats() *ChainStats {
	return &ChainStats{
		ChainID:            c.ChainID,
		ChainType:          c.ChainType,
		LatestIndexedBlock: c.LatestIndexedBlock,
		LatestChainBlock:   c.LatestChainBlock,
		BlocksBehind:       c.GetBlocksBehind(),
		SyncProgress:       c.GetSyncProgress(),
		Status:             c.Status,
		LastUpdated:        c.LastUpdated,
	}
}

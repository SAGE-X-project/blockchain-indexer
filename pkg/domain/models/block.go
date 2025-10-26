package models

import (
	"encoding/json"
	"time"
)

// Block represents a normalized blockchain block
// This model is chain-agnostic and contains common fields across all chains
type Block struct {
	// Chain information
	ChainType ChainType `json:"chain_type"`
	ChainID   string    `json:"chain_id"`

	// Block identification
	Number     uint64 `json:"number"`      // Block height/number
	Hash       string `json:"hash"`        // Block hash
	ParentHash string `json:"parent_hash"` // Parent block hash

	// Timestamps
	Timestamp *Timestamp `json:"timestamp"` // Block timestamp with optional slot/epoch

	// Block producer/validator
	Proposer string `json:"proposer"` // Miner, validator, or block producer address

	// Transaction information
	TxCount      int      `json:"tx_count"`      // Number of transactions
	TxHashes     []string `json:"tx_hashes"`     // Transaction hashes
	Transactions []*Transaction `json:"transactions,omitempty"` // Full transactions (optional)

	// Block size and limits
	Size     uint64 `json:"size"`      // Block size in bytes
	GasUsed  uint64 `json:"gas_used"`  // Gas used (EVM chains)
	GasLimit uint64 `json:"gas_limit"` // Gas limit (EVM chains)

	// State information
	StateRoot        string `json:"state_root,omitempty"`         // State root hash
	TransactionsRoot string `json:"transactions_root,omitempty"`  // Transactions root hash
	ReceiptsRoot     string `json:"receipts_root,omitempty"`      // Receipts root hash

	// Chain-specific metadata
	// This field holds additional chain-specific data that doesn't fit in common fields
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Indexing metadata
	IndexedAt time.Time `json:"indexed_at"` // When this block was indexed
}

// NewBlock creates a new Block with basic information
func NewBlock(chainType ChainType, chainID string, number uint64, hash string) *Block {
	return &Block{
		ChainType: chainType,
		ChainID:   chainID,
		Number:    number,
		Hash:      hash,
		Timestamp: NewTimestamp(time.Now().Unix()),
		TxHashes:  make([]string, 0),
		Metadata:  make(map[string]interface{}),
		IndexedAt: time.Now(),
	}
}

// Validate validates the block data
func (b *Block) Validate() error {
	if !b.ChainType.IsValid() {
		return ErrInvalidChainType
	}
	if b.ChainID == "" {
		return ErrInvalidChainID
	}
	if b.Hash == "" {
		return ErrInvalidBlockHash
	}
	if b.Timestamp == nil {
		return ErrInvalidTimestamp
	}
	return nil
}

// GetMetadata retrieves a metadata value by key
func (b *Block) GetMetadata(key string) (interface{}, bool) {
	if b.Metadata == nil {
		return nil, false
	}
	value, exists := b.Metadata[key]
	return value, exists
}

// SetMetadata sets a metadata value
func (b *Block) SetMetadata(key string, value interface{}) {
	if b.Metadata == nil {
		b.Metadata = make(map[string]interface{})
	}
	b.Metadata[key] = value
}

// MarshalJSON implements json.Marshaler interface
func (b *Block) MarshalJSON() ([]byte, error) {
	type Alias Block
	return json.Marshal(&struct {
		*Alias
		ChainType string `json:"chain_type"`
	}{
		Alias:     (*Alias)(b),
		ChainType: b.ChainType.String(),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface
func (b *Block) UnmarshalJSON(data []byte) error {
	type Alias Block
	aux := &struct {
		*Alias
		ChainType string `json:"chain_type"`
	}{
		Alias: (*Alias)(b),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert string chain type back to ChainType
	if aux.ChainType != "" {
		b.ChainType = ChainType(aux.ChainType)
	}

	return nil
}

// BlockSummary represents a lightweight block summary
type BlockSummary struct {
	ChainType  ChainType  `json:"chain_type"`
	ChainID    string     `json:"chain_id"`
	Number     uint64     `json:"number"`
	Hash       string     `json:"hash"`
	ParentHash string     `json:"parent_hash"`
	Timestamp  *Timestamp `json:"timestamp"`
	Proposer   string     `json:"proposer"`
	TxCount    int        `json:"tx_count"`
}

// ToSummary converts a Block to a BlockSummary
func (b *Block) ToSummary() *BlockSummary {
	return &BlockSummary{
		ChainType:  b.ChainType,
		ChainID:    b.ChainID,
		Number:     b.Number,
		Hash:       b.Hash,
		ParentHash: b.ParentHash,
		Timestamp:  b.Timestamp,
		Proposer:   b.Proposer,
		TxCount:    b.TxCount,
	}
}

// BlockFilter represents filtering criteria for block queries
type BlockFilter struct {
	ChainType  *ChainType `json:"chain_type,omitempty"`
	ChainID    *string    `json:"chain_id,omitempty"`
	NumberMin  *uint64    `json:"number_min,omitempty"`
	NumberMax  *uint64    `json:"number_max,omitempty"`
	Proposer   *string    `json:"proposer,omitempty"`
	TimeMin    *time.Time `json:"time_min,omitempty"`
	TimeMax    *time.Time `json:"time_max,omitempty"`
	TxCountMin *int       `json:"tx_count_min,omitempty"`
	TxCountMax *int       `json:"tx_count_max,omitempty"`
}

// PaginationOptions represents pagination parameters
type PaginationOptions struct {
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
	Cursor *string `json:"cursor,omitempty"`
}

// DefaultPaginationOptions returns default pagination options
func DefaultPaginationOptions() *PaginationOptions {
	return &PaginationOptions{
		Limit:  20,
		Offset: 0,
	}
}

// Validate validates pagination options
func (p *PaginationOptions) Validate() error {
	if p.Limit <= 0 {
		return ErrInvalidLimit
	}
	if p.Limit > 1000 {
		return ErrLimitTooLarge
	}
	if p.Offset < 0 {
		return ErrInvalidOffset
	}
	return nil
}

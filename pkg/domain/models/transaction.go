package models

import (
	"encoding/json"
	"time"
)

// Transaction represents a normalized blockchain transaction
// This model is chain-agnostic and contains common fields across all chains
type Transaction struct {
	// Chain information
	ChainType ChainType `json:"chain_type"`
	ChainID   string    `json:"chain_id"`

	// Transaction identification
	Hash  string `json:"hash"`  // Transaction hash
	Index uint64 `json:"index"` // Transaction index within block

	// Block information
	BlockNumber uint64 `json:"block_number"` // Block number/height
	BlockHash   string `json:"block_hash"`   // Block hash

	// Transaction participants
	From string `json:"from"`          // Sender address
	To   string `json:"to,omitempty"`  // Receiver address (empty for contract creation)

	// Value and fees
	Value    string `json:"value"`     // Transfer amount (as string to handle big numbers)
	Fee      string `json:"fee"`       // Transaction fee
	GasUsed  uint64 `json:"gas_used"`  // Gas used (EVM chains)
	GasPrice string `json:"gas_price"` // Gas price (EVM chains)

	// Transaction status
	Status TxStatus `json:"status"` // Success, failed, or pending

	// Transaction data
	Input      []byte `json:"input,omitempty"`       // Transaction input data
	Output     []byte `json:"output,omitempty"`      // Transaction output data
	Nonce      uint64 `json:"nonce"`                 // Transaction nonce
	Type       uint8  `json:"type"`                  // Transaction type
	Signature  []byte `json:"signature,omitempty"`   // Transaction signature

	// Contract information (for smart contract transactions)
	ContractAddress string   `json:"contract_address,omitempty"` // Created contract address
	Logs            []*Log   `json:"logs,omitempty"`             // Transaction logs/events

	// Timestamps
	Timestamp *Timestamp `json:"timestamp"` // Transaction timestamp

	// Chain-specific metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Indexing metadata
	IndexedAt time.Time `json:"indexed_at"` // When this transaction was indexed
}

// NewTransaction creates a new Transaction with basic information
func NewTransaction(chainType ChainType, chainID string, hash string) *Transaction {
	return &Transaction{
		ChainType: chainType,
		ChainID:   chainID,
		Hash:      hash,
		Timestamp: NewTimestamp(time.Now().Unix()),
		Metadata:  make(map[string]interface{}),
		IndexedAt: time.Now(),
		Logs:      make([]*Log, 0),
	}
}

// Validate validates the transaction data
func (t *Transaction) Validate() error {
	if !t.ChainType.IsValid() {
		return ErrInvalidChainType
	}
	if t.ChainID == "" {
		return ErrInvalidChainID
	}
	if t.Hash == "" {
		return ErrInvalidTxHash
	}
	if t.From == "" {
		return ErrInvalidFromAddress
	}
	if t.Timestamp == nil {
		return ErrInvalidTimestamp
	}
	return nil
}

// IsContractCreation returns true if this is a contract creation transaction
func (t *Transaction) IsContractCreation() bool {
	return t.To == "" && t.ContractAddress != ""
}

// IsSuccess returns true if the transaction succeeded
func (t *Transaction) IsSuccess() bool {
	return t.Status == TxStatusSuccess
}

// GetMetadata retrieves a metadata value by key
func (t *Transaction) GetMetadata(key string) (interface{}, bool) {
	if t.Metadata == nil {
		return nil, false
	}
	value, exists := t.Metadata[key]
	return value, exists
}

// SetMetadata sets a metadata value
func (t *Transaction) SetMetadata(key string, value interface{}) {
	if t.Metadata == nil {
		t.Metadata = make(map[string]interface{})
	}
	t.Metadata[key] = value
}

// MarshalJSON implements json.Marshaler interface
func (t *Transaction) MarshalJSON() ([]byte, error) {
	type Alias Transaction
	return json.Marshal(&struct {
		*Alias
		ChainType string `json:"chain_type"`
		Status    string `json:"status"`
	}{
		Alias:     (*Alias)(t),
		ChainType: t.ChainType.String(),
		Status:    t.Status.String(),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface
func (t *Transaction) UnmarshalJSON(data []byte) error {
	type Alias Transaction
	aux := &struct {
		*Alias
		ChainType string `json:"chain_type"`
		Status    string `json:"status"`
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert string status back to TxStatus
	switch aux.Status {
	case "pending":
		t.Status = TxStatusPending
	case "success":
		t.Status = TxStatusSuccess
	case "failed":
		t.Status = TxStatusFailed
	default:
		// If status is empty or unknown, keep the value from aux.Alias
		// which would be the numeric value
	}

	// Convert string chain type back to ChainType
	if aux.ChainType != "" {
		t.ChainType = ChainType(aux.ChainType)
	}

	return nil
}

// TransactionSummary represents a lightweight transaction summary
type TransactionSummary struct {
	ChainType   ChainType  `json:"chain_type"`
	ChainID     string     `json:"chain_id"`
	Hash        string     `json:"hash"`
	BlockNumber uint64     `json:"block_number"`
	From        string     `json:"from"`
	To          string     `json:"to,omitempty"`
	Value       string     `json:"value"`
	Status      TxStatus   `json:"status"`
	Timestamp   *Timestamp `json:"timestamp"`
}

// ToSummary converts a Transaction to a TransactionSummary
func (t *Transaction) ToSummary() *TransactionSummary {
	return &TransactionSummary{
		ChainType:   t.ChainType,
		ChainID:     t.ChainID,
		Hash:        t.Hash,
		BlockNumber: t.BlockNumber,
		From:        t.From,
		To:          t.To,
		Value:       t.Value,
		Status:      t.Status,
		Timestamp:   t.Timestamp,
	}
}

// TransactionFilter represents filtering criteria for transaction queries
type TransactionFilter struct {
	ChainType      *ChainType `json:"chain_type,omitempty"`
	ChainID        *string    `json:"chain_id,omitempty"`
	BlockNumberMin *uint64    `json:"block_number_min,omitempty"`
	BlockNumberMax *uint64    `json:"block_number_max,omitempty"`
	From           *string    `json:"from,omitempty"`
	To             *string    `json:"to,omitempty"`
	Status         *TxStatus  `json:"status,omitempty"`
	TimeMin        *time.Time `json:"time_min,omitempty"`
	TimeMax        *time.Time `json:"time_max,omitempty"`
}

// Log represents a transaction log/event
type Log struct {
	Index   uint64   `json:"index"`   // Log index within transaction
	Address string   `json:"address"` // Contract address that emitted the log
	Topics  []string `json:"topics"`  // Event topics
	Data    []byte   `json:"data"`    // Event data
}

// NewLog creates a new Log
func NewLog(index uint64, address string) *Log {
	return &Log{
		Index:   index,
		Address: address,
		Topics:  make([]string, 0),
	}
}

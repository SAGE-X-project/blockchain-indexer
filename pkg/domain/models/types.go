package models

import "time"

// ChainType represents the type of blockchain
type ChainType string

const (
	// ChainTypeEVM represents EVM-compatible blockchains
	ChainTypeEVM ChainType = "evm"

	// ChainTypeSolana represents Solana blockchain
	ChainTypeSolana ChainType = "solana"

	// ChainTypeCosmos represents Cosmos ecosystem blockchains
	ChainTypeCosmos ChainType = "cosmos"

	// ChainTypePolkadot represents Polkadot/Substrate-based chains
	ChainTypePolkadot ChainType = "polkadot"

	// ChainTypeAvalanche represents Avalanche blockchain
	ChainTypeAvalanche ChainType = "avalanche"

	// ChainTypeRipple represents Ripple (XRPL) blockchain
	ChainTypeRipple ChainType = "ripple"
)

// String returns the string representation of ChainType
func (c ChainType) String() string {
	return string(c)
}

// IsValid checks if the chain type is valid
func (c ChainType) IsValid() bool {
	switch c {
	case ChainTypeEVM, ChainTypeSolana, ChainTypeCosmos,
		ChainTypePolkadot, ChainTypeAvalanche, ChainTypeRipple:
		return true
	default:
		return false
	}
}

// TxStatus represents the status of a transaction
type TxStatus uint8

const (
	// TxStatusPending indicates the transaction is pending
	TxStatusPending TxStatus = 0

	// TxStatusSuccess indicates the transaction succeeded
	TxStatusSuccess TxStatus = 1

	// TxStatusFailed indicates the transaction failed
	TxStatusFailed TxStatus = 2
)

// String returns the string representation of TxStatus
func (s TxStatus) String() string {
	switch s {
	case TxStatusPending:
		return "pending"
	case TxStatusSuccess:
		return "success"
	case TxStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// ChainInfo represents basic chain information
type ChainInfo struct {
	ChainType ChainType `json:"chain_type"`
	ChainID   string    `json:"chain_id"`
	Name      string    `json:"name"`
	Network   string    `json:"network"` // mainnet, testnet, devnet, etc.
}

// Timestamp represents a blockchain timestamp
type Timestamp struct {
	Unix  int64     `json:"unix"`
	Time  time.Time `json:"time"`
	Slot  *uint64   `json:"slot,omitempty"`  // For Solana and other slot-based chains
	Epoch *uint64   `json:"epoch,omitempty"` // For epoch-based chains
}

// NewTimestamp creates a new Timestamp from Unix timestamp
func NewTimestamp(unix int64) *Timestamp {
	return &Timestamp{
		Unix: unix,
		Time: time.Unix(unix, 0),
	}
}

// NewTimestampWithSlot creates a new Timestamp with slot information
func NewTimestampWithSlot(unix int64, slot uint64) *Timestamp {
	return &Timestamp{
		Unix: unix,
		Time: time.Unix(unix, 0),
		Slot: &slot,
	}
}

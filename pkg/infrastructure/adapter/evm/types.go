package evm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// EVMBlock represents an Ethereum-compatible block with all its data
type EVMBlock struct {
	Number           *big.Int
	Hash             common.Hash
	ParentHash       common.Hash
	Nonce            uint64
	Sha3Uncles       common.Hash
	TransactionsRoot common.Hash
	StateRoot        common.Hash
	ReceiptsRoot     common.Hash
	Miner            common.Address
	Difficulty       *big.Int
	TotalDifficulty  *big.Int
	ExtraData        []byte
	Size             uint64
	GasLimit         uint64
	GasUsed          uint64
	Timestamp        uint64
	Transactions     []EVMTransaction
	Uncles           []common.Hash
	BaseFeePerGas    *big.Int // EIP-1559
}

// EVMTransaction represents an Ethereum-compatible transaction
type EVMTransaction struct {
	Hash             common.Hash
	BlockHash        common.Hash
	BlockNumber      *big.Int
	From             common.Address
	To               *common.Address // nil for contract creation
	Value            *big.Int
	Gas              uint64
	GasPrice         *big.Int
	GasFeeCap        *big.Int // EIP-1559
	GasTipCap        *big.Int // EIP-1559
	Input            []byte
	Nonce            uint64
	TransactionIndex uint64
	V                *big.Int
	R                *big.Int
	S                *big.Int
	Type             uint8 // 0: legacy, 1: access list, 2: dynamic fee
	ChainID          *big.Int
	AccessList       types.AccessList
	Receipt          *EVMReceipt
}

// EVMReceipt represents a transaction receipt
type EVMReceipt struct {
	TransactionHash   common.Hash
	TransactionIndex  uint64
	BlockHash         common.Hash
	BlockNumber       *big.Int
	From              common.Address
	To                *common.Address
	CumulativeGasUsed uint64
	GasUsed           uint64
	EffectiveGasPrice *big.Int
	ContractAddress   *common.Address // nil if not a contract creation
	Logs              []EVMLog
	LogsBloom         []byte
	Status            uint64 // 1 = success, 0 = failure
	Type              uint8
}

// EVMLog represents an event log
type EVMLog struct {
	Address          common.Address
	Topics           []common.Hash
	Data             []byte
	BlockNumber      uint64
	TransactionHash  common.Hash
	TransactionIndex uint64
	BlockHash        common.Hash
	LogIndex         uint64
	Removed          bool
}

// ChainInfo represents EVM chain information
type ChainInfo struct {
	ChainID         *big.Int
	NetworkID       *big.Int
	LatestBlock     uint64
	SyncProgress    *SyncProgress
	PeerCount       uint64
	ProtocolVersion string
}

// SyncProgress represents synchronization progress
type SyncProgress struct {
	StartingBlock uint64
	CurrentBlock  uint64
	HighestBlock  uint64
	PulledStates  uint64
	KnownStates   uint64
}

// BlockRange represents a range of blocks to fetch
type BlockRange struct {
	From uint64
	To   uint64
}

// LogFilter represents parameters for filtering logs
type LogFilter struct {
	FromBlock *big.Int
	ToBlock   *big.Int
	Addresses []common.Address
	Topics    [][]common.Hash
}

// SubscriptionType represents the type of subscription
type SubscriptionType int

const (
	SubscriptionTypeNewHeads SubscriptionType = iota
	SubscriptionTypeNewPendingTransactions
	SubscriptionTypeLogs
	SubscriptionTypeSyncing
)

// Subscription represents an active subscription
type Subscription struct {
	Type   SubscriptionType
	ID     string
	Active bool
	Ch     chan interface{}
	ErrCh  chan error
}

// RPCError represents an RPC error
type RPCError struct {
	Code    int
	Message string
	Data    interface{}
}

func (e *RPCError) Error() string {
	return e.Message
}

// HealthStatus represents the health status of the adapter
type HealthStatus struct {
	Connected       bool
	BlockHeight     uint64
	PeerCount       uint64
	Syncing         bool
	SyncProgress    *SyncProgress
	LastBlockTime   uint64
	ResponseTime    int64 // milliseconds
	ErrorCount      uint64
	LastError       string
	LastChecked     int64 // unix timestamp
}

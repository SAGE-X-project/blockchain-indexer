package graphql

import (
	"fmt"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// GraphQL scalar types
type BigInt string
type Time time.Time

// ChainType enum
type ChainType string

const (
	ChainTypeEVM    ChainType = "EVM"
	ChainTypeSolana ChainType = "SOLANA"
	ChainTypeCosmos ChainType = "COSMOS"
)

// ChainStatus enum
type ChainStatus string

const (
	ChainStatusActive   ChainStatus = "ACTIVE"
	ChainStatusInactive ChainStatus = "INACTIVE"
	ChainStatusSyncing  ChainStatus = "SYNCING"
	ChainStatusError    ChainStatus = "ERROR"
)

// TransactionStatus enum
type TransactionStatus string

const (
	TransactionStatusPending TransactionStatus = "PENDING"
	TransactionStatusSuccess TransactionStatus = "SUCCESS"
	TransactionStatusFailed  TransactionStatus = "FAILED"
)

// Chain represents chain information
type Chain struct {
	ChainID            string
	ChainType          ChainType
	Name               string
	Network            string
	Status             ChainStatus
	StartBlock         BigInt
	LatestIndexedBlock BigInt
	LatestChainBlock   BigInt
	LastUpdated        Time
}

// Block represents a blockchain block
type Block struct {
	ChainID        string
	ChainType      ChainType
	Number         BigInt
	Hash           string
	ParentHash     string
	Timestamp      Time
	GasUsed        *BigInt
	GasLimit       *BigInt
	BaseFee        *BigInt
	Difficulty     *BigInt
	Miner          *string
	ExtraData      *string
	TxCount        int
	Transactions   []*Transaction
	CreatedAt      Time
}

// Transaction represents a blockchain transaction
type Transaction struct {
	ChainID         string
	Hash            string
	BlockNumber     BigInt
	BlockHash       string
	BlockTimestamp  Time
	TxIndex         int
	From            string
	To              *string
	Value           BigInt
	GasPrice        *BigInt
	GasLimit        BigInt
	GasUsed         *BigInt
	Nonce           BigInt
	Input           *string
	Status          TransactionStatus
	ContractAddress *string
	Logs            []*Log
	CreatedAt       Time
}

// Log represents a transaction log/event
type Log struct {
	Address  string
	Topics   []string
	Data     string
	LogIndex int
}

// Progress represents indexing progress
type Progress struct {
	ChainID            string
	ChainType          string
	LatestIndexedBlock BigInt
	LatestChainBlock   BigInt
	TargetBlock        BigInt
	StartBlock         BigInt
	BlocksBehind       BigInt
	ProgressPercentage float64
	BlocksPerSecond    float64
	EstimatedTimeLeft  string
	LastUpdated        Time
	Status             string
}

// Gap represents a gap in indexed blocks
type Gap struct {
	ChainID    string
	StartBlock BigInt
	EndBlock   BigInt
	Size       BigInt
}

// Stats represents statistics
type Stats struct {
	TotalBlocks       BigInt
	TotalTransactions BigInt
	ChainsIndexed     int
	AverageBlockTime  float64
	AverageTxPerBlock float64
}

// PageInfo represents pagination information
type PageInfo struct {
	HasNextPage     bool
	HasPreviousPage bool
	StartCursor     *string
	EndCursor       *string
	TotalCount      int
}

// BlockConnection represents a paginated list of blocks
type BlockConnection struct {
	Edges    []*BlockEdge
	PageInfo *PageInfo
}

// BlockEdge represents a block edge in pagination
type BlockEdge struct {
	Node   *Block
	Cursor string
}

// TransactionConnection represents a paginated list of transactions
type TransactionConnection struct {
	Edges    []*TransactionEdge
	PageInfo *PageInfo
}

// TransactionEdge represents a transaction edge in pagination
type TransactionEdge struct {
	Node   *Transaction
	Cursor string
}

// Converters from domain models to GraphQL types

// ToGraphQLChainType converts domain ChainType to GraphQL ChainType
func ToGraphQLChainType(ct models.ChainType) ChainType {
	switch ct {
	case models.ChainTypeEVM:
		return ChainTypeEVM
	case models.ChainTypeSolana:
		return ChainTypeSolana
	case models.ChainTypeCosmos:
		return ChainTypeCosmos
	default:
		return ChainTypeEVM
	}
}

// ToGraphQLTxStatus converts domain TxStatus to GraphQL TransactionStatus
func ToGraphQLTxStatus(status models.TxStatus) TransactionStatus {
	switch status {
	case models.TxStatusPending:
		return TransactionStatusPending
	case models.TxStatusSuccess:
		return TransactionStatusSuccess
	case models.TxStatusFailed:
		return TransactionStatusFailed
	default:
		return TransactionStatusPending
	}
}

// ToGraphQLBlock converts domain Block to GraphQL Block
func ToGraphQLBlock(block *models.Block) *Block {
	if block == nil {
		return nil
	}

	gqlBlock := &Block{
		ChainID:      block.ChainID,
		ChainType:    ToGraphQLChainType(block.ChainType),
		Number:       BigInt(uint64ToString(block.Number)),
		Hash:         block.Hash,
		ParentHash:   block.ParentHash,
		Timestamp:    Time(block.Timestamp.Time),
		TxCount:      block.TxCount,
		Transactions: make([]*Transaction, 0, len(block.Transactions)),
		CreatedAt:    Time(block.IndexedAt),
	}

	// Gas fields
	gasUsed := BigInt(uint64ToString(block.GasUsed))
	gqlBlock.GasUsed = &gasUsed
	gasLimit := BigInt(uint64ToString(block.GasLimit))
	gqlBlock.GasLimit = &gasLimit

	// Optional fields from metadata
	if block.Proposer != "" {
		gqlBlock.Miner = &block.Proposer
	}

	// Convert transactions
	for _, tx := range block.Transactions {
		gqlBlock.Transactions = append(gqlBlock.Transactions, ToGraphQLTransaction(tx))
	}

	return gqlBlock
}

// ToGraphQLTransaction converts domain Transaction to GraphQL Transaction
func ToGraphQLTransaction(tx *models.Transaction) *Transaction {
	if tx == nil {
		return nil
	}

	gqlTx := &Transaction{
		ChainID:        tx.ChainID,
		Hash:           tx.Hash,
		BlockNumber:    BigInt(uint64ToString(tx.BlockNumber)),
		BlockHash:      tx.BlockHash,
		BlockTimestamp: Time(tx.Timestamp.Time),
		TxIndex:        int(tx.Index),
		From:           tx.From,
		Value:          BigInt(tx.Value),
		GasLimit:       BigInt("0"), // Not in domain model
		Nonce:          BigInt(uint64ToString(tx.Nonce)),
		Status:         ToGraphQLTxStatus(tx.Status),
		Logs:           make([]*Log, 0, len(tx.Logs)),
		CreatedAt:      Time(tx.IndexedAt),
	}

	// Optional fields
	if tx.To != "" {
		gqlTx.To = &tx.To
	}

	// Gas price and gas used
	gasPrice := BigInt(tx.GasPrice)
	gqlTx.GasPrice = &gasPrice
	gasUsed := BigInt(uint64ToString(tx.GasUsed))
	gqlTx.GasUsed = &gasUsed

	// Input data
	if len(tx.Input) > 0 {
		inputStr := fmt.Sprintf("0x%x", tx.Input)
		gqlTx.Input = &inputStr
	}

	// Contract address
	if tx.ContractAddress != "" {
		gqlTx.ContractAddress = &tx.ContractAddress
	}

	// Convert logs
	for _, log := range tx.Logs {
		gqlTx.Logs = append(gqlTx.Logs, &Log{
			Address:  log.Address,
			Topics:   log.Topics,
			Data:     fmt.Sprintf("0x%x", log.Data),
			LogIndex: int(log.Index),
		})
	}

	return gqlTx
}

// Helper function to convert uint64 to string
func uint64ToString(n uint64) string {
	return fmt.Sprintf("%d", n)
}

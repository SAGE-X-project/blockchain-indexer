package handler

import (
	"time"
)

// API Response types

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ChainResponse represents a chain in the API
type ChainResponse struct {
	ChainID            string    `json:"chain_id"`
	ChainType          string    `json:"chain_type"`
	Name               string    `json:"name"`
	Network            string    `json:"network"`
	Status             string    `json:"status"`
	StartBlock         uint64    `json:"start_block"`
	LatestIndexedBlock uint64    `json:"latest_indexed_block"`
	LatestChainBlock   uint64    `json:"latest_chain_block"`
	LastUpdated        time.Time `json:"last_updated"`
}

// BlockResponse represents a block in the API
type BlockResponse struct {
	ChainID      string              `json:"chain_id"`
	ChainType    string              `json:"chain_type"`
	Number       uint64              `json:"number"`
	Hash         string              `json:"hash"`
	ParentHash   string              `json:"parent_hash"`
	Timestamp    time.Time           `json:"timestamp"`
	GasUsed      uint64              `json:"gas_used"`
	GasLimit     uint64              `json:"gas_limit"`
	Miner        string              `json:"miner,omitempty"`
	TxCount      int                 `json:"tx_count"`
	Transactions []TransactionResponse `json:"transactions,omitempty"`
	IndexedAt    time.Time           `json:"indexed_at"`
}

// TransactionResponse represents a transaction in the API
type TransactionResponse struct {
	ChainID         string        `json:"chain_id"`
	Hash            string        `json:"hash"`
	BlockNumber     uint64        `json:"block_number"`
	BlockHash       string        `json:"block_hash"`
	BlockTimestamp  time.Time     `json:"block_timestamp"`
	TxIndex         uint64        `json:"tx_index"`
	From            string        `json:"from"`
	To              string        `json:"to,omitempty"`
	Value           string        `json:"value"`
	GasPrice        string        `json:"gas_price"`
	GasUsed         uint64        `json:"gas_used"`
	Nonce           uint64        `json:"nonce"`
	Input           string        `json:"input,omitempty"`
	Status          string        `json:"status"`
	ContractAddress string        `json:"contract_address,omitempty"`
	Logs            []LogResponse `json:"logs,omitempty"`
	IndexedAt       time.Time     `json:"indexed_at"`
}

// LogResponse represents a transaction log in the API
type LogResponse struct {
	Address  string   `json:"address"`
	Topics   []string `json:"topics"`
	Data     string   `json:"data"`
	LogIndex uint64   `json:"log_index"`
}

// ProgressResponse represents indexing progress
type ProgressResponse struct {
	ChainID            string    `json:"chain_id"`
	ChainType          string    `json:"chain_type"`
	LatestIndexedBlock uint64    `json:"latest_indexed_block"`
	LatestChainBlock   uint64    `json:"latest_chain_block"`
	TargetBlock        uint64    `json:"target_block"`
	StartBlock         uint64    `json:"start_block"`
	BlocksBehind       uint64    `json:"blocks_behind"`
	ProgressPercentage float64   `json:"progress_percentage"`
	BlocksPerSecond    float64   `json:"blocks_per_second"`
	EstimatedTimeLeft  string    `json:"estimated_time_left"`
	LastUpdated        time.Time `json:"last_updated"`
	Status             string    `json:"status"`
}

// GapInfo represents a gap in indexed blocks
type GapInfo struct {
	ChainID    string `json:"chain_id"`
	StartBlock uint64 `json:"start_block"`
	EndBlock   uint64 `json:"end_block"`
	Size       uint64 `json:"size"`
}

// GapsResponse represents a list of gaps
type GapsResponse struct {
	Gaps  []GapInfo `json:"gaps"`
	Count int       `json:"count"`
}

// StatsResponse represents statistics
type StatsResponse struct {
	TotalBlocks       uint64  `json:"total_blocks"`
	TotalTransactions uint64  `json:"total_transactions"`
	ChainsIndexed     int     `json:"chains_indexed"`
	AverageBlockTime  float64 `json:"average_block_time"`
	AverageTxPerBlock float64 `json:"average_tx_per_block"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{}     `json:"data"`
	Pagination PaginationInfo  `json:"pagination"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalCount int  `json:"total_count"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Checks    map[string]string `json:"checks"`
}

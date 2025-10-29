package cosmos

import (
	"encoding/json"
	"time"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
)

// Block represents a Cosmos/Tendermint block
type Block struct {
	*coretypes.ResultBlock
}

// Transaction represents a Cosmos transaction
type Transaction struct {
	TxHash    string          `json:"tx_hash"`
	Height    int64           `json:"height"`
	Index     uint32          `json:"index"`
	Code      uint32          `json:"code"`
	Data      string          `json:"data"`
	RawLog    string          `json:"raw_log"`
	Info      string          `json:"info"`
	GasWanted int64           `json:"gas_wanted"`
	GasUsed   int64           `json:"gas_used"`
	Timestamp time.Time       `json:"timestamp"`
	Events    []Event         `json:"events"`
	Tx        json.RawMessage `json:"tx"`
	Codespace string          `json:"codespace"`
}

// Event represents a transaction event
type Event struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

// Attribute represents an event attribute
type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Index bool   `json:"index"`
}

// BlockResult represents a block's results
type BlockResult struct {
	Height                int64                  `json:"height"`
	TxsResults            []*ExecTxResult        `json:"txs_results"`
	BeginBlockEvents      []Event                `json:"begin_block_events"`
	EndBlockEvents        []Event                `json:"end_block_events"`
	ValidatorUpdates      []ValidatorUpdate      `json:"validator_updates"`
	ConsensusParamUpdates *ConsensusParamUpdates `json:"consensus_param_updates"`
}

// ExecTxResult represents a transaction execution result
type ExecTxResult struct {
	Code      uint32  `json:"code"`
	Data      []byte  `json:"data"`
	Log       string  `json:"log"`
	Info      string  `json:"info"`
	GasWanted int64   `json:"gas_wanted"`
	GasUsed   int64   `json:"gas_used"`
	Events    []Event `json:"events"`
	Codespace string  `json:"codespace"`
}

// ValidatorUpdate represents a validator update
type ValidatorUpdate struct {
	PubKey PubKey `json:"pub_key"`
	Power  int64  `json:"power"`
}

// PubKey represents a public key
type PubKey struct {
	Type string `json:"type"`
	Key  []byte `json:"key"`
}

// ConsensusParamUpdates represents consensus parameter updates
type ConsensusParamUpdates struct {
	Block     *BlockParams     `json:"block"`
	Evidence  *EvidenceParams  `json:"evidence"`
	Validator *ValidatorParams `json:"validator"`
	Version   *VersionParams   `json:"version"`
}

// BlockParams represents block parameters
type BlockParams struct {
	MaxBytes int64 `json:"max_bytes"`
	MaxGas   int64 `json:"max_gas"`
}

// EvidenceParams represents evidence parameters
type EvidenceParams struct {
	MaxAgeNumBlocks int64         `json:"max_age_num_blocks"`
	MaxAgeDuration  time.Duration `json:"max_age_duration"`
	MaxBytes        int64         `json:"max_bytes"`
}

// ValidatorParams represents validator parameters
type ValidatorParams struct {
	PubKeyTypes []string `json:"pub_key_types"`
}

// VersionParams represents version parameters
type VersionParams struct {
	App uint64 `json:"app"`
}

// NodeInfo represents node information
type NodeInfo struct {
	ProtocolVersion ProtocolVersion `json:"protocol_version"`
	ID              string          `json:"id"`
	ListenAddr      string          `json:"listen_addr"`
	Network         string          `json:"network"`
	Version         string          `json:"version"`
	Channels        string          `json:"channels"`
	Moniker         string          `json:"moniker"`
	Other           NodeInfoOther   `json:"other"`
}

// ProtocolVersion represents protocol version
type ProtocolVersion struct {
	P2P   uint64 `json:"p2p"`
	Block uint64 `json:"block"`
	App   uint64 `json:"app"`
}

// NodeInfoOther represents other node information
type NodeInfoOther struct {
	TxIndex    string `json:"tx_index"`
	RPCAddress string `json:"rpc_address"`
}

// SyncInfo represents sync information
type SyncInfo struct {
	LatestBlockHash     string    `json:"latest_block_hash"`
	LatestAppHash       string    `json:"latest_app_hash"`
	LatestBlockHeight   int64     `json:"latest_block_height"`
	LatestBlockTime     time.Time `json:"latest_block_time"`
	EarliestBlockHash   string    `json:"earliest_block_hash"`
	EarliestAppHash     string    `json:"earliest_app_hash"`
	EarliestBlockHeight int64     `json:"earliest_block_height"`
	EarliestBlockTime   time.Time `json:"earliest_block_time"`
	CatchingUp          bool      `json:"catching_up"`
}

// ValidatorInfo represents validator information
type ValidatorInfo struct {
	Address     string `json:"address"`
	PubKey      PubKey `json:"pub_key"`
	VotingPower int64  `json:"voting_power"`
}

// Status represents node status
type Status struct {
	NodeInfo      NodeInfo      `json:"node_info"`
	SyncInfo      SyncInfo      `json:"sync_info"`
	ValidatorInfo ValidatorInfo `json:"validator_info"`
}

// BlockMeta represents block metadata
type BlockMeta struct {
	BlockID   BlockID         `json:"block_id"`
	BlockSize int             `json:"block_size"`
	Header    tmtypes.Header  `json:"header"`
	NumTxs    int             `json:"num_txs"`
}

// BlockID represents a block identifier
type BlockID struct {
	Hash  string        `json:"hash"`
	Parts PartSetHeader `json:"parts"`
}

// PartSetHeader represents a part set header
type PartSetHeader struct {
	Total uint32 `json:"total"`
	Hash  string `json:"hash"`
}

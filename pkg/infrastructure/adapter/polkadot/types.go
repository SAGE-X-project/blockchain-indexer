package polkadot

import (
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// Block represents a Polkadot/Substrate block
type Block struct {
	Header       *Header          `json:"header"`
	Extrinsics   []types.Extrinsic `json:"extrinsics"`
	Justification *types.Bytes     `json:"justification,omitempty"`
}

// Header represents a block header
type Header struct {
	ParentHash     string `json:"parent_hash"`
	Number         uint64 `json:"number"`
	StateRoot      string `json:"state_root"`
	ExtrinsicsRoot string `json:"extrinsics_root"`
	Digest         Digest `json:"digest"`
}

// Digest represents block digest
type Digest struct {
	Logs []DigestItem `json:"logs"`
}

// DigestItem represents a digest item
type DigestItem struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// SignedBlock represents a signed block with justification
type SignedBlock struct {
	Block         *Block       `json:"block"`
	Justification *types.Bytes `json:"justification,omitempty"`
}

// BlockHash represents a block hash
type BlockHash = types.Hash

// Extrinsic represents a blockchain transaction/extrinsic
type Extrinsic struct {
	Version   byte                   `json:"version"`
	Signature *ExtrinsicSignature    `json:"signature,omitempty"`
	Method    *Call                  `json:"method"`
	Era       *ExtrinsicEra          `json:"era,omitempty"`
	Nonce     uint32                 `json:"nonce,omitempty"`
	Tip       uint64                 `json:"tip,omitempty"`
	IsSigned  bool                   `json:"is_signed"`
	Hash      string                 `json:"hash"`
}

// ExtrinsicSignature represents an extrinsic signature
type ExtrinsicSignature struct {
	Signer    string `json:"signer"`
	Signature string `json:"signature"`
}

// ExtrinsicEra represents the era of an extrinsic
type ExtrinsicEra struct {
	IsMortalEra bool   `json:"is_mortal_era"`
	Period      uint64 `json:"period,omitempty"`
	Phase       uint64 `json:"phase,omitempty"`
}

// Call represents a method call
type Call struct {
	CallIndex string                 `json:"call_index"`
	Module    string                 `json:"module"`
	Function  string                 `json:"function"`
	Args      map[string]interface{} `json:"args"`
}

// Event represents a blockchain event
type Event struct {
	Phase  EventPhase             `json:"phase"`
	Event  EventRecord            `json:"event"`
	Topics []string               `json:"topics"`
}

// EventPhase represents the phase of an event
type EventPhase struct {
	IsApplyExtrinsic bool   `json:"is_apply_extrinsic"`
	AsApplyExtrinsic uint32 `json:"as_apply_extrinsic,omitempty"`
	IsFinalization   bool   `json:"is_finalization"`
	IsInitialization bool   `json:"is_initialization"`
}

// EventRecord represents an event record
type EventRecord struct {
	Module string                 `json:"module"`
	Event  string                 `json:"event"`
	Data   map[string]interface{} `json:"data"`
}

// RuntimeVersion represents the runtime version
type RuntimeVersion struct {
	SpecName         string            `json:"spec_name"`
	ImplName         string            `json:"impl_name"`
	AuthoringVersion uint32            `json:"authoring_version"`
	SpecVersion      uint32            `json:"spec_version"`
	ImplVersion      uint32            `json:"impl_version"`
	Apis             []RuntimeVersionAPI `json:"apis"`
	TransactionVersion uint32          `json:"transaction_version"`
	StateVersion     uint8             `json:"state_version"`
}

// RuntimeVersionAPI represents a runtime API
type RuntimeVersionAPI struct {
	ID      string `json:"id"`
	Version uint32 `json:"version"`
}

// Metadata represents chain metadata
type Metadata struct {
	Version    uint8                  `json:"version"`
	Modules    []MetadataModule       `json:"modules"`
}

// MetadataModule represents a metadata module
type MetadataModule struct {
	Name       string                 `json:"name"`
	Storage    *MetadataStorage       `json:"storage,omitempty"`
	Calls      []MetadataCall         `json:"calls,omitempty"`
	Events     []MetadataEvent        `json:"events,omitempty"`
	Constants  []MetadataConstant     `json:"constants,omitempty"`
	Errors     []MetadataError        `json:"errors,omitempty"`
	Index      uint8                  `json:"index"`
}

// MetadataStorage represents metadata storage
type MetadataStorage struct {
	Prefix string               `json:"prefix"`
	Items  []MetadataStorageItem `json:"items"`
}

// MetadataStorageItem represents a storage item
type MetadataStorageItem struct {
	Name     string `json:"name"`
	Modifier string `json:"modifier"`
	Type     string `json:"type"`
}

// MetadataCall represents a metadata call
type MetadataCall struct {
	Name string                 `json:"name"`
	Args []MetadataCallArg      `json:"args"`
}

// MetadataCallArg represents a call argument
type MetadataCallArg struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// MetadataEvent represents a metadata event
type MetadataEvent struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

// MetadataConstant represents a metadata constant
type MetadataConstant struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// MetadataError represents a metadata error
type MetadataError struct {
	Name string `json:"name"`
	Docs string `json:"docs"`
}

// Health represents node health status
type Health struct {
	Peers           int  `json:"peers"`
	IsSyncing       bool `json:"is_syncing"`
	ShouldHavePeers bool `json:"should_have_peers"`
}

// SyncState represents synchronization state
type SyncState struct {
	StartingBlock uint64 `json:"starting_block"`
	CurrentBlock  uint64 `json:"current_block"`
	HighestBlock  uint64 `json:"highest_block"`
}

// ChainProperties represents chain properties
type ChainProperties struct {
	SS58Format    uint16 `json:"ss58_format"`
	TokenDecimals uint8  `json:"token_decimals"`
	TokenSymbol   string `json:"token_symbol"`
}

// AccountInfo represents account information
type AccountInfo struct {
	Nonce       uint32         `json:"nonce"`
	Consumers   uint32         `json:"consumers"`
	Providers   uint32         `json:"providers"`
	Sufficients uint32         `json:"sufficients"`
	Data        AccountData    `json:"data"`
}

// AccountData represents account balance data
type AccountData struct {
	Free       uint64 `json:"free"`
	Reserved   uint64 `json:"reserved"`
	MiscFrozen uint64 `json:"misc_frozen"`
	FeeFrozen  uint64 `json:"fee_frozen"`
}

// StorageChangeSet represents a storage change set
type StorageChangeSet struct {
	Block   string          `json:"block"`
	Changes []StorageChange `json:"changes"`
}

// StorageChange represents a storage change
type StorageChange struct {
	Key   string  `json:"key"`
	Value *string `json:"value,omitempty"`
}

// BlockTimestamp represents block timestamp
type BlockTimestamp struct {
	Height    uint64    `json:"height"`
	Timestamp time.Time `json:"timestamp"`
}

package solana

import "encoding/json"

// Solana RPC types - these match the Solana JSON-RPC API responses

// GetBlockResponse represents a Solana block
type GetBlockResponse struct {
	BlockHeight       *uint64                  `json:"blockHeight,omitempty"`
	BlockTime         *int64                   `json:"blockTime,omitempty"`
	Blockhash         string                   `json:"blockhash"`
	ParentSlot        uint64                   `json:"parentSlot"`
	PreviousBlockhash string                   `json:"previousBlockhash"`
	Transactions      []TransactionWithMeta    `json:"transactions"`
	Rewards           []Reward                 `json:"rewards,omitempty"`
	Signatures        []string                 `json:"signatures,omitempty"`
}

// TransactionWithMeta represents a transaction with metadata
type TransactionWithMeta struct {
	Meta        *TransactionMeta `json:"meta,omitempty"`
	Transaction interface{}      `json:"transaction"` // Can be Transaction or []string (signatures only)
	Version     interface{}      `json:"version,omitempty"`
}

// Transaction represents a Solana transaction
type Transaction struct {
	Signatures []string        `json:"signatures"`
	Message    TransactionMessage `json:"message"`
}

// TransactionMessage represents a transaction message
type TransactionMessage struct {
	AccountKeys          []string                 `json:"accountKeys"`
	RecentBlockhash      string                   `json:"recentBlockhash"`
	Instructions         []CompiledInstruction    `json:"instructions"`
	AddressTableLookups  []AddressTableLookup     `json:"addressTableLookups,omitempty"`
	Header               MessageHeader            `json:"header"`
}

// MessageHeader represents the message header
type MessageHeader struct {
	NumRequiredSignatures       uint8 `json:"numRequiredSignatures"`
	NumReadonlySignedAccounts   uint8 `json:"numReadonlySignedAccounts"`
	NumReadonlyUnsignedAccounts uint8 `json:"numReadonlyUnsignedAccounts"`
}

// CompiledInstruction represents a compiled instruction
type CompiledInstruction struct {
	ProgramIdIndex uint16 `json:"programIdIndex"`
	Accounts       []uint8 `json:"accounts"`
	Data           string  `json:"data"` // Base58 encoded
}

// AddressTableLookup represents an address table lookup
type AddressTableLookup struct {
	AccountKey      string   `json:"accountKey"`
	WritableIndexes []uint8  `json:"writableIndexes"`
	ReadonlyIndexes []uint8  `json:"readonlyIndexes"`
}

// TransactionMeta represents transaction metadata
type TransactionMeta struct {
	Err               interface{}          `json:"err"`
	Fee               uint64               `json:"fee"`
	PreBalances       []uint64             `json:"preBalances"`
	PostBalances      []uint64             `json:"postBalances"`
	InnerInstructions []InnerInstruction   `json:"innerInstructions,omitempty"`
	LogMessages       []string             `json:"logMessages,omitempty"`
	PreTokenBalances  []TokenBalance       `json:"preTokenBalances,omitempty"`
	PostTokenBalances []TokenBalance       `json:"postTokenBalances,omitempty"`
	Rewards           []Reward             `json:"rewards,omitempty"`
	LoadedAddresses   *LoadedAddresses     `json:"loadedAddresses,omitempty"`
	ComputeUnitsConsumed *uint64           `json:"computeUnitsConsumed,omitempty"`
	Status            map[string]interface{} `json:"status,omitempty"`
}

// InnerInstruction represents an inner instruction
type InnerInstruction struct {
	Index        uint16                `json:"index"`
	Instructions []CompiledInstruction `json:"instructions"`
}

// TokenBalance represents a token balance
type TokenBalance struct {
	AccountIndex  uint16      `json:"accountIndex"`
	Mint          string      `json:"mint"`
	Owner         string      `json:"owner,omitempty"`
	ProgramId     string      `json:"programId,omitempty"`
	UiTokenAmount UiTokenAmount `json:"uiTokenAmount"`
}

// UiTokenAmount represents a UI-friendly token amount
type UiTokenAmount struct {
	Amount         string  `json:"amount"`
	Decimals       uint8   `json:"decimals"`
	UiAmount       *float64 `json:"uiAmount"`
	UiAmountString string  `json:"uiAmountString"`
}

// Reward represents a block reward
type Reward struct {
	Pubkey      string  `json:"pubkey"`
	Lamports    int64   `json:"lamports"`
	PostBalance uint64  `json:"postBalance"`
	RewardType  *string `json:"rewardType,omitempty"`
	Commission  *uint8  `json:"commission,omitempty"`
}

// LoadedAddresses represents loaded addresses
type LoadedAddresses struct {
	Writable []string `json:"writable"`
	Readonly []string `json:"readonly"`
}

// GetSlotResponse represents the current slot
type GetSlotResponse uint64

// GetBlockHeightResponse represents the current block height
type GetBlockHeightResponse uint64

// GetHealthResponse represents health status
type GetHealthResponse string

// GetVersionResponse represents the version
type GetVersionResponse struct {
	SolanaCore    string `json:"solana-core"`
	FeatureSet    uint32 `json:"feature-set"`
}

// GetTransactionResponse represents a transaction response
type GetTransactionResponse struct {
	Slot        uint64                `json:"slot"`
	BlockTime   *int64                `json:"blockTime,omitempty"`
	Meta        *TransactionMeta      `json:"meta,omitempty"`
	Transaction interface{}           `json:"transaction"` // Can be Transaction or []string
	Version     interface{}           `json:"version,omitempty"`
}

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params,omitempty"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SlotUpdate represents a slot update notification
type SlotUpdate struct {
	Parent uint64 `json:"parent"`
	Root   uint64 `json:"root"`
	Slot   uint64 `json:"slot"`
	Type   string `json:"type"`
}

// Encoding types for transaction details
const (
	EncodingJSON       = "json"
	EncodingJSONParsed = "jsonParsed"
	EncodingBase58     = "base58"
	EncodingBase64     = "base64"
)

// Commitment levels
const (
	CommitmentProcessed = "processed"
	CommitmentConfirmed = "confirmed"
	CommitmentFinalized = "finalized"
)

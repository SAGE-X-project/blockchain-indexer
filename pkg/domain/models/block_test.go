package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewBlock(t *testing.T) {
	chainType := ChainTypeEVM
	chainID := "ethereum"
	number := uint64(12345)
	hash := "0xabc123"

	block := NewBlock(chainType, chainID, number, hash)

	if block.ChainType != chainType {
		t.Errorf("NewBlock().ChainType = %v, want %v", block.ChainType, chainType)
	}

	if block.ChainID != chainID {
		t.Errorf("NewBlock().ChainID = %v, want %v", block.ChainID, chainID)
	}

	if block.Number != number {
		t.Errorf("NewBlock().Number = %v, want %v", block.Number, number)
	}

	if block.Hash != hash {
		t.Errorf("NewBlock().Hash = %v, want %v", block.Hash, hash)
	}

	if block.TxHashes == nil {
		t.Error("NewBlock().TxHashes should be initialized")
	}

	if block.Metadata == nil {
		t.Error("NewBlock().Metadata should be initialized")
	}

	if block.Timestamp == nil {
		t.Error("NewBlock().Timestamp should be initialized")
	}

	if block.IndexedAt.IsZero() {
		t.Error("NewBlock().IndexedAt should be set")
	}
}

func TestBlock_Validate(t *testing.T) {
	tests := []struct {
		name    string
		block   *Block
		wantErr error
	}{
		{
			name: "valid block",
			block: &Block{
				ChainType: ChainTypeEVM,
				ChainID:   "ethereum",
				Hash:      "0xabc123",
				Timestamp: NewTimestamp(time.Now().Unix()),
			},
			wantErr: nil,
		},
		{
			name: "invalid chain type",
			block: &Block{
				ChainType: ChainType("invalid"),
				ChainID:   "ethereum",
				Hash:      "0xabc123",
				Timestamp: NewTimestamp(time.Now().Unix()),
			},
			wantErr: ErrInvalidChainType,
		},
		{
			name: "empty chain ID",
			block: &Block{
				ChainType: ChainTypeEVM,
				ChainID:   "",
				Hash:      "0xabc123",
				Timestamp: NewTimestamp(time.Now().Unix()),
			},
			wantErr: ErrInvalidChainID,
		},
		{
			name: "empty hash",
			block: &Block{
				ChainType: ChainTypeEVM,
				ChainID:   "ethereum",
				Hash:      "",
				Timestamp: NewTimestamp(time.Now().Unix()),
			},
			wantErr: ErrInvalidBlockHash,
		},
		{
			name: "nil timestamp",
			block: &Block{
				ChainType: ChainTypeEVM,
				ChainID:   "ethereum",
				Hash:      "0xabc123",
				Timestamp: nil,
			},
			wantErr: ErrInvalidTimestamp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.block.Validate()
			if err != tt.wantErr {
				t.Errorf("Block.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlock_Metadata(t *testing.T) {
	block := NewBlock(ChainTypeEVM, "ethereum", 12345, "0xabc123")

	t.Run("SetMetadata and GetMetadata", func(t *testing.T) {
		key := "extra_data"
		value := "test value"

		block.SetMetadata(key, value)

		got, exists := block.GetMetadata(key)
		if !exists {
			t.Error("GetMetadata() should return true for existing key")
		}

		if got != value {
			t.Errorf("GetMetadata() = %v, want %v", got, value)
		}
	})

	t.Run("GetMetadata for non-existent key", func(t *testing.T) {
		_, exists := block.GetMetadata("non_existent")
		if exists {
			t.Error("GetMetadata() should return false for non-existent key")
		}
	})

	t.Run("SetMetadata with nil metadata", func(t *testing.T) {
		block.Metadata = nil
		block.SetMetadata("key", "value")

		if block.Metadata == nil {
			t.Error("SetMetadata() should initialize Metadata if nil")
		}
	})
}

func TestBlock_ToSummary(t *testing.T) {
	block := NewBlock(ChainTypeEVM, "ethereum", 12345, "0xabc123")
	block.ParentHash = "0xparent"
	block.Proposer = "0xminer"
	block.TxCount = 10

	summary := block.ToSummary()

	if summary.ChainType != block.ChainType {
		t.Errorf("ToSummary().ChainType = %v, want %v", summary.ChainType, block.ChainType)
	}

	if summary.ChainID != block.ChainID {
		t.Errorf("ToSummary().ChainID = %v, want %v", summary.ChainID, block.ChainID)
	}

	if summary.Number != block.Number {
		t.Errorf("ToSummary().Number = %v, want %v", summary.Number, block.Number)
	}

	if summary.Hash != block.Hash {
		t.Errorf("ToSummary().Hash = %v, want %v", summary.Hash, block.Hash)
	}

	if summary.ParentHash != block.ParentHash {
		t.Errorf("ToSummary().ParentHash = %v, want %v", summary.ParentHash, block.ParentHash)
	}

	if summary.Proposer != block.Proposer {
		t.Errorf("ToSummary().Proposer = %v, want %v", summary.Proposer, block.Proposer)
	}

	if summary.TxCount != block.TxCount {
		t.Errorf("ToSummary().TxCount = %v, want %v", summary.TxCount, block.TxCount)
	}
}

func TestBlock_MarshalJSON(t *testing.T) {
	block := NewBlock(ChainTypeEVM, "ethereum", 12345, "0xabc123")

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Check that chain_type is a string
	if chainType, ok := result["chain_type"].(string); !ok {
		t.Error("chain_type should be a string in JSON")
	} else if chainType != "evm" {
		t.Errorf("chain_type = %v, want evm", chainType)
	}
}

func TestPaginationOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    *PaginationOptions
		wantErr error
	}{
		{
			name: "valid options",
			opts: &PaginationOptions{
				Limit:  20,
				Offset: 0,
			},
			wantErr: nil,
		},
		{
			name: "invalid limit (zero)",
			opts: &PaginationOptions{
				Limit:  0,
				Offset: 0,
			},
			wantErr: ErrInvalidLimit,
		},
		{
			name: "invalid limit (negative)",
			opts: &PaginationOptions{
				Limit:  -1,
				Offset: 0,
			},
			wantErr: ErrInvalidLimit,
		},
		{
			name: "limit too large",
			opts: &PaginationOptions{
				Limit:  1001,
				Offset: 0,
			},
			wantErr: ErrLimitTooLarge,
		},
		{
			name: "invalid offset",
			opts: &PaginationOptions{
				Limit:  20,
				Offset: -1,
			},
			wantErr: ErrInvalidOffset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if err != tt.wantErr {
				t.Errorf("PaginationOptions.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultPaginationOptions(t *testing.T) {
	opts := DefaultPaginationOptions()

	if opts.Limit != 20 {
		t.Errorf("DefaultPaginationOptions().Limit = %v, want 20", opts.Limit)
	}

	if opts.Offset != 0 {
		t.Errorf("DefaultPaginationOptions().Offset = %v, want 0", opts.Offset)
	}

	if opts.Cursor != nil {
		t.Errorf("DefaultPaginationOptions().Cursor should be nil")
	}
}

// Benchmark tests
func BenchmarkNewBlock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewBlock(ChainTypeEVM, "ethereum", 12345, "0xabc123")
	}
}

func BenchmarkBlock_Validate(b *testing.B) {
	block := NewBlock(ChainTypeEVM, "ethereum", 12345, "0xabc123")
	for i := 0; i < b.N; i++ {
		_ = block.Validate()
	}
}

func BenchmarkBlock_ToSummary(b *testing.B) {
	block := NewBlock(ChainTypeEVM, "ethereum", 12345, "0xabc123")
	for i := 0; i < b.N; i++ {
		_ = block.ToSummary()
	}
}

func BenchmarkBlock_MarshalJSON(b *testing.B) {
	block := NewBlock(ChainTypeEVM, "ethereum", 12345, "0xabc123")
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(block)
	}
}

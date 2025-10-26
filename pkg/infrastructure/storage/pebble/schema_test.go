package pebble

import (
	"bytes"
	"testing"
)

func TestBlockKey(t *testing.T) {
	tests := []struct {
		name        string
		chainID     string
		blockNumber uint64
		want        string
	}{
		{
			name:        "ethereum block 0",
			chainID:     "ethereum",
			blockNumber: 0,
			want:        "block:ethereum:0",
		},
		{
			name:        "ethereum block 12345",
			chainID:     "ethereum",
			blockNumber: 12345,
			want:        "block:ethereum:12345",
		},
		{
			name:        "solana block 1000000",
			chainID:     "solana",
			blockNumber: 1000000,
			want:        "block:solana:1000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BlockKey(tt.chainID, tt.blockNumber)
			if string(got) != tt.want {
				t.Errorf("BlockKey() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestBlockHashKey(t *testing.T) {
	tests := []struct {
		name    string
		chainID string
		hash    string
		want    string
	}{
		{
			name:    "ethereum block hash",
			chainID: "ethereum",
			hash:    "0xabc123",
			want:    "block_hash:ethereum:0xabc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BlockHashKey(tt.chainID, tt.hash)
			if string(got) != tt.want {
				t.Errorf("BlockHashKey() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestTransactionKey(t *testing.T) {
	tests := []struct {
		name    string
		chainID string
		txHash  string
		want    string
	}{
		{
			name:    "ethereum transaction",
			chainID: "ethereum",
			txHash:  "0xtx123",
			want:    "tx:ethereum:0xtx123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransactionKey(tt.chainID, tt.txHash)
			if string(got) != tt.want {
				t.Errorf("TransactionKey() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestTransactionByBlockKey(t *testing.T) {
	tests := []struct {
		name        string
		chainID     string
		blockNumber uint64
		txIndex     uint64
		want        string
	}{
		{
			name:        "first transaction",
			chainID:     "ethereum",
			blockNumber: 12345,
			txIndex:     0,
			want:        "tx_block:ethereum:12345:0",
		},
		{
			name:        "tenth transaction",
			chainID:     "ethereum",
			blockNumber: 12345,
			txIndex:     9,
			want:        "tx_block:ethereum:12345:9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransactionByBlockKey(tt.chainID, tt.blockNumber, tt.txIndex)
			if string(got) != tt.want {
				t.Errorf("TransactionByBlockKey() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestTransactionByBlockPrefix(t *testing.T) {
	chainID := "ethereum"
	blockNumber := uint64(12345)
	want := "tx_block:ethereum:12345:"

	got := TransactionByBlockPrefix(chainID, blockNumber)
	if string(got) != want {
		t.Errorf("TransactionByBlockPrefix() = %v, want %v", string(got), want)
	}

	// Verify that transaction keys match this prefix
	txKey := TransactionByBlockKey(chainID, blockNumber, 0)
	if !bytes.HasPrefix(txKey, got) {
		t.Error("Transaction key should have the transaction-by-block prefix")
	}
}

func TestAddressTxKey(t *testing.T) {
	tests := []struct {
		name        string
		chainID     string
		address     string
		blockNumber uint64
		txIndex     uint64
		want        string
	}{
		{
			name:        "address transaction",
			chainID:     "ethereum",
			address:     "0xabc",
			blockNumber: 12345,
			txIndex:     0,
			want:        "addr_tx:ethereum:0xabc:12345:0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddressTxKey(tt.chainID, tt.address, tt.blockNumber, tt.txIndex)
			if string(got) != tt.want {
				t.Errorf("AddressTxKey() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestAddressTxPrefix(t *testing.T) {
	chainID := "ethereum"
	address := "0xabc"
	want := "addr_tx:ethereum:0xabc:"

	got := AddressTxPrefix(chainID, address)
	if string(got) != want {
		t.Errorf("AddressTxPrefix() = %v, want %v", string(got), want)
	}

	// Verify that address tx keys match this prefix
	addrTxKey := AddressTxKey(chainID, address, 12345, 0)
	if !bytes.HasPrefix(addrTxKey, got) {
		t.Error("Address tx key should have the address tx prefix")
	}
}

func TestChainKey(t *testing.T) {
	chainID := "ethereum"
	want := "chain:ethereum"

	got := ChainKey(chainID)
	if string(got) != want {
		t.Errorf("ChainKey() = %v, want %v", string(got), want)
	}
}

func TestLatestHeightKey(t *testing.T) {
	chainID := "ethereum"
	want := "latest:ethereum"

	got := LatestHeightKey(chainID)
	if string(got) != want {
		t.Errorf("LatestHeightKey() = %v, want %v", string(got), want)
	}
}

func TestParseBlockKey(t *testing.T) {
	tests := []struct {
		name        string
		key         []byte
		wantChainID string
		wantNumber  uint64
		wantErr     bool
	}{
		{
			name:        "valid block key",
			key:         BlockKey("ethereum", 12345),
			wantChainID: "ethereum",
			wantNumber:  12345,
			wantErr:     false,
		},
		{
			name:    "invalid prefix",
			key:     []byte("invalid:ethereum:12345"),
			wantErr: true,
		},
		{
			name:    "invalid format",
			key:     []byte("block:ethereum"),
			wantErr: true,
		},
		{
			name:    "invalid number",
			key:     []byte("block:ethereum:abc"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chainID, number, err := ParseBlockKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBlockKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if chainID != tt.wantChainID {
					t.Errorf("ParseBlockKey() chainID = %v, want %v", chainID, tt.wantChainID)
				}
				if number != tt.wantNumber {
					t.Errorf("ParseBlockKey() number = %v, want %v", number, tt.wantNumber)
				}
			}
		})
	}
}

func TestParseTransactionByBlockKey(t *testing.T) {
	tests := []struct {
		name        string
		key         []byte
		wantChainID string
		wantBlock   uint64
		wantIndex   uint64
		wantErr     bool
	}{
		{
			name:        "valid tx by block key",
			key:         TransactionByBlockKey("ethereum", 12345, 9),
			wantChainID: "ethereum",
			wantBlock:   12345,
			wantIndex:   9,
			wantErr:     false,
		},
		{
			name:    "invalid prefix",
			key:     []byte("invalid:ethereum:12345:9"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chainID, block, index, err := ParseTransactionByBlockKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTransactionByBlockKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if chainID != tt.wantChainID {
					t.Errorf("ParseTransactionByBlockKey() chainID = %v, want %v", chainID, tt.wantChainID)
				}
				if block != tt.wantBlock {
					t.Errorf("ParseTransactionByBlockKey() block = %v, want %v", block, tt.wantBlock)
				}
				if index != tt.wantIndex {
					t.Errorf("ParseTransactionByBlockKey() index = %v, want %v", index, tt.wantIndex)
				}
			}
		})
	}
}

func TestParseAddressTxKey(t *testing.T) {
	tests := []struct {
		name        string
		key         []byte
		wantChainID string
		wantAddress string
		wantBlock   uint64
		wantIndex   uint64
		wantErr     bool
	}{
		{
			name:        "valid address tx key",
			key:         AddressTxKey("ethereum", "0xabc", 12345, 9),
			wantChainID: "ethereum",
			wantAddress: "0xabc",
			wantBlock:   12345,
			wantIndex:   9,
			wantErr:     false,
		},
		{
			name:    "invalid prefix",
			key:     []byte("invalid:ethereum:0xabc:12345:9"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chainID, address, block, index, err := ParseAddressTxKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAddressTxKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if chainID != tt.wantChainID {
					t.Errorf("ParseAddressTxKey() chainID = %v, want %v", chainID, tt.wantChainID)
				}
				if address != tt.wantAddress {
					t.Errorf("ParseAddressTxKey() address = %v, want %v", address, tt.wantAddress)
				}
				if block != tt.wantBlock {
					t.Errorf("ParseAddressTxKey() block = %v, want %v", block, tt.wantBlock)
				}
				if index != tt.wantIndex {
					t.Errorf("ParseAddressTxKey() index = %v, want %v", index, tt.wantIndex)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkBlockKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BlockKey("ethereum", 12345)
	}
}

func BenchmarkTransactionKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = TransactionKey("ethereum", "0xtx123")
	}
}

func BenchmarkParseBlockKey(b *testing.B) {
	key := BlockKey("ethereum", 12345)
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseBlockKey(key)
	}
}

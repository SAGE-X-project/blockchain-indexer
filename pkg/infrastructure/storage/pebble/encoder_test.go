package pebble

import (
	"testing"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

func TestEncoder_EncodeDecodeBlock(t *testing.T) {
	encoder := NewEncoder()
	original := models.NewBlock(models.ChainTypeEVM, "ethereum", 12345, "0xabc123")
	original.ParentHash = "0xparent"
	original.Proposer = "0xminer"
	original.TxCount = 10
	original.GasUsed = 21000
	original.GasLimit = 30000

	t.Run("encode and decode", func(t *testing.T) {
		// Encode
		data, err := encoder.EncodeBlock(original)
		if err != nil {
			t.Fatalf("EncodeBlock() error = %v", err)
		}

		if len(data) == 0 {
			t.Error("EncodeBlock() should return non-empty data")
		}

		// Decode
		decoded, err := encoder.DecodeBlock(data)
		if err != nil {
			t.Fatalf("DecodeBlock() error = %v", err)
		}

		// Verify
		if decoded.ChainType != original.ChainType {
			t.Errorf("ChainType = %v, want %v", decoded.ChainType, original.ChainType)
		}
		if decoded.ChainID != original.ChainID {
			t.Errorf("ChainID = %v, want %v", decoded.ChainID, original.ChainID)
		}
		if decoded.Number != original.Number {
			t.Errorf("Number = %v, want %v", decoded.Number, original.Number)
		}
		if decoded.Hash != original.Hash {
			t.Errorf("Hash = %v, want %v", decoded.Hash, original.Hash)
		}
		if decoded.ParentHash != original.ParentHash {
			t.Errorf("ParentHash = %v, want %v", decoded.ParentHash, original.ParentHash)
		}
	})

	t.Run("encode nil block", func(t *testing.T) {
		_, err := encoder.EncodeBlock(nil)
		if err == nil {
			t.Error("EncodeBlock(nil) should return error")
		}
	})

	t.Run("decode empty data", func(t *testing.T) {
		_, err := encoder.DecodeBlock([]byte{})
		if err == nil {
			t.Error("DecodeBlock(empty) should return error")
		}
	})

	t.Run("decode invalid data", func(t *testing.T) {
		_, err := encoder.DecodeBlock([]byte("invalid json"))
		if err == nil {
			t.Error("DecodeBlock(invalid) should return error")
		}
	})
}

func TestEncoder_EncodeDecodeTransaction(t *testing.T) {
	encoder := NewEncoder()
	original := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx123")
	original.BlockNumber = 12345
	original.BlockHash = "0xblock"
	original.From = "0xfrom"
	original.To = "0xto"
	original.Value = "1000000000000000000"
	original.Status = models.TxStatusSuccess

	t.Run("encode and decode", func(t *testing.T) {
		// Encode
		data, err := encoder.EncodeTransaction(original)
		if err != nil {
			t.Fatalf("EncodeTransaction() error = %v", err)
		}

		if len(data) == 0 {
			t.Error("EncodeTransaction() should return non-empty data")
		}

		// Decode
		decoded, err := encoder.DecodeTransaction(data)
		if err != nil {
			t.Fatalf("DecodeTransaction() error = %v", err)
		}

		// Verify
		if decoded.ChainType != original.ChainType {
			t.Errorf("ChainType = %v, want %v", decoded.ChainType, original.ChainType)
		}
		if decoded.ChainID != original.ChainID {
			t.Errorf("ChainID = %v, want %v", decoded.ChainID, original.ChainID)
		}
		if decoded.Hash != original.Hash {
			t.Errorf("Hash = %v, want %v", decoded.Hash, original.Hash)
		}
		if decoded.BlockNumber != original.BlockNumber {
			t.Errorf("BlockNumber = %v, want %v", decoded.BlockNumber, original.BlockNumber)
		}
		if decoded.From != original.From {
			t.Errorf("From = %v, want %v", decoded.From, original.From)
		}
		if decoded.To != original.To {
			t.Errorf("To = %v, want %v", decoded.To, original.To)
		}
	})

	t.Run("encode nil transaction", func(t *testing.T) {
		_, err := encoder.EncodeTransaction(nil)
		if err == nil {
			t.Error("EncodeTransaction(nil) should return error")
		}
	})

	t.Run("decode empty data", func(t *testing.T) {
		_, err := encoder.DecodeTransaction([]byte{})
		if err == nil {
			t.Error("DecodeTransaction(empty) should return error")
		}
	})
}

func TestEncoder_EncodeDecodeChain(t *testing.T) {
	encoder := NewEncoder()
	original := models.NewChain(models.ChainTypeEVM, "ethereum", "Ethereum Mainnet")
	original.Network = "mainnet"
	original.RPCEndpoints = []string{"http://localhost:8545"}
	original.StartBlock = 0
	original.BatchSize = 100

	t.Run("encode and decode", func(t *testing.T) {
		// Encode
		data, err := encoder.EncodeChain(original)
		if err != nil {
			t.Fatalf("EncodeChain() error = %v", err)
		}

		if len(data) == 0 {
			t.Error("EncodeChain() should return non-empty data")
		}

		// Decode
		decoded, err := encoder.DecodeChain(data)
		if err != nil {
			t.Fatalf("DecodeChain() error = %v", err)
		}

		// Verify
		if decoded.ChainType != original.ChainType {
			t.Errorf("ChainType = %v, want %v", decoded.ChainType, original.ChainType)
		}
		if decoded.ChainID != original.ChainID {
			t.Errorf("ChainID = %v, want %v", decoded.ChainID, original.ChainID)
		}
		if decoded.Name != original.Name {
			t.Errorf("Name = %v, want %v", decoded.Name, original.Name)
		}
		if decoded.Network != original.Network {
			t.Errorf("Network = %v, want %v", decoded.Network, original.Network)
		}
		if decoded.BatchSize != original.BatchSize {
			t.Errorf("BatchSize = %v, want %v", decoded.BatchSize, original.BatchSize)
		}
	})

	t.Run("encode nil chain", func(t *testing.T) {
		_, err := encoder.EncodeChain(nil)
		if err == nil {
			t.Error("EncodeChain(nil) should return error")
		}
	})

	t.Run("decode empty data", func(t *testing.T) {
		_, err := encoder.DecodeChain([]byte{})
		if err == nil {
			t.Error("DecodeChain(empty) should return error")
		}
	})
}

func TestEncoder_EncodeDecodeUint64(t *testing.T) {
	encoder := NewEncoder()

	tests := []struct {
		name  string
		value uint64
	}{
		{"zero", 0},
		{"small", 42},
		{"large", 18446744073709551615}, // max uint64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			data := encoder.EncodeUint64(tt.value)
			if len(data) == 0 {
				t.Error("EncodeUint64() should return non-empty data")
			}

			// Decode
			decoded, err := encoder.DecodeUint64(data)
			if err != nil {
				t.Fatalf("DecodeUint64() error = %v", err)
			}

			if decoded != tt.value {
				t.Errorf("DecodeUint64() = %v, want %v", decoded, tt.value)
			}
		})
	}

	t.Run("decode empty data", func(t *testing.T) {
		_, err := encoder.DecodeUint64([]byte{})
		if err == nil {
			t.Error("DecodeUint64(empty) should return error")
		}
	})

	t.Run("decode invalid data", func(t *testing.T) {
		_, err := encoder.DecodeUint64([]byte("not a number"))
		if err == nil {
			t.Error("DecodeUint64(invalid) should return error")
		}
	})
}

func TestEncoder_EncodeDecodeString(t *testing.T) {
	encoder := NewEncoder()

	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"simple", "hello"},
		{"with spaces", "hello world"},
		{"with special chars", "hello\nworld\t!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			data := encoder.EncodeString(tt.value)

			// Decode
			decoded := encoder.DecodeString(data)

			if decoded != tt.value {
				t.Errorf("DecodeString() = %v, want %v", decoded, tt.value)
			}
		})
	}
}

// Benchmark tests
func BenchmarkEncoder_EncodeBlock(b *testing.B) {
	encoder := NewEncoder()
	block := models.NewBlock(models.ChainTypeEVM, "ethereum", 12345, "0xabc123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encoder.EncodeBlock(block)
	}
}

func BenchmarkEncoder_DecodeBlock(b *testing.B) {
	encoder := NewEncoder()
	block := models.NewBlock(models.ChainTypeEVM, "ethereum", 12345, "0xabc123")
	data, _ := encoder.EncodeBlock(block)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encoder.DecodeBlock(data)
	}
}

func BenchmarkEncoder_EncodeTransaction(b *testing.B) {
	encoder := NewEncoder()
	tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encoder.EncodeTransaction(tx)
	}
}

func BenchmarkEncoder_DecodeTransaction(b *testing.B) {
	encoder := NewEncoder()
	tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx123")
	data, _ := encoder.EncodeTransaction(tx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encoder.DecodeTransaction(data)
	}
}

func BenchmarkEncoder_EncodeUint64(b *testing.B) {
	encoder := NewEncoder()
	value := uint64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = encoder.EncodeUint64(value)
	}
}

func BenchmarkEncoder_DecodeUint64(b *testing.B) {
	encoder := NewEncoder()
	data := encoder.EncodeUint64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encoder.DecodeUint64(data)
	}
}

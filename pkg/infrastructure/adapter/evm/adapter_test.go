package evm

import (
	"context"
	"testing"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxConnections <= 0 {
		t.Error("MaxConnections should be positive")
	}

	if cfg.BatchSize <= 0 {
		t.Error("BatchSize should be positive")
	}

	if cfg.ConcurrentFetches <= 0 {
		t.Error("ConcurrentFetches should be positive")
	}

	if cfg.RequestTimeout <= 0 {
		t.Error("RequestTimeout should be positive")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config with RPC",
			config: &Config{
				ChainID:           "ethereum",
				RPCEndpoints:      []string{"http://localhost:8545"},
				MaxConnections:    10,
				BatchSize:         100,
				ConcurrentFetches: 10,
				MaxRetries:        3,
				RetryBackoff:      2.0,
			},
			wantErr: false,
		},
		{
			name: "valid config with WebSocket",
			config: &Config{
				ChainID:           "ethereum",
				WSEndpoints:       []string{"ws://localhost:8546"},
				MaxConnections:    10,
				BatchSize:         100,
				ConcurrentFetches: 10,
				MaxRetries:        3,
				RetryBackoff:      2.0,
			},
			wantErr: false,
		},
		{
			name: "missing chain ID",
			config: &Config{
				RPCEndpoints:      []string{"http://localhost:8545"},
				MaxConnections:    10,
				BatchSize:         100,
				ConcurrentFetches: 10,
			},
			wantErr: true,
		},
		{
			name: "missing endpoints",
			config: &Config{
				ChainID:           "ethereum",
				MaxConnections:    10,
				BatchSize:         100,
				ConcurrentFetches: 10,
			},
			wantErr: true,
		},
		{
			name: "invalid max connections",
			config: &Config{
				ChainID:           "ethereum",
				RPCEndpoints:      []string{"http://localhost:8545"},
				MaxConnections:    0,
				BatchSize:         100,
				ConcurrentFetches: 10,
			},
			wantErr: true,
		},
		{
			name: "invalid batch size",
			config: &Config{
				ChainID:           "ethereum",
				RPCEndpoints:      []string{"http://localhost:8545"},
				MaxConnections:    10,
				BatchSize:         0,
				ConcurrentFetches: 10,
			},
			wantErr: true,
		},
		{
			name: "invalid concurrent fetches",
			config: &Config{
				ChainID:           "ethereum",
				RPCEndpoints:      []string{"http://localhost:8545"},
				MaxConnections:    10,
				BatchSize:         100,
				ConcurrentFetches: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid retry backoff",
			config: &Config{
				ChainID:           "ethereum",
				RPCEndpoints:      []string{"http://localhost:8545"},
				MaxConnections:    10,
				BatchSize:         100,
				ConcurrentFetches: 10,
				RetryBackoff:      1.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigClone(t *testing.T) {
	original := &Config{
		ChainID:           "ethereum",
		RPCEndpoints:      []string{"http://localhost:8545", "http://localhost:8546"},
		WSEndpoints:       []string{"ws://localhost:8547"},
		MaxConnections:    10,
		BatchSize:         100,
		ConcurrentFetches: 10,
	}

	clone := original.Clone()

	// Verify values are equal
	if clone.ChainID != original.ChainID {
		t.Error("ChainID not cloned correctly")
	}

	if len(clone.RPCEndpoints) != len(original.RPCEndpoints) {
		t.Error("RPCEndpoints not cloned correctly")
	}

	// Verify deep copy (modifying clone shouldn't affect original)
	clone.RPCEndpoints[0] = "http://modified:8545"
	if original.RPCEndpoints[0] == clone.RPCEndpoints[0] {
		t.Error("RPCEndpoints not deep copied")
	}
}

func TestNormalizer(t *testing.T) {
	normalizer := NewNormalizer("ethereum", "mainnet")

	if normalizer.chainID != "ethereum" {
		t.Errorf("chainID = %s, want ethereum", normalizer.chainID)
	}

	if normalizer.network != "mainnet" {
		t.Errorf("network = %s, want mainnet", normalizer.network)
	}
}

func TestParseHash(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr bool
	}{
		{
			name:    "valid hash with 0x prefix",
			hash:    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantErr: false,
		},
		{
			name:    "valid hash without 0x prefix",
			hash:    "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantErr: false,
		},
		{
			name:    "invalid hash length",
			hash:    "0x1234",
			wantErr: true,
		},
		{
			name:    "invalid hex characters",
			hash:    "0x1234567890abcdefGHIJ567890abcdef1234567890abcdef1234567890abcdef",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseHash(tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHexToByte(t *testing.T) {
	tests := []struct {
		input byte
		want  byte
	}{
		{'0', 0x00},
		{'9', 0x09},
		{'a', 0x0a},
		{'f', 0x0f},
		{'A', 0x0A},
		{'F', 0x0F},
		{'G', 0xff}, // Invalid
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := hexToByte(tt.input)
			if got != tt.want {
				t.Errorf("hexToByte(%c) = %x, want %x", tt.input, got, tt.want)
			}
		})
	}
}

func TestHexToUint64(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  uint64
	}{
		{"zero", "0x0", 0},
		{"one", "0x1", 1},
		{"hex a", "0xa", 10},
		{"hex ff", "0xff", 255},
		{"hex 100", "0x100", 256},
		{"no prefix", "ff", 255},
		{"uppercase", "0xFF", 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hexToUint64(tt.input)
			if got != tt.want {
				t.Errorf("hexToUint64(%s) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestContainsSubstring(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"found at start", "hello world", "hello", true},
		{"found at end", "hello world", "world", true},
		{"found in middle", "hello world", "lo wo", true},
		{"not found", "hello world", "xyz", false},
		{"exact match", "hello", "hello", true},
		{"empty substr", "hello", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

// Integration tests (require actual RPC endpoint)

func TestAdapterIntegration(t *testing.T) {
	// Skip if no RPC endpoint is configured
	rpcEndpoint := getTestRPCEndpoint()
	if rpcEndpoint == "" {
		t.Skip("Skipping integration test: no RPC endpoint configured")
	}

	config := &Config{
		ChainID:            "ethereum-test",
		ChainName:          "Ethereum Testnet",
		Network:            "testnet",
		RPCEndpoints:       []string{rpcEndpoint},
		MaxConnections:     1,
		ConnectionTimeout:  30 * time.Second,
		RequestTimeout:     10 * time.Second,
		BatchSize:          10,
		ConcurrentFetches:  2,
		MaxRetries:         2,
		RetryDelay:         1 * time.Second,
		RetryBackoff:       2.0,
		MaxRetryDelay:      5 * time.Second,
		EnableReceiptFetch: true,
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	defer adapter.Disconnect()

	ctx := context.Background()

	t.Run("GetChainInfo", func(t *testing.T) {
		info := adapter.GetChainInfo()
		if info == nil {
			t.Fatal("GetChainInfo() returned nil")
		}

		if info.ChainType != models.ChainTypeEVM {
			t.Errorf("ChainType = %v, want %v", info.ChainType, models.ChainTypeEVM)
		}

		if info.ChainID != config.ChainID {
			t.Errorf("ChainID = %s, want %s", info.ChainID, config.ChainID)
		}
	})

	t.Run("GetLatestBlockNumber", func(t *testing.T) {
		blockNumber, err := adapter.GetLatestBlockNumber(ctx)
		if err != nil {
			t.Fatalf("GetLatestBlockNumber() error = %v", err)
		}

		if blockNumber == 0 {
			t.Error("Expected non-zero block number")
		}
	})

	t.Run("GetBlockByNumber", func(t *testing.T) {
		// Get latest block first
		latestBlockNumber, err := adapter.GetLatestBlockNumber(ctx)
		if err != nil {
			t.Fatalf("GetLatestBlockNumber() error = %v", err)
		}

		// Fetch a recent block (latest - 10 to ensure it's confirmed)
		blockNumber := latestBlockNumber - 10
		block, err := adapter.GetBlockByNumber(ctx, blockNumber)
		if err != nil {
			t.Fatalf("GetBlockByNumber() error = %v", err)
		}

		if block.Number != blockNumber {
			t.Errorf("Block number = %d, want %d", block.Number, blockNumber)
		}

		if block.Hash == "" {
			t.Error("Block hash is empty")
		}

		if block.ChainID != config.ChainID {
			t.Errorf("ChainID = %s, want %s", block.ChainID, config.ChainID)
		}
	})

	t.Run("IsHealthy", func(t *testing.T) {
		healthy := adapter.IsHealthy(ctx)
		if !healthy {
			t.Error("Adapter should be healthy")
		}
	})
}

// Helper functions

func getTestRPCEndpoint() string {
	// Override with environment variable or return empty to skip tests
	// export TEST_EVM_RPC_ENDPOINT="https://eth.llamarpc.com"
	// For now, return empty to skip integration tests by default
	return ""
}

func BenchmarkNewNormalizer(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewNormalizer("ethereum", "mainnet")
	}
}

func BenchmarkParseHash(b *testing.B) {
	hash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseHash(hash)
	}
}

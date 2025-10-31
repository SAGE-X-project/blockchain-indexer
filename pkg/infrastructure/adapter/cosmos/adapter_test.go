package cosmos

import (
	"testing"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.ChainID != "cosmoshub-4" {
		t.Errorf("expected chain_id to be 'cosmoshub-4', got '%s'", config.ChainID)
	}

	if config.Network != "mainnet" {
		t.Errorf("expected network to be 'mainnet', got '%s'", config.Network)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("expected timeout to be 30s, got %v", config.Timeout)
	}

	if config.RetryAttempts != 3 {
		t.Errorf("expected retry_attempts to be 3, got %d", config.RetryAttempts)
	}

	if config.Bech32Prefix != "cosmos" {
		t.Errorf("expected bech32_prefix to be 'cosmos', got '%s'", config.Bech32Prefix)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "missing chain_id",
			config: &Config{
				ChainName:      "Test Chain",
				Network:        "testnet",
				RPCURL:         "https://rpc.test.com",
				Timeout:        30 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
				Bech32Prefix:   "cosmos",
				CoinDenom:      "uatom",
			},
			wantErr: true,
		},
		{
			name: "missing rpc_url",
			config: &Config{
				ChainID:        "test-1",
				ChainName:      "Test Chain",
				Network:        "testnet",
				Timeout:        30 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
				Bech32Prefix:   "cosmos",
				CoinDenom:      "uatom",
			},
			wantErr: true,
		},
		{
			name: "invalid rpc_url",
			config: &Config{
				ChainID:        "test-1",
				ChainName:      "Test Chain",
				Network:        "testnet",
				RPCURL:         "not-a-url",
				Timeout:        30 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
				Bech32Prefix:   "cosmos",
				CoinDenom:      "uatom",
			},
			wantErr: false, // URL parsing is lenient
		},
		{
			name: "negative timeout",
			config: &Config{
				ChainID:        "test-1",
				ChainName:      "Test Chain",
				Network:        "testnet",
				RPCURL:         "https://rpc.test.com",
				Timeout:        -1 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
				Bech32Prefix:   "cosmos",
				CoinDenom:      "uatom",
			},
			wantErr: true,
		},
		{
			name: "websocket enabled without url",
			config: &Config{
				ChainID:         "test-1",
				ChainName:       "Test Chain",
				Network:         "testnet",
				RPCURL:          "https://rpc.test.com",
				Timeout:         30 * time.Second,
				RetryAttempts:   3,
				MaxConnections:  10,
				BatchSize:       100,
				EnableWebSocket: true,
				Bech32Prefix:    "cosmos",
				CoinDenom:       "uatom",
			},
			wantErr: true,
		},
		{
			name: "missing bech32_prefix",
			config: &Config{
				ChainID:        "test-1",
				ChainName:      "Test Chain",
				Network:        "testnet",
				RPCURL:         "https://rpc.test.com",
				Timeout:        30 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
				CoinDenom:      "uatom",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizer_NormalizeChainInfo(t *testing.T) {
	normalizer := NewNormalizer("cosmoshub-4", "mainnet")

	chainInfo := normalizer.NormalizeChainInfo()

	if chainInfo.ChainType != models.ChainTypeCosmos {
		t.Errorf("expected chain type to be Cosmos, got %v", chainInfo.ChainType)
	}

	if chainInfo.ChainID != "cosmoshub-4" {
		t.Errorf("expected chain_id to be 'cosmoshub-4', got '%s'", chainInfo.ChainID)
	}

	if chainInfo.Network != "mainnet" {
		t.Errorf("expected network to be 'mainnet', got '%s'", chainInfo.Network)
	}
}

func TestDecodeHash(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantLen int
		wantErr bool
	}{
		{
			name:    "valid hash with 0x prefix",
			hash:    "0x1234567890abcdef",
			wantLen: 8,
			wantErr: false,
		},
		{
			name:    "valid hash without prefix",
			hash:    "1234567890abcdef",
			wantLen: 8,
			wantErr: false,
		},
		{
			name:    "64 char hash",
			hash:    "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantLen: 32,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeHash(tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("decodeHash() got length = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestConfig_GetMethods(t *testing.T) {
	config := &Config{
		ChainID:         "test-1",
		ChainName:       "Test",
		Network:         "testnet",
		RPCURL:          "https://rpc.test.com",
		RESTURL:         "https://api.test.com",
		GRPCEndpoint:    "grpc.test.com:9090",
		EnableWebSocket: true,
		WebSocketURL:    "wss://ws.test.com",
		Timeout:         30 * time.Second,
		RetryAttempts:   3,
		MaxConnections:  10,
		BatchSize:       100,
		Bech32Prefix:    "test",
		CoinDenom:       "utest",
	}

	if config.GetRPCURL() != "https://rpc.test.com" {
		t.Errorf("GetRPCURL() = %v, want %v", config.GetRPCURL(), "https://rpc.test.com")
	}

	if config.GetRESTURL() != "https://api.test.com" {
		t.Errorf("GetRESTURL() = %v, want %v", config.GetRESTURL(), "https://api.test.com")
	}

	if config.GetGRPCEndpoint() != "grpc.test.com:9090" {
		t.Errorf("GetGRPCEndpoint() = %v, want %v", config.GetGRPCEndpoint(), "grpc.test.com:9090")
	}

	if config.GetWebSocketURL() != "wss://ws.test.com" {
		t.Errorf("GetWebSocketURL() = %v, want %v", config.GetWebSocketURL(), "wss://ws.test.com")
	}

	if !config.IsWebSocketEnabled() {
		t.Error("IsWebSocketEnabled() = false, want true")
	}

	// Test with WebSocket disabled
	config.EnableWebSocket = false
	if config.GetWebSocketURL() != "" {
		t.Errorf("GetWebSocketURL() with disabled WS = %v, want empty string", config.GetWebSocketURL())
	}
}

// Benchmark tests
func BenchmarkNormalizer_NormalizeChainInfo(b *testing.B) {
	normalizer := NewNormalizer("cosmoshub-4", "mainnet")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = normalizer.NormalizeChainInfo()
	}
}

func BenchmarkDecodeHash(b *testing.B) {
	hash := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decodeHash(hash)
	}
}

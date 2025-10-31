package polkadot

import (
	"testing"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.ChainID != "polkadot" {
		t.Errorf("expected chain_id to be 'polkadot', got '%s'", config.ChainID)
	}

	if config.Network != "mainnet" {
		t.Errorf("expected network to be 'mainnet', got '%s'", config.Network)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("expected timeout to be 30s, got %v", config.Timeout)
	}

	if config.SS58Format != 0 {
		t.Errorf("expected ss58_format to be 0, got %d", config.SS58Format)
	}

	if config.TokenSymbol != "DOT" {
		t.Errorf("expected token_symbol to be 'DOT', got '%s'", config.TokenSymbol)
	}
}

func TestKusamaConfig(t *testing.T) {
	config := KusamaConfig()

	if config.ChainID != "kusama" {
		t.Errorf("expected chain_id to be 'kusama', got '%s'", config.ChainID)
	}

	if config.SS58Format != 2 {
		t.Errorf("expected ss58_format to be 2, got %d", config.SS58Format)
	}

	if config.TokenSymbol != "KSM" {
		t.Errorf("expected token_symbol to be 'KSM', got '%s'", config.TokenSymbol)
	}

	if config.TokenDecimals != 12 {
		t.Errorf("expected token_decimals to be 12, got %d", config.TokenDecimals)
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
				TokenSymbol:    "TEST",
			},
			wantErr: true,
		},
		{
			name: "missing rpc_url",
			config: &Config{
				ChainID:        "test",
				ChainName:      "Test Chain",
				Network:        "testnet",
				Timeout:        30 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
				TokenSymbol:    "TEST",
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: &Config{
				ChainID:        "test",
				ChainName:      "Test Chain",
				Network:        "testnet",
				RPCURL:         "https://rpc.test.com",
				Timeout:        -1 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
				TokenSymbol:    "TEST",
			},
			wantErr: true,
		},
		{
			name: "websocket enabled without url",
			config: &Config{
				ChainID:         "test",
				ChainName:       "Test Chain",
				Network:         "testnet",
				RPCURL:          "https://rpc.test.com",
				Timeout:         30 * time.Second,
				RetryAttempts:   3,
				MaxConnections:  10,
				BatchSize:       100,
				EnableWebSocket: true,
				TokenSymbol:     "TEST",
			},
			wantErr: true,
		},
		{
			name: "missing token_symbol",
			config: &Config{
				ChainID:        "test",
				ChainName:      "Test Chain",
				Network:        "testnet",
				RPCURL:         "https://rpc.test.com",
				Timeout:        30 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
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
	normalizer := NewNormalizer("polkadot", "mainnet")

	chainInfo := normalizer.NormalizeChainInfo()

	if chainInfo.ChainType != models.ChainTypePolkadot {
		t.Errorf("expected chain type to be Polkadot, got %v", chainInfo.ChainType)
	}

	if chainInfo.ChainID != "polkadot" {
		t.Errorf("expected chain_id to be 'polkadot', got '%s'", chainInfo.ChainID)
	}

	if chainInfo.Network != "mainnet" {
		t.Errorf("expected network to be 'mainnet', got '%s'", chainInfo.Network)
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
			hash:    "0x" + "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantErr: false,
		},
		{
			name:    "valid hash without prefix",
			hash:    "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantErr: false,
		},
		{
			name:    "invalid hash length",
			hash:    "1234",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			hash:    "zzzz567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
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

func TestConfig_GetMethods(t *testing.T) {
	config := &Config{
		ChainID:         "test",
		ChainName:       "Test",
		Network:         "testnet",
		RPCURL:          "https://rpc.test.com",
		WSURL:           "wss://ws.test.com",
		EnableWebSocket: true,
		Timeout:         30 * time.Second,
		RetryAttempts:   3,
		MaxConnections:  10,
		BatchSize:       100,
		TokenSymbol:     "TEST",
	}

	if config.GetRPCURL() != "https://rpc.test.com" {
		t.Errorf("GetRPCURL() = %v, want %v", config.GetRPCURL(), "https://rpc.test.com")
	}

	if config.GetWSURL() != "wss://ws.test.com" {
		t.Errorf("GetWSURL() = %v, want %v", config.GetWSURL(), "wss://ws.test.com")
	}

	if !config.IsWebSocketEnabled() {
		t.Error("IsWebSocketEnabled() = false, want true")
	}

	// Test with WebSocket disabled
	config.EnableWebSocket = false
	if config.IsWebSocketEnabled() {
		t.Error("IsWebSocketEnabled() = true, want false")
	}
}

// Benchmark tests
func BenchmarkNormalizer_NormalizeChainInfo(b *testing.B) {
	normalizer := NewNormalizer("polkadot", "mainnet")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = normalizer.NormalizeChainInfo()
	}
}

func BenchmarkParseHash(b *testing.B) {
	hash := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseHash(hash)
	}
}

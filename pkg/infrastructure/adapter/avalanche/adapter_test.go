package avalanche

import (
	"testing"
	"time"
)

func TestDefaultCChainConfig(t *testing.T) {
	config := DefaultCChainConfig()

	if config.ChainID != "avalanche-c" {
		t.Errorf("expected chain_id to be 'avalanche-c', got '%s'", config.ChainID)
	}

	if config.ChainType != CChain {
		t.Errorf("expected chain_type to be C-Chain, got '%s'", config.ChainType)
	}

	if config.Network != "mainnet" {
		t.Errorf("expected network to be 'mainnet', got '%s'", config.Network)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("expected timeout to be 30s, got %v", config.Timeout)
	}
}

func TestFujiCChainConfig(t *testing.T) {
	config := FujiCChainConfig()

	if config.ChainID != "avalanche-fuji-c" {
		t.Errorf("expected chain_id to be 'avalanche-fuji-c', got '%s'", config.ChainID)
	}

	if config.Network != "testnet" {
		t.Errorf("expected network to be 'testnet', got '%s'", config.Network)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid C-Chain config",
			config:  DefaultCChainConfig(),
			wantErr: false,
		},
		{
			name: "missing chain_id",
			config: &Config{
				ChainName:      "Test",
				Network:        "testnet",
				ChainType:      CChain,
				RPCURL:         "https://test.com",
				Timeout:        30 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
			},
			wantErr: true,
		},
		{
			name: "missing chain_type",
			config: &Config{
				ChainID:        "test",
				ChainName:      "Test",
				Network:        "testnet",
				RPCURL:         "https://test.com",
				Timeout:        30 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
			},
			wantErr: true,
		},
		{
			name: "invalid chain_type",
			config: &Config{
				ChainID:        "test",
				ChainName:      "Test",
				Network:        "testnet",
				ChainType:      "invalid",
				RPCURL:         "https://test.com",
				Timeout:        30 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
			},
			wantErr: true,
		},
		{
			name: "missing rpc_url",
			config: &Config{
				ChainID:        "test",
				ChainName:      "Test",
				Network:        "testnet",
				ChainType:      CChain,
				Timeout:        30 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
			},
			wantErr: true,
		},
		{
			name: "websocket enabled without url",
			config: &Config{
				ChainID:         "test",
				ChainName:       "Test",
				Network:         "testnet",
				ChainType:       CChain,
				RPCURL:          "https://test.com",
				Timeout:         30 * time.Second,
				RetryAttempts:   3,
				MaxConnections:  10,
				BatchSize:       100,
				EnableWebSocket: true,
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

func TestConfig_GetMethods(t *testing.T) {
	config := &Config{
		ChainID:         "test",
		ChainName:       "Test",
		Network:         "testnet",
		ChainType:       CChain,
		RPCURL:          "https://rpc.test.com",
		WSURL:           "wss://ws.test.com",
		EnableWebSocket: true,
		Timeout:         30 * time.Second,
		RetryAttempts:   3,
		MaxConnections:  10,
		BatchSize:       100,
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

	if !config.IsCChain() {
		t.Error("IsCChain() = false, want true")
	}

	if config.IsXChain() {
		t.Error("IsXChain() = true, want false")
	}

	if config.IsPChain() {
		t.Error("IsPChain() = true, want false")
	}
}

func TestChainTypeChecks(t *testing.T) {
	tests := []struct {
		name      string
		chainType ChainType
		isCChain  bool
		isXChain  bool
		isPChain  bool
	}{
		{
			name:      "C-Chain",
			chainType: CChain,
			isCChain:  true,
			isXChain:  false,
			isPChain:  false,
		},
		{
			name:      "X-Chain",
			chainType: XChain,
			isCChain:  false,
			isXChain:  true,
			isPChain:  false,
		},
		{
			name:      "P-Chain",
			chainType: PChain,
			isCChain:  false,
			isXChain:  false,
			isPChain:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				ChainID:        "test",
				ChainName:      "Test",
				Network:        "testnet",
				ChainType:      tt.chainType,
				RPCURL:         "https://test.com",
				Timeout:        30 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
			}

			if config.IsCChain() != tt.isCChain {
				t.Errorf("IsCChain() = %v, want %v", config.IsCChain(), tt.isCChain)
			}

			if config.IsXChain() != tt.isXChain {
				t.Errorf("IsXChain() = %v, want %v", config.IsXChain(), tt.isXChain)
			}

			if config.IsPChain() != tt.isPChain {
				t.Errorf("IsPChain() = %v, want %v", config.IsPChain(), tt.isPChain)
			}
		})
	}
}

func TestAdapter_GetChainType(t *testing.T) {
	// Note: This test will fail without a real RPC endpoint
	// but we can test the chain type
	config := DefaultCChainConfig()
	config.RPCURL = "http://localhost:9650/ext/bc/C/rpc" // Local endpoint

	// We can't create the adapter without a real endpoint,
	// so we just verify the config is correct
	if config.ChainType != CChain {
		t.Errorf("ChainType = %v, want %v", config.ChainType, CChain)
	}
}

// Benchmark tests
func BenchmarkConfig_Validate(b *testing.B) {
	config := DefaultCChainConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

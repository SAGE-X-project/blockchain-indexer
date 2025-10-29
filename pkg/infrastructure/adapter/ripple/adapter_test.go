package ripple

import (
	"testing"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.ChainID != "xrpl-mainnet" {
		t.Errorf("expected chain_id to be 'xrpl-mainnet', got '%s'", config.ChainID)
	}

	if config.Network != "mainnet" {
		t.Errorf("expected network to be 'mainnet', got '%s'", config.Network)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("expected timeout to be 30s, got %v", config.Timeout)
	}

	if !config.IncludeTransactionMetadata {
		t.Error("expected include_transaction_metadata to be true")
	}
}

func TestTestnetConfig(t *testing.T) {
	config := TestnetConfig()

	if config.ChainID != "xrpl-testnet" {
		t.Errorf("expected chain_id to be 'xrpl-testnet', got '%s'", config.ChainID)
	}

	if config.Network != "testnet" {
		t.Errorf("expected network to be 'testnet', got '%s'", config.Network)
	}
}

func TestDevnetConfig(t *testing.T) {
	config := DevnetConfig()

	if config.ChainID != "xrpl-devnet" {
		t.Errorf("expected chain_id to be 'xrpl-devnet', got '%s'", config.ChainID)
	}

	if config.Network != "devnet" {
		t.Errorf("expected network to be 'devnet', got '%s'", config.Network)
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
			name: "missing rpc_url",
			config: &Config{
				ChainID:        "test",
				ChainName:      "Test",
				Network:        "testnet",
				Timeout:        30 * time.Second,
				RetryAttempts:  3,
				MaxConnections: 10,
				BatchSize:      100,
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: &Config{
				ChainID:        "test",
				ChainName:      "Test",
				Network:        "testnet",
				RPCURL:         "https://test.com",
				Timeout:        -1 * time.Second,
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
		RPCURL:          "https://rpc.test.com",
		WebSocketURL:    "wss://ws.test.com",
		EnableWebSocket: true,
		Timeout:         30 * time.Second,
		RetryAttempts:   3,
		MaxConnections:  10,
		BatchSize:       100,
	}

	if config.GetRPCURL() != "https://rpc.test.com" {
		t.Errorf("GetRPCURL() = %v, want %v", config.GetRPCURL(), "https://rpc.test.com")
	}

	if config.GetWebSocketURL() != "wss://ws.test.com" {
		t.Errorf("GetWebSocketURL() = %v, want %v", config.GetWebSocketURL(), "wss://ws.test.com")
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

func TestNewAdapter(t *testing.T) {
	config := DefaultConfig()

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter() error = %v", err)
	}

	if adapter.GetChainType() != models.ChainTypeRipple {
		t.Errorf("GetChainType() = %v, want %v", adapter.GetChainType(), models.ChainTypeRipple)
	}

	if adapter.GetChainID() != config.ChainID {
		t.Errorf("GetChainID() = %v, want %v", adapter.GetChainID(), config.ChainID)
	}

	chainInfo := adapter.GetChainInfo()
	if chainInfo.ChainType != models.ChainTypeRipple {
		t.Errorf("ChainInfo.ChainType = %v, want %v", chainInfo.ChainType, models.ChainTypeRipple)
	}

	if chainInfo.Network != config.Network {
		t.Errorf("ChainInfo.Network = %v, want %v", chainInfo.Network, config.Network)
	}
}

func TestAdapter_ConnectDisconnect(t *testing.T) {
	config := DefaultConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter() error = %v", err)
	}

	// Initially not connected
	if adapter.IsConnected() {
		t.Error("IsConnected() = true, want false initially")
	}

	// Connect
	if err := adapter.Connect(nil); err != nil {
		t.Errorf("Connect() error = %v", err)
	}

	if !adapter.IsConnected() {
		t.Error("IsConnected() = false, want true after Connect()")
	}

	// Disconnect
	if err := adapter.Disconnect(); err != nil {
		t.Errorf("Disconnect() error = %v", err)
	}

	if adapter.IsConnected() {
		t.Error("IsConnected() = true, want false after Disconnect()")
	}
}

// Benchmark tests
func BenchmarkConfig_Validate(b *testing.B) {
	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkNewAdapter(b *testing.B) {
	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewAdapter(config)
	}
}

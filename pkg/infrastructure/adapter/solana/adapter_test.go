package solana

import (
	"context"
	"testing"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

func TestNewAdapter(t *testing.T) {
	config := DefaultConfig("solana-devnet", "devnet", "https://api.devnet.solana.com")

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("Adapter is nil")
	}

	if adapter.GetChainType() != models.ChainTypeSolana {
		t.Errorf("Expected chain type Solana, got %s", adapter.GetChainType())
	}

	if adapter.GetChainID() != "solana-devnet" {
		t.Errorf("Expected chain ID solana-devnet, got %s", adapter.GetChainID())
	}
}

func TestAdapterChainInfo(t *testing.T) {
	config := DefaultConfig("solana-devnet", "devnet", "https://api.devnet.solana.com")

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	chainInfo := adapter.GetChainInfo()
	if chainInfo == nil {
		t.Fatal("Chain info is nil")
	}

	if chainInfo.ChainType != models.ChainTypeSolana {
		t.Errorf("Expected chain type Solana, got %s", chainInfo.ChainType)
	}

	if chainInfo.ChainID != "solana-devnet" {
		t.Errorf("Expected chain ID solana-devnet, got %s", chainInfo.ChainID)
	}

	if chainInfo.Network != "devnet" {
		t.Errorf("Expected network devnet, got %s", chainInfo.Network)
	}
}

func TestAdapterConnection(t *testing.T) {
	config := DefaultConfig("solana-devnet", "devnet", "https://api.devnet.solana.com")

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test connection
	err = adapter.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Test health check
	if !adapter.IsHealthy(ctx) {
		t.Error("Adapter is not healthy after connection")
	}

	// Test disconnect
	err = adapter.Disconnect()
	if err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}
}

func TestGetLatestBlockNumber(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig("solana-devnet", "devnet", "https://api.devnet.solana.com")
	config.RequestTimeout = 30 * time.Second

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slot, err := adapter.GetLatestBlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get latest block number: %v", err)
	}

	if slot == 0 {
		t.Error("Expected non-zero slot number")
	}

	t.Logf("Latest slot: %d", slot)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig("solana-devnet", "devnet", "https://api.devnet.solana.com"),
			wantErr: false,
		},
		{
			name: "missing chain_id",
			config: &Config{
				ChainName:   "Solana",
				Network:     "devnet",
				RPCEndpoint: "https://api.devnet.solana.com",
			},
			wantErr: true,
		},
		{
			name: "missing rpc_endpoint",
			config: &Config{
				ChainID:   "solana-devnet",
				ChainName: "Solana",
				Network:   "devnet",
			},
			wantErr: true,
		},
		{
			name: "invalid transaction details",
			config: &Config{
				ChainID:            "solana-devnet",
				ChainName:          "Solana",
				Network:            "devnet",
				RPCEndpoint:        "https://api.devnet.solana.com",
				TransactionDetails: "invalid",
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

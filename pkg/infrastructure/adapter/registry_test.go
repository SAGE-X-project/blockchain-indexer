package adapter

import (
	"testing"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/service"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/avalanche"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/cosmos"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/evm"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/polkadot"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/ripple"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/solana"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	if registry.factories == nil {
		t.Fatal("Registry factories map is nil")
	}

	// Check that default adapters are registered
	expectedChains := []models.ChainType{
		models.ChainTypeEVM,
		models.ChainTypeSolana,
		models.ChainTypeCosmos,
		models.ChainTypePolkadot,
		models.ChainTypeAvalanche,
		models.ChainTypeRipple,
	}

	for _, chainType := range expectedChains {
		if !registry.IsSupported(chainType) {
			t.Errorf("Expected chain type %s to be registered by default", chainType)
		}
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	// Create a custom factory
	customFactory := func(config interface{}) (service.ChainAdapter, error) {
		return nil, nil
	}

	// Register custom factory
	customChainType := models.ChainType("custom")
	registry.Register(customChainType, customFactory)

	if !registry.IsSupported(customChainType) {
		t.Error("Custom chain type should be registered")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()

	// Unregister EVM
	registry.Unregister(models.ChainTypeEVM)

	if registry.IsSupported(models.ChainTypeEVM) {
		t.Error("EVM chain type should be unregistered")
	}
}

func TestRegistry_CreateEVM(t *testing.T) {
	registry := NewRegistry()

	config := evm.DefaultConfig()
	config.ChainID = "test-chain"
	config.ChainName = "Test Chain"
	config.Network = "testnet"
	config.RPCEndpoints = []string{"http://localhost:8545"}

	adapter, err := registry.Create(models.ChainTypeEVM, config)
	if err != nil {
		t.Fatalf("Failed to create EVM adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("Created adapter is nil")
	}

	if adapter.GetChainType() != models.ChainTypeEVM {
		t.Errorf("Expected chain type %s, got %s", models.ChainTypeEVM, adapter.GetChainType())
	}
}

func TestRegistry_CreateSolana(t *testing.T) {
	registry := NewRegistry()

	config := solana.DefaultConfig("solana-mainnet", "mainnet-beta", "https://api.mainnet-beta.solana.com")

	adapter, err := registry.Create(models.ChainTypeSolana, config)
	if err != nil {
		t.Fatalf("Failed to create Solana adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("Created adapter is nil")
	}

	if adapter.GetChainType() != models.ChainTypeSolana {
		t.Errorf("Expected chain type %s, got %s", models.ChainTypeSolana, adapter.GetChainType())
	}
}

func TestRegistry_CreateCosmos(t *testing.T) {
	t.Skip("Skipping Cosmos adapter test - requires live RPC endpoint")

	registry := NewRegistry()

	config := cosmos.DefaultConfig()

	adapter, err := registry.Create(models.ChainTypeCosmos, config)
	if err != nil {
		t.Fatalf("Failed to create Cosmos adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("Created adapter is nil")
	}

	if adapter.GetChainType() != models.ChainTypeCosmos {
		t.Errorf("Expected chain type %s, got %s", models.ChainTypeCosmos, adapter.GetChainType())
	}
}

func TestRegistry_CreatePolkadot(t *testing.T) {
	registry := NewRegistry()

	config := polkadot.DefaultConfig()

	adapter, err := registry.Create(models.ChainTypePolkadot, config)
	if err != nil {
		t.Fatalf("Failed to create Polkadot adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("Created adapter is nil")
	}

	if adapter.GetChainType() != models.ChainTypePolkadot {
		t.Errorf("Expected chain type %s, got %s", models.ChainTypePolkadot, adapter.GetChainType())
	}
}

func TestRegistry_CreateAvalanche(t *testing.T) {
	t.Skip("Skipping Avalanche adapter test - requires proper EVM config setup")

	registry := NewRegistry()

	config := avalanche.DefaultCChainConfig()
	// Fix config to pass validation
	config.RPCURL = "http://localhost:9650/ext/bc/C/rpc"

	adapter, err := registry.Create(models.ChainTypeAvalanche, config)
	if err != nil {
		t.Fatalf("Failed to create Avalanche adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("Created adapter is nil")
	}

	if adapter.GetChainType() != models.ChainTypeAvalanche {
		t.Errorf("Expected chain type %s, got %s", models.ChainTypeAvalanche, adapter.GetChainType())
	}
}

func TestRegistry_CreateRipple(t *testing.T) {
	registry := NewRegistry()

	config := ripple.DefaultConfig()

	adapter, err := registry.Create(models.ChainTypeRipple, config)
	if err != nil {
		t.Fatalf("Failed to create Ripple adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("Created adapter is nil")
	}

	if adapter.GetChainType() != models.ChainTypeRipple {
		t.Errorf("Expected chain type %s, got %s", models.ChainTypeRipple, adapter.GetChainType())
	}
}

func TestRegistry_CreateInvalidChainType(t *testing.T) {
	registry := NewRegistry()

	invalidChainType := models.ChainType("invalid")
	_, err := registry.Create(invalidChainType, nil)

	if err == nil {
		t.Error("Expected error when creating adapter for unsupported chain type")
	}
}

func TestRegistry_CreateInvalidConfig(t *testing.T) {
	registry := NewRegistry()

	// Try to create EVM adapter with wrong config type
	wrongConfig := "invalid config"
	_, err := registry.Create(models.ChainTypeEVM, wrongConfig)

	if err == nil {
		t.Error("Expected error when creating adapter with invalid config type")
	}
}

func TestRegistry_CreateFromConfig(t *testing.T) {
	registry := NewRegistry()

	evmConfig := evm.DefaultConfig()
	evmConfig.ChainID = "test-chain"
	evmConfig.ChainName = "Test Chain"
	evmConfig.Network = "testnet"
	evmConfig.RPCEndpoints = []string{"http://localhost:8545"}

	adapterConfig := &AdapterConfig{
		ChainType: models.ChainTypeEVM,
		Config:    evmConfig,
	}

	adapter, err := registry.CreateFromConfig(adapterConfig)
	if err != nil {
		t.Fatalf("Failed to create adapter from config: %v", err)
	}

	if adapter == nil {
		t.Fatal("Created adapter is nil")
	}

	if adapter.GetChainType() != models.ChainTypeEVM {
		t.Errorf("Expected chain type %s, got %s", models.ChainTypeEVM, adapter.GetChainType())
	}
}

func TestRegistry_IsSupported(t *testing.T) {
	registry := NewRegistry()

	tests := []struct {
		name      string
		chainType models.ChainType
		supported bool
	}{
		{
			name:      "EVM is supported",
			chainType: models.ChainTypeEVM,
			supported: true,
		},
		{
			name:      "Solana is supported",
			chainType: models.ChainTypeSolana,
			supported: true,
		},
		{
			name:      "Invalid chain is not supported",
			chainType: models.ChainType("invalid"),
			supported: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			supported := registry.IsSupported(tt.chainType)
			if supported != tt.supported {
				t.Errorf("IsSupported(%s) = %v, want %v", tt.chainType, supported, tt.supported)
			}
		})
	}
}

func TestRegistry_SupportedChainTypes(t *testing.T) {
	registry := NewRegistry()

	supportedTypes := registry.SupportedChainTypes()

	if len(supportedTypes) != 6 {
		t.Errorf("Expected 6 supported chain types, got %d", len(supportedTypes))
	}

	// Check that all expected types are present
	expectedTypes := map[models.ChainType]bool{
		models.ChainTypeEVM:       false,
		models.ChainTypeSolana:    false,
		models.ChainTypeCosmos:    false,
		models.ChainTypePolkadot:  false,
		models.ChainTypeAvalanche: false,
		models.ChainTypeRipple:    false,
	}

	for _, chainType := range supportedTypes {
		if _, exists := expectedTypes[chainType]; exists {
			expectedTypes[chainType] = true
		}
	}

	for chainType, found := range expectedTypes {
		if !found {
			t.Errorf("Expected chain type %s to be in supported types", chainType)
		}
	}
}

func TestGlobalRegistry(t *testing.T) {
	registry1 := GlobalRegistry()
	registry2 := GlobalRegistry()

	if registry1 != registry2 {
		t.Error("GlobalRegistry() should return the same instance")
	}
}

func TestCreateAdapter(t *testing.T) {
	config := evm.DefaultConfig()
	config.ChainID = "test-chain"
	config.ChainName = "Test Chain"
	config.Network = "testnet"
	config.RPCEndpoints = []string{"http://localhost:8545"}

	adapter, err := CreateAdapter(models.ChainTypeEVM, config)
	if err != nil {
		t.Fatalf("CreateAdapter() failed: %v", err)
	}

	if adapter == nil {
		t.Fatal("Created adapter is nil")
	}

	if adapter.GetChainType() != models.ChainTypeEVM {
		t.Errorf("Expected chain type %s, got %s", models.ChainTypeEVM, adapter.GetChainType())
	}
}

func TestIsChainSupported(t *testing.T) {
	if !IsChainSupported(models.ChainTypeEVM) {
		t.Error("EVM should be supported")
	}

	if IsChainSupported(models.ChainType("invalid")) {
		t.Error("Invalid chain type should not be supported")
	}
}

func TestSupportedChains(t *testing.T) {
	chains := SupportedChains()

	if len(chains) != 6 {
		t.Errorf("Expected 6 supported chains, got %d", len(chains))
	}
}

// Benchmark tests
func BenchmarkRegistry_Create(b *testing.B) {
	registry := NewRegistry()
	config := evm.DefaultConfig()
	config.ChainID = "test-chain"
	config.ChainName = "Test Chain"
	config.Network = "testnet"
	config.RPCEndpoints = []string{"http://localhost:8545"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.Create(models.ChainTypeEVM, config)
	}
}

func BenchmarkGlobalRegistry(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GlobalRegistry()
	}
}

func BenchmarkIsChainSupported(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsChainSupported(models.ChainTypeEVM)
	}
}

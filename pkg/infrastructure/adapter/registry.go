package adapter

import (
	"fmt"
	"sync"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/service"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/avalanche"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/cosmos"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/evm"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/polkadot"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/ripple"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/solana"
)

// AdapterConfig represents the configuration for creating a chain adapter
type AdapterConfig struct {
	ChainType models.ChainType
	Config    interface{} // Chain-specific config
}

// AdapterFactory is a function that creates a chain adapter
type AdapterFactory func(config interface{}) (service.ChainAdapter, error)

// Registry manages chain adapter creation and registration
type Registry struct {
	mu        sync.RWMutex
	factories map[models.ChainType]AdapterFactory
}

// NewRegistry creates a new adapter registry with default adapters registered
func NewRegistry() *Registry {
	r := &Registry{
		factories: make(map[models.ChainType]AdapterFactory),
	}

	// Register default adapters
	r.registerDefaults()

	return r
}

// registerDefaults registers all built-in chain adapters
func (r *Registry) registerDefaults() {
	// EVM adapter
	r.Register(models.ChainTypeEVM, func(config interface{}) (service.ChainAdapter, error) {
		evmConfig, ok := config.(*evm.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for EVM adapter, expected *evm.Config")
		}
		return evm.NewAdapter(evmConfig)
	})

	// Solana adapter
	r.Register(models.ChainTypeSolana, func(config interface{}) (service.ChainAdapter, error) {
		solanaConfig, ok := config.(*solana.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for Solana adapter, expected *solana.Config")
		}
		return solana.NewAdapter(solanaConfig)
	})

	// Cosmos adapter
	r.Register(models.ChainTypeCosmos, func(config interface{}) (service.ChainAdapter, error) {
		cosmosConfig, ok := config.(*cosmos.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for Cosmos adapter, expected *cosmos.Config")
		}
		return cosmos.NewAdapter(cosmosConfig)
	})

	// Polkadot adapter
	r.Register(models.ChainTypePolkadot, func(config interface{}) (service.ChainAdapter, error) {
		polkadotConfig, ok := config.(*polkadot.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for Polkadot adapter, expected *polkadot.Config")
		}
		return polkadot.NewAdapter(polkadotConfig)
	})

	// Avalanche adapter
	r.Register(models.ChainTypeAvalanche, func(config interface{}) (service.ChainAdapter, error) {
		avalancheConfig, ok := config.(*avalanche.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for Avalanche adapter, expected *avalanche.Config")
		}
		return avalanche.NewAdapter(avalancheConfig)
	})

	// Ripple adapter
	r.Register(models.ChainTypeRipple, func(config interface{}) (service.ChainAdapter, error) {
		rippleConfig, ok := config.(*ripple.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for Ripple adapter, expected *ripple.Config")
		}
		return ripple.NewAdapter(rippleConfig)
	})
}

// Register registers a new adapter factory for a chain type
func (r *Registry) Register(chainType models.ChainType, factory AdapterFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[chainType] = factory
}

// Unregister removes an adapter factory for a chain type
func (r *Registry) Unregister(chainType models.ChainType) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.factories, chainType)
}

// Create creates a chain adapter for the given chain type and config
func (r *Registry) Create(chainType models.ChainType, config interface{}) (service.ChainAdapter, error) {
	r.mu.RLock()
	factory, exists := r.factories[chainType]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no adapter registered for chain type: %s", chainType)
	}

	return factory(config)
}

// CreateFromConfig creates a chain adapter from an AdapterConfig
func (r *Registry) CreateFromConfig(config *AdapterConfig) (service.ChainAdapter, error) {
	return r.Create(config.ChainType, config.Config)
}

// IsSupported checks if a chain type is supported
func (r *Registry) IsSupported(chainType models.ChainType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.factories[chainType]
	return exists
}

// SupportedChainTypes returns a list of all supported chain types
func (r *Registry) SupportedChainTypes() []models.ChainType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]models.ChainType, 0, len(r.factories))
	for chainType := range r.factories {
		types = append(types, chainType)
	}
	return types
}

// Global registry instance
var (
	globalRegistry     *Registry
	globalRegistryOnce sync.Once
)

// GlobalRegistry returns the global adapter registry
func GlobalRegistry() *Registry {
	globalRegistryOnce.Do(func() {
		globalRegistry = NewRegistry()
	})
	return globalRegistry
}

// CreateAdapter creates a chain adapter using the global registry
func CreateAdapter(chainType models.ChainType, config interface{}) (service.ChainAdapter, error) {
	return GlobalRegistry().Create(chainType, config)
}

// IsChainSupported checks if a chain type is supported using the global registry
func IsChainSupported(chainType models.ChainType) bool {
	return GlobalRegistry().IsSupported(chainType)
}

// SupportedChains returns a list of all supported chain types using the global registry
func SupportedChains() []models.ChainType {
	return GlobalRegistry().SupportedChainTypes()
}

package cmd

import (
	"fmt"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/service"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/avalanche"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/cosmos"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/evm"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/polkadot"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/ripple"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/adapter/solana"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/config"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
)

// CreateChainAdapter creates a chain adapter based on the chain configuration
func CreateChainAdapter(chainCfg *config.ChainConfig, log *logger.Logger) (service.ChainAdapter, error) {
	if chainCfg == nil {
		return nil, fmt.Errorf("chain config is nil")
	}

	if len(chainCfg.RPCEndpoints) == 0 {
		return nil, fmt.Errorf("no RPC endpoints configured for chain %s", chainCfg.ChainID)
	}

	// Parse retry delay
	retryDelay, err := time.ParseDuration(chainCfg.RetryDelay)
	if err != nil {
		return nil, fmt.Errorf("invalid retry delay for chain %s: %w", chainCfg.ChainID, err)
	}

	// Create adapter based on chain type
	switch chainCfg.ChainType {
	case "evm", "ethereum", "bsc", "polygon":
		return createEVMAdapter(chainCfg, retryDelay, log)
	case "solana":
		return createSolanaAdapter(chainCfg, retryDelay, log)
	case "cosmos":
		return createCosmosAdapter(chainCfg, retryDelay, log)
	case "polkadot", "substrate":
		return createPolkadotAdapter(chainCfg, retryDelay, log)
	case "avalanche":
		return createAvalancheAdapter(chainCfg, retryDelay, log)
	case "ripple", "xrpl":
		return createRippleAdapter(chainCfg, retryDelay, log)
	default:
		return nil, fmt.Errorf("unsupported chain type: %s", chainCfg.ChainType)
	}
}

func createEVMAdapter(chainCfg *config.ChainConfig, retryDelay time.Duration, log *logger.Logger) (service.ChainAdapter, error) {
	adapterCfg := evm.DefaultConfig()
	adapterCfg.ChainID = chainCfg.ChainID
	adapterCfg.ChainName = chainCfg.Name
	adapterCfg.Network = chainCfg.Network
	adapterCfg.RPCEndpoints = chainCfg.RPCEndpoints
	adapterCfg.WSEndpoints = chainCfg.WSEndpoints
	adapterCfg.MaxRetries = chainCfg.RetryAttempts
	adapterCfg.RetryDelay = retryDelay
	adapterCfg.BlockConfirmations = chainCfg.ConfirmationBlocks
	adapterCfg.BatchSize = chainCfg.BatchSize
	adapterCfg.ConcurrentFetches = chainCfg.Workers

	return evm.NewAdapter(adapterCfg)
}

func createSolanaAdapter(chainCfg *config.ChainConfig, retryDelay time.Duration, log *logger.Logger) (service.ChainAdapter, error) {
	rpcEndpoint := chainCfg.RPCEndpoints[0]
	adapterCfg := solana.DefaultConfig(chainCfg.ChainID, chainCfg.Network, rpcEndpoint)

	if len(chainCfg.WSEndpoints) > 0 {
		adapterCfg.WSEndpoint = chainCfg.WSEndpoints[0]
	}
	adapterCfg.MaxRetries = chainCfg.RetryAttempts
	adapterCfg.RetryDelay = retryDelay

	return solana.NewAdapter(adapterCfg)
}

func createCosmosAdapter(chainCfg *config.ChainConfig, retryDelay time.Duration, log *logger.Logger) (service.ChainAdapter, error) {
	adapterCfg := cosmos.DefaultConfig()
	adapterCfg.ChainID = chainCfg.ChainID
	adapterCfg.ChainName = chainCfg.Name
	adapterCfg.Network = chainCfg.Network
	adapterCfg.RPCURL = chainCfg.RPCEndpoints[0]
	adapterCfg.RetryDelay = retryDelay

	return cosmos.NewAdapter(adapterCfg)
}

func createPolkadotAdapter(chainCfg *config.ChainConfig, retryDelay time.Duration, log *logger.Logger) (service.ChainAdapter, error) {
	adapterCfg := polkadot.DefaultConfig()
	adapterCfg.ChainID = chainCfg.ChainID
	adapterCfg.ChainName = chainCfg.Name
	adapterCfg.Network = chainCfg.Network
	adapterCfg.RPCURL = chainCfg.RPCEndpoints[0]
	if len(chainCfg.WSEndpoints) > 0 {
		adapterCfg.WSURL = chainCfg.WSEndpoints[0]
	}
	adapterCfg.RetryAttempts = chainCfg.RetryAttempts
	adapterCfg.RetryDelay = retryDelay

	return polkadot.NewAdapter(adapterCfg)
}

func createAvalancheAdapter(chainCfg *config.ChainConfig, retryDelay time.Duration, log *logger.Logger) (service.ChainAdapter, error) {
	// Avalanche doesn't have DefaultConfig, create manually
	adapterCfg := &avalanche.Config{
		ChainID:   chainCfg.ChainID,
		ChainName: chainCfg.Name,
		Network:   chainCfg.Network,
		ChainType: avalanche.CChain, // Default to C-Chain (EVM compatible)
		RPCURL:    chainCfg.RPCEndpoints[0],
		Timeout:   30 * time.Second,
	}

	if len(chainCfg.WSEndpoints) > 0 {
		adapterCfg.WSURL = chainCfg.WSEndpoints[0]
	}

	return avalanche.NewAdapter(adapterCfg)
}

func createRippleAdapter(chainCfg *config.ChainConfig, retryDelay time.Duration, log *logger.Logger) (service.ChainAdapter, error) {
	adapterCfg := ripple.DefaultConfig()
	adapterCfg.ChainID = chainCfg.ChainID
	adapterCfg.ChainName = chainCfg.Name
	adapterCfg.Network = chainCfg.Network
	adapterCfg.RPCURL = chainCfg.RPCEndpoints[0]
	if len(chainCfg.WSEndpoints) > 0 {
		adapterCfg.WebSocketURL = chainCfg.WSEndpoints[0]
	}
	adapterCfg.RetryDelay = retryDelay

	return ripple.NewAdapter(adapterCfg)
}

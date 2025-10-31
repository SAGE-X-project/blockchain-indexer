package repository

import (
	"context"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// ChainRepository defines the interface for chain configuration data access
// Following the Repository Pattern and Interface Segregation Principle
type ChainRepository interface {
	// Read operations
	GetChain(ctx context.Context, chainID string) (*models.Chain, error)
	GetAllChains(ctx context.Context) ([]*models.Chain, error)
	GetChainsByType(ctx context.Context, chainType models.ChainType) ([]*models.Chain, error)
	GetEnabledChains(ctx context.Context) ([]*models.Chain, error)
	HasChain(ctx context.Context, chainID string) (bool, error)

	// Write operations
	SaveChain(ctx context.Context, chain *models.Chain) error
	UpdateChain(ctx context.Context, chain *models.Chain) error
	DeleteChain(ctx context.Context, chainID string) error

	// Status operations
	UpdateChainStatus(ctx context.Context, chainID string, status models.ChainStatus) error
	UpdateLatestBlock(ctx context.Context, chainID string, indexedBlock, chainBlock uint64) error

	// Statistics
	GetChainStats(ctx context.Context, chainID string) (*models.ChainStats, error)
	GetAllChainStats(ctx context.Context) ([]*models.ChainStats, error)
}

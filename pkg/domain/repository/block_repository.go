package repository

import (
	"context"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// BlockRepository defines the interface for block data access
// Following the Repository Pattern and Interface Segregation Principle
type BlockRepository interface {
	// Read operations
	GetBlock(ctx context.Context, chainID string, number uint64) (*models.Block, error)
	GetBlockByHash(ctx context.Context, chainID string, hash string) (*models.Block, error)
	GetBlocks(ctx context.Context, chainID string, start, end uint64) ([]*models.Block, error)
	GetLatestBlock(ctx context.Context, chainID string) (*models.Block, error)
	GetLatestHeight(ctx context.Context, chainID string) (uint64, error)
	HasBlock(ctx context.Context, chainID string, number uint64) (bool, error)

	// Query operations with filtering and pagination
	QueryBlocks(ctx context.Context, filter *models.BlockFilter, pagination *models.PaginationOptions) ([]*models.Block, error)
	QueryBlockSummaries(ctx context.Context, filter *models.BlockFilter, pagination *models.PaginationOptions) ([]*models.BlockSummary, error)
	CountBlocks(ctx context.Context, filter *models.BlockFilter) (uint64, error)

	// Write operations
	SaveBlock(ctx context.Context, block *models.Block) error
	SaveBlocks(ctx context.Context, blocks []*models.Block) error
	UpdateBlock(ctx context.Context, block *models.Block) error
	DeleteBlock(ctx context.Context, chainID string, number uint64) error

	// Batch operations
	SaveBlocksBatch(ctx context.Context, blocks []*models.Block, batchSize int) error
}

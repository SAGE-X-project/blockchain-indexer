package repository

import (
	"context"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// TransactionRepository defines the interface for transaction data access
// Following the Repository Pattern and Interface Segregation Principle
type TransactionRepository interface {
	// Read operations
	GetTransaction(ctx context.Context, chainID string, hash string) (*models.Transaction, error)
	GetTransactionsByBlock(ctx context.Context, chainID string, blockNumber uint64) ([]*models.Transaction, error)
	GetTransactionsByAddress(ctx context.Context, chainID string, address string, pagination *models.PaginationOptions) ([]*models.Transaction, error)
	HasTransaction(ctx context.Context, chainID string, hash string) (bool, error)

	// Query operations with filtering and pagination
	QueryTransactions(ctx context.Context, filter *models.TransactionFilter, pagination *models.PaginationOptions) ([]*models.Transaction, error)
	QueryTransactionSummaries(ctx context.Context, filter *models.TransactionFilter, pagination *models.PaginationOptions) ([]*models.TransactionSummary, error)
	CountTransactions(ctx context.Context, filter *models.TransactionFilter) (uint64, error)

	// Write operations
	SaveTransaction(ctx context.Context, tx *models.Transaction) error
	SaveTransactions(ctx context.Context, txs []*models.Transaction) error
	UpdateTransaction(ctx context.Context, tx *models.Transaction) error
	DeleteTransaction(ctx context.Context, chainID string, hash string) error

	// Index operations (for address lookups)
	AddAddressIndex(ctx context.Context, chainID string, address string, txHash string) error
	GetAddressTransactions(ctx context.Context, chainID string, address string, pagination *models.PaginationOptions) ([]string, error)

	// Batch operations
	SaveTransactionsBatch(ctx context.Context, txs []*models.Transaction, batchSize int) error
}

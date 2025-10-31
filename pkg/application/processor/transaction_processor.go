package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/metrics"
	"go.uber.org/zap"
)

// TransactionProcessor processes and indexes blockchain transactions
type TransactionProcessor struct {
	txRepo  repository.TransactionRepository
	logger  *logger.Logger
	metrics *metrics.Metrics
}

// NewTransactionProcessor creates a new transaction processor
func NewTransactionProcessor(
	txRepo repository.TransactionRepository,
	logger *logger.Logger,
	metrics *metrics.Metrics,
) *TransactionProcessor {
	return &TransactionProcessor{
		txRepo:  txRepo,
		logger:  logger,
		metrics: metrics,
	}
}

// ProcessTransaction processes a single transaction and stores it
func (p *TransactionProcessor) ProcessTransaction(ctx context.Context, tx *models.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	startTime := time.Now()
	chainID := tx.ChainID

	p.logger.Debug("processing transaction",
		zap.String("chain_id", chainID),
		zap.String("tx_hash", tx.Hash),
		zap.Uint64("block_number", tx.BlockNumber),
	)

	// Validate transaction
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	// Save transaction
	if err := p.txRepo.SaveTransaction(ctx, tx); err != nil {
		return fmt.Errorf("failed to save transaction %s: %w", tx.Hash, err)
	}

	// Record metrics
	duration := time.Since(startTime)
	p.metrics.RecordTransactionIndexed(chainID)

	p.logger.Debug("transaction processed successfully",
		zap.String("chain_id", chainID),
		zap.String("tx_hash", tx.Hash),
		zap.Duration("duration", duration),
	)

	return nil
}

// ProcessTransactions processes multiple transactions
func (p *TransactionProcessor) ProcessTransactions(ctx context.Context, transactions []*models.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	chainID := transactions[0].ChainID
	p.logger.Info("processing transactions batch",
		zap.String("chain_id", chainID),
		zap.Int("count", len(transactions)),
	)

	successCount := 0
	errorCount := 0

	for _, tx := range transactions {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := p.ProcessTransaction(ctx, tx); err != nil {
			p.logger.Error("failed to process transaction",
				zap.String("chain_id", chainID),
				zap.String("tx_hash", tx.Hash),
				zap.Error(err),
			)
			errorCount++
			// Continue processing other transactions
		} else {
			successCount++
		}
	}

	p.logger.Info("transactions batch processed",
		zap.String("chain_id", chainID),
		zap.Int("success", successCount),
		zap.Int("errors", errorCount),
	)

	if errorCount > 0 {
		return fmt.Errorf("processed %d transactions with %d errors", len(transactions), errorCount)
	}

	return nil
}

// GetTransactionStats returns transaction statistics for a chain
func (p *TransactionProcessor) GetTransactionStats(ctx context.Context, chainID string) (*TransactionStats, error) {
	// Get total transaction count
	// Note: This is a simplified version. In production, you'd want to cache these stats
	// or compute them incrementally

	stats := &TransactionStats{
		ChainID: chainID,
	}

	// For now, return basic stats
	// In production, you'd implement proper stats collection

	return stats, nil
}

// TransactionStats represents transaction statistics
type TransactionStats struct {
	ChainID              string
	TotalTransactions    uint64
	SuccessTransactions  uint64
	FailedTransactions   uint64
	PendingTransactions  uint64
	AverageGasUsed       uint64
	AverageTransactionFee string
}

// String returns a string representation of the stats
func (s *TransactionStats) String() string {
	return fmt.Sprintf(
		"Chain: %s, Total: %d, Success: %d, Failed: %d, Pending: %d",
		s.ChainID,
		s.TotalTransactions,
		s.SuccessTransactions,
		s.FailedTransactions,
		s.PendingTransactions,
	)
}

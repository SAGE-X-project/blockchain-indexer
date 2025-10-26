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

// BlockProcessor processes and stores blockchain blocks
type BlockProcessor struct {
	blockRepo repository.BlockRepository
	txRepo    repository.TransactionRepository
	chainRepo repository.ChainRepository
	logger    *logger.Logger
	metrics   *metrics.Metrics
}

// NewBlockProcessor creates a new block processor
func NewBlockProcessor(
	blockRepo repository.BlockRepository,
	txRepo repository.TransactionRepository,
	chainRepo repository.ChainRepository,
	logger *logger.Logger,
	metrics *metrics.Metrics,
) *BlockProcessor {
	return &BlockProcessor{
		blockRepo: blockRepo,
		txRepo:    txRepo,
		chainRepo: chainRepo,
		logger:    logger,
		metrics:   metrics,
	}
}

// ProcessBlock processes a single block and stores it
func (p *BlockProcessor) ProcessBlock(ctx context.Context, block *models.Block) error {
	if block == nil {
		return fmt.Errorf("block is nil")
	}

	startTime := time.Now()
	chainID := block.ChainID

	p.logger.Debug("processing block",
		zap.String("chain_id", chainID),
		zap.Uint64("block_number", block.Number),
		zap.String("block_hash", block.Hash),
		zap.Int("tx_count", block.TxCount),
	)

	// Validate block
	if err := block.Validate(); err != nil {
		p.metrics.RecordBlockProcessed(chainID, false)
		return fmt.Errorf("invalid block: %w", err)
	}

	// Save block
	if err := p.blockRepo.SaveBlock(ctx, block); err != nil {
		p.metrics.RecordBlockProcessed(chainID, false)
		return fmt.Errorf("failed to save block %d: %w", block.Number, err)
	}

	// Save transactions if present
	if len(block.Transactions) > 0 {
		for _, tx := range block.Transactions {
			if err := p.txRepo.SaveTransaction(ctx, tx); err != nil {
				p.logger.Error("failed to save transaction",
					zap.String("chain_id", chainID),
					zap.String("tx_hash", tx.Hash),
					zap.Error(err),
				)
				// Continue processing other transactions
			} else {
				p.metrics.RecordTransactionIndexed(chainID)
			}
		}
	}

	// Update chain latest indexed block
	if err := p.updateChainProgress(ctx, chainID, block.Number); err != nil {
		p.logger.Warn("failed to update chain progress",
			zap.String("chain_id", chainID),
			zap.Uint64("block_number", block.Number),
			zap.Error(err),
		)
	}

	// Record metrics
	duration := time.Since(startTime)
	p.metrics.RecordBlockIndexed(chainID)
	p.metrics.RecordBlockProcessed(chainID, true)
	p.metrics.RecordBlockProcessTime(chainID, duration)
	p.metrics.UpdateLatestBlockHeight(chainID, block.Number)

	p.logger.Debug("block processed successfully",
		zap.String("chain_id", chainID),
		zap.Uint64("block_number", block.Number),
		zap.Duration("duration", duration),
	)

	return nil
}

// ProcessBlocks processes multiple blocks in order
func (p *BlockProcessor) ProcessBlocks(ctx context.Context, blocks []*models.Block) error {
	if len(blocks) == 0 {
		return nil
	}

	chainID := blocks[0].ChainID
	p.logger.Info("processing blocks batch",
		zap.String("chain_id", chainID),
		zap.Int("count", len(blocks)),
		zap.Uint64("start_block", blocks[0].Number),
		zap.Uint64("end_block", blocks[len(blocks)-1].Number),
	)

	successCount := 0
	errorCount := 0

	for _, block := range blocks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := p.ProcessBlock(ctx, block); err != nil {
			p.logger.Error("failed to process block",
				zap.String("chain_id", chainID),
				zap.Uint64("block_number", block.Number),
				zap.Error(err),
			)
			errorCount++
			// Continue processing other blocks
		} else {
			successCount++
		}
	}

	p.logger.Info("blocks batch processed",
		zap.String("chain_id", chainID),
		zap.Int("success", successCount),
		zap.Int("errors", errorCount),
	)

	if errorCount > 0 {
		return fmt.Errorf("processed %d blocks with %d errors", len(blocks), errorCount)
	}

	return nil
}

// updateChainProgress updates the chain's latest indexed block
func (p *BlockProcessor) updateChainProgress(ctx context.Context, chainID string, blockNumber uint64) error {
	chain, err := p.chainRepo.GetChain(ctx, chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain: %w", err)
	}

	if chain == nil {
		return fmt.Errorf("chain not found: %s", chainID)
	}

	// Only update if this block is newer
	if blockNumber > chain.LatestIndexedBlock {
		chain.LatestIndexedBlock = blockNumber
		chain.LastUpdated = time.Now()

		if err := p.chainRepo.UpdateChain(ctx, chain); err != nil {
			return fmt.Errorf("failed to update chain: %w", err)
		}
	}

	return nil
}

// GetProgress returns the processing progress for a chain
func (p *BlockProcessor) GetProgress(ctx context.Context, chainID string) (*ProcessProgress, error) {
	chain, err := p.chainRepo.GetChain(ctx, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain: %w", err)
	}

	if chain == nil {
		return nil, fmt.Errorf("chain not found: %s", chainID)
	}

	latestBlock, err := p.blockRepo.GetLatestBlock(ctx, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	var latestBlockNumber uint64
	if latestBlock != nil {
		latestBlockNumber = latestBlock.Number
	}

	progress := &ProcessProgress{
		ChainID:            chainID,
		LatestIndexedBlock: latestBlockNumber,
		TargetBlock:        chain.LatestChainBlock,
		StartBlock:         chain.StartBlock,
		LastUpdated:        chain.LastUpdated,
	}

	// Calculate progress percentage
	if chain.LatestChainBlock > chain.StartBlock {
		total := chain.LatestChainBlock - chain.StartBlock
		indexed := latestBlockNumber - chain.StartBlock
		if indexed > total {
			indexed = total
		}
		progress.ProgressPercentage = float64(indexed) / float64(total) * 100
	}

	// Calculate blocks behind
	if chain.LatestChainBlock > latestBlockNumber {
		progress.BlocksBehind = chain.LatestChainBlock - latestBlockNumber
	}

	return progress, nil
}

// ProcessProgress represents block processing progress
type ProcessProgress struct {
	ChainID            string
	LatestIndexedBlock uint64
	TargetBlock        uint64
	StartBlock         uint64
	BlocksBehind       uint64
	ProgressPercentage float64
	LastUpdated        time.Time
}

// String returns a string representation of the progress
func (p *ProcessProgress) String() string {
	return fmt.Sprintf(
		"Chain: %s, Progress: %.2f%%, Indexed: %d, Target: %d, Behind: %d",
		p.ChainID,
		p.ProgressPercentage,
		p.LatestIndexedBlock,
		p.TargetBlock,
		p.BlocksBehind,
	)
}

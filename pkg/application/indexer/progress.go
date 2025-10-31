package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/application/processor"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/service"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/metrics"
	"go.uber.org/zap"
)

// ProgressTracker tracks indexing progress
type ProgressTracker struct {
	adapter        service.ChainAdapter
	blockRepo      repository.BlockRepository
	chainRepo      repository.ChainRepository
	blockProcessor *processor.BlockProcessor
	logger         *logger.Logger
	metrics        *metrics.Metrics
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(
	adapter service.ChainAdapter,
	blockRepo repository.BlockRepository,
	chainRepo repository.ChainRepository,
	blockProcessor *processor.BlockProcessor,
	logger *logger.Logger,
	metrics *metrics.Metrics,
) *ProgressTracker {
	return &ProgressTracker{
		adapter:        adapter,
		blockRepo:      blockRepo,
		chainRepo:      chainRepo,
		blockProcessor: blockProcessor,
		logger:         logger,
		metrics:        metrics,
	}
}

// Progress represents indexing progress information
type Progress struct {
	ChainID            string
	ChainType          string
	LatestIndexedBlock uint64
	LatestChainBlock   uint64
	TargetBlock        uint64 // Same as LatestChainBlock for compatibility
	StartBlock         uint64
	BlocksBehind       uint64
	ProgressPercentage float64
	BlocksPerSecond    float64
	EstimatedTimeLeft  time.Duration
	LastUpdated        time.Time
	Status             string
}

// String returns a string representation of the progress
func (p *Progress) String() string {
	return fmt.Sprintf(
		"Chain: %s (%s), Progress: %.2f%%, Indexed: %d, Latest: %d, Behind: %d, Speed: %.2f blocks/s, ETA: %v, Status: %s",
		p.ChainID,
		p.ChainType,
		p.ProgressPercentage,
		p.LatestIndexedBlock,
		p.LatestChainBlock,
		p.BlocksBehind,
		p.BlocksPerSecond,
		p.EstimatedTimeLeft,
		p.Status,
	)
}

// GetProgress returns current indexing progress
func (t *ProgressTracker) GetProgress(ctx context.Context, chainID string) (*Progress, error) {
	// Get chain configuration
	chain, err := t.chainRepo.GetChain(ctx, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain: %w", err)
	}

	if chain == nil {
		return nil, fmt.Errorf("chain not found: %s", chainID)
	}

	// Get latest indexed block
	latestBlock, err := t.blockRepo.GetLatestBlock(ctx, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	var latestIndexedBlock uint64
	if latestBlock != nil {
		latestIndexedBlock = latestBlock.Number
	}

	// Get latest chain block
	latestChainBlock, err := t.adapter.GetLatestBlockNumber(ctx)
	if err != nil {
		t.logger.Warn("failed to get latest chain block, using cached value",
			zap.String("chain_id", chainID),
			zap.Error(err),
		)
		latestChainBlock = chain.LatestChainBlock
	} else {
		// Update chain's latest block
		chain.LatestChainBlock = latestChainBlock
		chain.LastUpdated = time.Now()
		if err := t.chainRepo.UpdateChain(ctx, chain); err != nil {
			t.logger.Warn("failed to update chain",
				zap.String("chain_id", chainID),
				zap.Error(err),
			)
		}
	}

	progress := &Progress{
		ChainID:            chainID,
		ChainType:          string(chain.ChainType),
		LatestIndexedBlock: latestIndexedBlock,
		LatestChainBlock:   latestChainBlock,
		TargetBlock:        latestChainBlock,
		StartBlock:         chain.StartBlock,
		LastUpdated:        time.Now(),
		Status:             string(chain.Status),
	}

	// Calculate blocks behind
	if latestChainBlock > latestIndexedBlock {
		progress.BlocksBehind = latestChainBlock - latestIndexedBlock
	}

	// Calculate progress percentage
	if latestChainBlock > chain.StartBlock {
		total := latestChainBlock - chain.StartBlock
		indexed := latestIndexedBlock - chain.StartBlock
		if indexed > total {
			indexed = total
		}
		progress.ProgressPercentage = float64(indexed) / float64(total) * 100
	}

	// Calculate blocks per second and ETA
	// This is a simplified calculation - in production you'd track historical rates
	if progress.BlocksBehind > 0 {
		// Estimate based on recent processing speed
		// For now, use a simple estimate
		progress.BlocksPerSecond = 10.0 // Placeholder
		if progress.BlocksPerSecond > 0 {
			secondsLeft := float64(progress.BlocksBehind) / progress.BlocksPerSecond
			progress.EstimatedTimeLeft = time.Duration(secondsLeft) * time.Second
		}
	}

	// Update metrics
	t.metrics.UpdateLatestBlockHeight(chainID, latestIndexedBlock)
	t.metrics.UpdateChainBlocksBehind(chainID, progress.BlocksBehind)
	t.metrics.UpdateChainSyncProgress(chainID, progress.ProgressPercentage)

	// Update sync status
	syncStatus := 0 // stopped
	if progress.BlocksBehind > 100 {
		syncStatus = 1 // syncing
	} else if progress.BlocksBehind < 10 {
		syncStatus = 2 // synced
	}
	t.metrics.UpdateChainSyncStatus(chainID, syncStatus)

	return progress, nil
}

// GetAllProgress returns progress for specific chains
// Note: In production, you'd want a ListChains method in the repository
// For now, this is a placeholder that requires chain IDs
func (t *ProgressTracker) GetAllProgress(ctx context.Context, chainIDs []string) ([]*Progress, error) {
	progressList := make([]*Progress, 0, len(chainIDs))

	for _, chainID := range chainIDs {
		progress, err := t.GetProgress(ctx, chainID)
		if err != nil {
			t.logger.Error("failed to get progress",
				zap.String("chain_id", chainID),
				zap.Error(err),
			)
			continue
		}

		progressList = append(progressList, progress)
	}

	return progressList, nil
}

// LogProgress logs progress for a chain
func (t *ProgressTracker) LogProgress(ctx context.Context, chainID string) error {
	progress, err := t.GetProgress(ctx, chainID)
	if err != nil {
		return err
	}

	t.logger.Info("indexing progress",
		zap.String("chain_id", progress.ChainID),
		zap.String("chain_type", progress.ChainType),
		zap.Uint64("indexed_block", progress.LatestIndexedBlock),
		zap.Uint64("latest_block", progress.LatestChainBlock),
		zap.Uint64("blocks_behind", progress.BlocksBehind),
		zap.Float64("progress_pct", progress.ProgressPercentage),
		zap.Float64("blocks_per_sec", progress.BlocksPerSecond),
		zap.Duration("eta", progress.EstimatedTimeLeft),
		zap.String("status", progress.Status),
	)

	return nil
}

// LogAllProgress logs progress for specific chains
func (t *ProgressTracker) LogAllProgress(ctx context.Context, chainIDs []string) error {
	progressList, err := t.GetAllProgress(ctx, chainIDs)
	if err != nil {
		return err
	}

	for _, progress := range progressList {
		t.logger.Info("chain progress",
			zap.String("chain_id", progress.ChainID),
			zap.Float64("progress_pct", progress.ProgressPercentage),
			zap.Uint64("blocks_behind", progress.BlocksBehind),
		)
	}

	return nil
}

// IsSynced checks if a chain is synced (within threshold blocks of latest)
func (t *ProgressTracker) IsSynced(ctx context.Context, chainID string, threshold uint64) (bool, error) {
	progress, err := t.GetProgress(ctx, chainID)
	if err != nil {
		return false, err
	}

	return progress.BlocksBehind <= threshold, nil
}

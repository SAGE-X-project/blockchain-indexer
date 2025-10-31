package indexer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/application/processor"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/service"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"go.uber.org/zap"
)

// BlockIndexer indexes blocks from a blockchain
type BlockIndexer struct {
	adapter        service.ChainAdapter
	processor      *processor.BlockProcessor
	workerPool     *WorkerPool
	gapRecovery    *GapRecovery
	progressTracker *ProgressTracker
	logger         *logger.Logger

	// State
	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}

	// Configuration
	config *BlockIndexerConfig
}

// BlockIndexerConfig holds block indexer configuration
type BlockIndexerConfig struct {
	ChainID           string
	StartBlock        uint64
	EndBlock          uint64 // 0 means continuous indexing
	BatchSize         int
	WorkerCount       int
	ConfirmationBlocks uint64
	PollInterval      time.Duration
	EnableGapRecovery bool
}

// DefaultBlockIndexerConfig returns default configuration
func DefaultBlockIndexerConfig(chainID string) *BlockIndexerConfig {
	return &BlockIndexerConfig{
		ChainID:            chainID,
		StartBlock:         0,
		EndBlock:           0,
		BatchSize:          100,
		WorkerCount:        10,
		ConfirmationBlocks: 12,
		PollInterval:       5 * time.Second,
		EnableGapRecovery:  true,
	}
}

// NewBlockIndexer creates a new block indexer
func NewBlockIndexer(
	adapter service.ChainAdapter,
	processor *processor.BlockProcessor,
	gapRecovery *GapRecovery,
	progressTracker *ProgressTracker,
	config *BlockIndexerConfig,
	logger *logger.Logger,
) *BlockIndexer {
	if config == nil {
		config = DefaultBlockIndexerConfig(adapter.GetChainID())
	}

	// Create worker pool
	poolConfig := &WorkerPoolConfig{
		WorkerCount: config.WorkerCount,
		QueueSize:   config.WorkerCount * 2,
		ResultSize:  config.WorkerCount * 2,
	}

	workerPool := NewWorkerPool(poolConfig, nil, logger)

	indexer := &BlockIndexer{
		adapter:         adapter,
		processor:       processor,
		workerPool:      workerPool,
		gapRecovery:     gapRecovery,
		progressTracker: progressTracker,
		config:          config,
		logger:          logger,
		stopChan:        make(chan struct{}),
	}

	// Set job handler for worker pool
	indexer.workerPool = NewWorkerPool(poolConfig, indexer.handleJob, logger)

	return indexer
}

// Start starts the block indexer
func (b *BlockIndexer) Start(ctx context.Context) error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return fmt.Errorf("indexer already running")
	}
	b.running = true
	b.mu.Unlock()

	b.logger.Info("starting block indexer",
		zap.String("chain_id", b.config.ChainID),
		zap.Uint64("start_block", b.config.StartBlock),
		zap.Uint64("end_block", b.config.EndBlock),
		zap.Int("batch_size", b.config.BatchSize),
		zap.Int("workers", b.config.WorkerCount),
	)

	// Start result handler
	go b.handleResults()

	// Start indexing
	go b.indexLoop(ctx)

	// Start gap recovery if enabled
	if b.config.EnableGapRecovery && b.gapRecovery != nil {
		go b.gapRecoveryLoop(ctx)
	}

	// Start progress tracking
	if b.progressTracker != nil {
		go b.progressTrackingLoop(ctx)
	}

	return nil
}

// Stop stops the block indexer
func (b *BlockIndexer) Stop() error {
	b.mu.Lock()
	if !b.running {
		b.mu.Unlock()
		return fmt.Errorf("indexer not running")
	}
	b.running = false
	b.mu.Unlock()

	b.logger.Info("stopping block indexer", zap.String("chain_id", b.config.ChainID))

	// Signal stop
	close(b.stopChan)

	// Stop worker pool
	b.workerPool.Stop()

	b.logger.Info("block indexer stopped", zap.String("chain_id", b.config.ChainID))

	return nil
}

// indexLoop is the main indexing loop
func (b *BlockIndexer) indexLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			b.logger.Error("indexing loop panicked",
				zap.String("chain_id", b.config.ChainID),
				zap.Any("panic", r),
			)
		}
	}()

	currentBlock := b.config.StartBlock
	ticker := time.NewTicker(b.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-b.stopChan:
			return
		case <-ticker.C:
			// Check if we reached end block
			if b.config.EndBlock > 0 && currentBlock >= b.config.EndBlock {
				b.logger.Info("reached end block",
					zap.String("chain_id", b.config.ChainID),
					zap.Uint64("end_block", b.config.EndBlock),
				)
				return
			}

			// Get latest block from chain
			latestBlock, err := b.adapter.GetLatestBlockNumber(ctx)
			if err != nil {
				b.logger.Error("failed to get latest block number",
					zap.String("chain_id", b.config.ChainID),
					zap.Error(err),
				)
				continue
			}

			// Apply confirmation blocks
			confirmedBlock := latestBlock
			if latestBlock > b.config.ConfirmationBlocks {
				confirmedBlock = latestBlock - b.config.ConfirmationBlocks
			}

			// Index blocks in batches
			for currentBlock <= confirmedBlock {
				select {
				case <-ctx.Done():
					return
				case <-b.stopChan:
					return
				default:
				}

				// Calculate batch end
				batchEnd := currentBlock + uint64(b.config.BatchSize) - 1
				if batchEnd > confirmedBlock {
					batchEnd = confirmedBlock
				}

				// Submit batch job
				job := Job{
					ID:   fmt.Sprintf("block-range-%d-%d", currentBlock, batchEnd),
					Type: JobTypeBlockRange,
					Payload: &BlockRangePayload{
						StartBlock: currentBlock,
						EndBlock:   batchEnd,
					},
				}

				if err := b.workerPool.Submit(job); err != nil {
					b.logger.Error("failed to submit job",
						zap.String("chain_id", b.config.ChainID),
						zap.String("job_id", job.ID),
						zap.Error(err),
					)
					// Wait a bit before retry
					time.Sleep(time.Second)
					continue
				}

				currentBlock = batchEnd + 1
			}
		}
	}
}

// gapRecoveryLoop periodically runs gap recovery
func (b *BlockIndexer) gapRecoveryLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-b.stopChan:
			return
		case <-ticker.C:
			gaps, err := b.gapRecovery.DetectGaps(ctx, b.config.ChainID)
			if err != nil {
				b.logger.Error("failed to detect gaps",
					zap.String("chain_id", b.config.ChainID),
					zap.Error(err),
				)
				continue
			}

			if len(gaps) > 0 {
				b.logger.Info("detected gaps",
					zap.String("chain_id", b.config.ChainID),
					zap.Int("gap_count", len(gaps)),
				)

				for _, gap := range gaps {
					if err := b.gapRecovery.RecoverGap(ctx, gap); err != nil {
						b.logger.Error("failed to recover gap",
							zap.String("chain_id", b.config.ChainID),
							zap.Uint64("start", gap.StartBlock),
							zap.Uint64("end", gap.EndBlock),
							zap.Error(err),
						)
					}
				}
			}
		}
	}
}

// progressTrackingLoop periodically logs progress
func (b *BlockIndexer) progressTrackingLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-b.stopChan:
			return
		case <-ticker.C:
			progress, err := b.progressTracker.GetProgress(ctx, b.config.ChainID)
			if err != nil {
				b.logger.Error("failed to get progress",
					zap.String("chain_id", b.config.ChainID),
					zap.Error(err),
				)
				continue
			}

			stats := b.workerPool.Stats()

			b.logger.Info("indexing progress",
				zap.String("chain_id", b.config.ChainID),
				zap.Uint64("indexed_block", progress.LatestIndexedBlock),
				zap.Uint64("target_block", progress.TargetBlock),
				zap.Uint64("blocks_behind", progress.BlocksBehind),
				zap.Float64("progress_pct", progress.ProgressPercentage),
				zap.Int("active_workers", stats.ActiveWorkers),
				zap.Uint64("completed_jobs", stats.CompletedJobs),
				zap.Uint64("failed_jobs", stats.FailedJobs),
			)
		}
	}
}

// handleResults handles job results
func (b *BlockIndexer) handleResults() {
	for result := range b.workerPool.Results() {
		if !result.Success {
			b.logger.Error("job failed",
				zap.String("chain_id", b.config.ChainID),
				zap.String("job_id", result.JobID),
				zap.Error(result.Error),
				zap.Duration("duration", result.Duration),
			)
		} else {
			b.logger.Debug("job completed",
				zap.String("chain_id", b.config.ChainID),
				zap.String("job_id", result.JobID),
				zap.Duration("duration", result.Duration),
			)
		}
	}
}

// handleJob processes a job
func (b *BlockIndexer) handleJob(ctx context.Context, job Job) Result {
	switch job.Type {
	case JobTypeBlockRange:
		return b.handleBlockRangeJob(ctx, job)
	case JobTypeBlock:
		return b.handleBlockJob(ctx, job)
	default:
		return Result{
			Success: false,
			Error:   fmt.Errorf("unknown job type: %v", job.Type),
		}
	}
}

// handleBlockRangeJob processes a block range job
func (b *BlockIndexer) handleBlockRangeJob(ctx context.Context, job Job) Result {
	payload, ok := job.Payload.(*BlockRangePayload)
	if !ok {
		return Result{
			Success: false,
			Error:   fmt.Errorf("invalid payload type"),
		}
	}

	// Fetch blocks
	blocks, err := b.adapter.GetBlocks(ctx, payload.StartBlock, payload.EndBlock)
	if err != nil {
		return Result{
			Success: false,
			Error:   fmt.Errorf("failed to fetch blocks: %w", err),
		}
	}

	// Process blocks
	if err := b.processor.ProcessBlocks(ctx, blocks); err != nil {
		return Result{
			Success: false,
			Error:   fmt.Errorf("failed to process blocks: %w", err),
		}
	}

	return Result{
		Success: true,
		Payload: len(blocks),
	}
}

// handleBlockJob processes a single block job
func (b *BlockIndexer) handleBlockJob(ctx context.Context, job Job) Result {
	payload, ok := job.Payload.(*BlockPayload)
	if !ok {
		return Result{
			Success: false,
			Error:   fmt.Errorf("invalid payload type"),
		}
	}

	// Fetch block
	block, err := b.adapter.GetBlockByNumber(ctx, payload.BlockNumber)
	if err != nil {
		return Result{
			Success: false,
			Error:   fmt.Errorf("failed to fetch block: %w", err),
		}
	}

	// Process block
	if err := b.processor.ProcessBlock(ctx, block); err != nil {
		return Result{
			Success: false,
			Error:   fmt.Errorf("failed to process block: %w", err),
		}
	}

	return Result{
		Success: true,
		Payload: block,
	}
}

// IsRunning returns true if the indexer is running
func (b *BlockIndexer) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

// GetStats returns indexer statistics
func (b *BlockIndexer) GetStats() *IndexerStats {
	return &IndexerStats{
		ChainID:     b.config.ChainID,
		Running:     b.IsRunning(),
		WorkerStats: b.workerPool.Stats(),
	}
}

// IndexerStats represents indexer statistics
type IndexerStats struct {
	ChainID     string
	Running     bool
	WorkerStats *WorkerPoolStats
}

// BlockRangePayload represents a block range job payload
type BlockRangePayload struct {
	StartBlock uint64
	EndBlock   uint64
}

// BlockPayload represents a single block job payload
type BlockPayload struct {
	BlockNumber uint64
}

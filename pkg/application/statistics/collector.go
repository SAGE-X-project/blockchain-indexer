package statistics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/event"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/metrics"
)

// Collector collects and aggregates blockchain statistics
type Collector struct {
	statsRepo repository.StatisticsRepository
	blockRepo repository.BlockRepository
	txRepo    repository.TransactionRepository
	chainRepo repository.ChainRepository
	eventBus  event.EventBus
	metrics   *metrics.Metrics
	logger    *logger.Logger

	// Configuration
	updateInterval     time.Duration
	snapshotInterval   time.Duration
	enableSnapshots    bool
	enableTimeSeries   bool

	// State
	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
}

// Config holds collector configuration
type Config struct {
	UpdateInterval   time.Duration // How often to update statistics (default: 30s)
	SnapshotInterval time.Duration // How often to take snapshots (default: 5m)
	EnableSnapshots  bool          // Enable snapshot creation
	EnableTimeSeries bool          // Enable time series data collection
}

// DefaultConfig returns default collector configuration
func DefaultConfig() *Config {
	return &Config{
		UpdateInterval:   30 * time.Second,
		SnapshotInterval: 5 * time.Minute,
		EnableSnapshots:  true,
		EnableTimeSeries: true,
	}
}

// NewCollector creates a new statistics collector
func NewCollector(
	statsRepo repository.StatisticsRepository,
	blockRepo repository.BlockRepository,
	txRepo repository.TransactionRepository,
	chainRepo repository.ChainRepository,
	eventBus event.EventBus,
	metrics *metrics.Metrics,
	logger *logger.Logger,
	config *Config,
) *Collector {
	if config == nil {
		config = DefaultConfig()
	}

	return &Collector{
		statsRepo:        statsRepo,
		blockRepo:        blockRepo,
		txRepo:           txRepo,
		chainRepo:        chainRepo,
		eventBus:         eventBus,
		metrics:          metrics,
		logger:           logger,
		updateInterval:   config.UpdateInterval,
		snapshotInterval: config.SnapshotInterval,
		enableSnapshots:  config.EnableSnapshots,
		enableTimeSeries: config.EnableTimeSeries,
		stopCh:           make(chan struct{}),
	}
}

// Start starts the statistics collector
func (c *Collector) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("collector already running")
	}
	c.running = true
	c.mu.Unlock()

	c.logger.Info("starting statistics collector",
		zap.Duration("update_interval", c.updateInterval),
		zap.Duration("snapshot_interval", c.snapshotInterval),
	)

	// Subscribe to block indexed events for real-time updates
	if c.eventBus != nil {
		c.eventBus.SubscribeType(event.EventTypeBlockIndexed, c.handleBlockIndexed)
		c.eventBus.SubscribeType(event.EventTypeTransactionIndexed, c.handleTransactionIndexed)
	}

	// Start periodic update loop
	go c.updateLoop(ctx)

	// Start snapshot loop if enabled
	if c.enableSnapshots {
		go c.snapshotLoop(ctx)
	}

	return nil
}

// Stop stops the statistics collector
func (c *Collector) Stop() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return fmt.Errorf("collector not running")
	}
	c.running = false
	c.mu.Unlock()

	c.logger.Info("stopping statistics collector")
	close(c.stopCh)

	return nil
}

// IsRunning returns true if the collector is running
func (c *Collector) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// updateLoop periodically updates statistics
func (c *Collector) updateLoop(ctx context.Context) {
	ticker := time.NewTicker(c.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			if err := c.UpdateAllStatistics(ctx); err != nil {
				c.logger.Error("failed to update statistics", zap.Error(err))
			}
		}
	}
}

// snapshotLoop periodically creates snapshots
func (c *Collector) snapshotLoop(ctx context.Context) {
	ticker := time.NewTicker(c.snapshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			if err := c.CreateSnapshots(ctx); err != nil {
				c.logger.Error("failed to create snapshots", zap.Error(err))
			}
		}
	}
}

// handleBlockIndexed handles block indexed events for real-time updates
func (c *Collector) handleBlockIndexed(evt *event.Event) {
	payload, ok := evt.Payload.(*event.BlockIndexedPayload)
	if !ok || payload == nil || payload.Block == nil {
		return
	}

	ctx := context.Background()
	chainID := payload.Block.ChainID

	// Increment block counter
	if err := c.statsRepo.IncrementChainCounter(ctx, chainID, "total_blocks", 1); err != nil {
		c.logger.Error("failed to increment block counter",
			zap.String("chain_id", chainID),
			zap.Error(err),
		)
	}

	// Increment transaction counter
	if payload.TransactionCount > 0 {
		if err := c.statsRepo.IncrementChainCounter(ctx, chainID, "total_transactions", uint64(payload.TransactionCount)); err != nil {
			c.logger.Error("failed to increment transaction counter",
				zap.String("chain_id", chainID),
				zap.Error(err),
			)
		}
	}

	// Update latest block number
	if err := c.statsRepo.UpdateChainStatistic(ctx, chainID, "latest_block_number", payload.Block.Number); err != nil {
		c.logger.Error("failed to update latest block number",
			zap.String("chain_id", chainID),
			zap.Error(err),
		)
	}
}

// handleTransactionIndexed handles transaction indexed events
func (c *Collector) handleTransactionIndexed(evt *event.Event) {
	payload, ok := evt.Payload.(*event.TransactionIndexedPayload)
	if !ok || payload == nil || payload.Transaction == nil {
		return
	}

	// Transaction count is already updated by block indexed event
	// This handler can be used for transaction-specific statistics if needed
}

// UpdateAllStatistics updates statistics for all chains
func (c *Collector) UpdateAllStatistics(ctx context.Context) error {
	chains, err := c.chainRepo.GetAllChains(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chains: %w", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(chains))

	for _, chain := range chains {
		wg.Add(1)
		go func(ch *models.Chain) {
			defer wg.Done()
			if err := c.UpdateChainStatistics(ctx, ch.ChainID); err != nil {
				errCh <- fmt.Errorf("chain %s: %w", ch.ChainID, err)
			}
		}(chain)
	}

	wg.Wait()
	close(errCh)

	// Collect errors
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		c.logger.Warn("some chain statistics updates failed",
			zap.Int("error_count", len(errors)),
		)
	}

	// Update global statistics
	if err := c.UpdateGlobalStatistics(ctx); err != nil {
		return fmt.Errorf("failed to update global statistics: %w", err)
	}

	return nil
}

// UpdateChainStatistics updates statistics for a specific chain
func (c *Collector) UpdateChainStatistics(ctx context.Context, chainID string) error {
	// Get chain info
	chain, err := c.chainRepo.GetChain(ctx, chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain: %w", err)
	}

	// Get or create statistics
	stats, err := c.statsRepo.GetChainStatistics(ctx, chainID)
	if err != nil {
		if err == repository.ErrNotFound {
			stats = &models.ChainStatistics{
				ChainID:           chainID,
				ChainType:         string(chain.ChainType),
				ChainName:         chain.Name,
				IndexingStartTime: time.Now(),
			}
		} else {
			return fmt.Errorf("failed to get statistics: %w", err)
		}
	}

	// Get latest and oldest blocks
	latestBlock, err := c.blockRepo.GetLatestBlock(ctx, chainID)
	if err != nil && err != repository.ErrNotFound {
		return fmt.Errorf("failed to get latest block: %w", err)
	}

	// Count total blocks (approximate by latest block number for now)
	// In production, you might want a more accurate count
	if latestBlock != nil {
		stats.LatestBlockNumber = latestBlock.Number
		if latestBlock.Timestamp != nil {
			stats.LatestBlockTime = latestBlock.Timestamp.Time
		}

		// Estimate total blocks indexed
		if stats.OldestBlockNumber == 0 {
			stats.OldestBlockNumber = chain.StartBlock
		}

		if latestBlock.Number >= stats.OldestBlockNumber {
			stats.TotalBlocks = latestBlock.Number - stats.OldestBlockNumber + 1
		}

		// Calculate sync progress
		if chain.LatestChainBlock > 0 {
			stats.BlocksBehind = 0
			if chain.LatestChainBlock > latestBlock.Number {
				stats.BlocksBehind = chain.LatestChainBlock - latestBlock.Number
			}

			if chain.LatestChainBlock > chain.StartBlock {
				total := chain.LatestChainBlock - chain.StartBlock
				indexed := latestBlock.Number - chain.StartBlock
				if indexed > total {
					indexed = total
				}
				stats.SyncProgress = float64(indexed) / float64(total) * 100
			}
		}
	}

	// Update indexing rate
	if !stats.IndexingStartTime.IsZero() {
		duration := time.Since(stats.IndexingStartTime).Seconds()
		if duration > 0 && stats.TotalBlocks > 0 {
			stats.IndexingRate = float64(stats.TotalBlocks) / duration
		}
	}

	// Calculate averages
	stats.CalculateAverages()

	// Save updated statistics
	if err := c.statsRepo.SaveChainStatistics(ctx, stats); err != nil {
		return fmt.Errorf("failed to save statistics: %w", err)
	}

	c.logger.Debug("updated chain statistics",
		zap.String("chain_id", chainID),
		zap.Uint64("total_blocks", stats.TotalBlocks),
		zap.Uint64("total_transactions", stats.TotalTransactions),
		zap.Float64("sync_progress", stats.SyncProgress),
	)

	return nil
}

// UpdateGlobalStatistics updates global statistics by aggregating all chains
func (c *Collector) UpdateGlobalStatistics(ctx context.Context) error {
	// Get all chain statistics
	chainStats, err := c.statsRepo.GetAllChainStatistics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain statistics: %w", err)
	}

	// Create global statistics
	globalStats := &models.GlobalStatistics{
		TotalChains:   len(chainStats),
		LastUpdated:   time.Now(),
	}

	// Aggregate chain statistics
	globalStats.MergeChainStatistics(chainStats)

	// Save global statistics
	if err := c.statsRepo.SaveGlobalStatistics(ctx, globalStats); err != nil {
		return fmt.Errorf("failed to save global statistics: %w", err)
	}

	c.logger.Debug("updated global statistics",
		zap.Int("total_chains", globalStats.TotalChains),
		zap.Int("active_chains", globalStats.ActiveChains),
		zap.Uint64("total_blocks", globalStats.TotalBlocks),
		zap.Uint64("total_transactions", globalStats.TotalTransactions),
	)

	return nil
}

// CreateSnapshots creates statistics snapshots for all chains
func (c *Collector) CreateSnapshots(ctx context.Context) error {
	// Get all chain statistics
	chainStats, err := c.statsRepo.GetAllChainStatistics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain statistics: %w", err)
	}

	timestamp := time.Now()

	// Create snapshot for each chain
	for _, stats := range chainStats {
		snapshot := &models.StatisticsSnapshot{
			Timestamp:  timestamp,
			ChainStats: stats,
		}

		if err := c.statsRepo.SaveStatisticsSnapshot(ctx, snapshot); err != nil {
			c.logger.Error("failed to save chain snapshot",
				zap.String("chain_id", stats.ChainID),
				zap.Error(err),
			)
		}
	}

	// Get and create global snapshot
	globalStats, err := c.statsRepo.GetGlobalStatistics(ctx)
	if err != nil && err != repository.ErrNotFound {
		return fmt.Errorf("failed to get global statistics: %w", err)
	}

	if globalStats != nil {
		snapshot := &models.StatisticsSnapshot{
			Timestamp:   timestamp,
			GlobalStats: globalStats,
		}

		if err := c.statsRepo.SaveStatisticsSnapshot(ctx, snapshot); err != nil {
			c.logger.Error("failed to save global snapshot", zap.Error(err))
		}
	}

	c.logger.Debug("created statistics snapshots",
		zap.Int("chain_count", len(chainStats)),
		zap.Time("timestamp", timestamp),
	)

	return nil
}

// GetChainStatistics returns statistics for a specific chain
func (c *Collector) GetChainStatistics(ctx context.Context, chainID string) (*models.ChainStatistics, error) {
	return c.statsRepo.GetChainStatistics(ctx, chainID)
}

// GetGlobalStatistics returns global statistics
func (c *Collector) GetGlobalStatistics(ctx context.Context) (*models.GlobalStatistics, error) {
	return c.statsRepo.GetGlobalStatistics(ctx)
}

// GetStatisticsSnapshots returns snapshots within a time range
func (c *Collector) GetStatisticsSnapshots(ctx context.Context, chainID string, startTime, endTime time.Time) ([]*models.StatisticsSnapshot, error) {
	return c.statsRepo.GetStatisticsSnapshots(ctx, chainID, startTime, endTime)
}

// CleanupOldSnapshots removes snapshots older than the specified duration
func (c *Collector) CleanupOldSnapshots(ctx context.Context, olderThan time.Duration) error {
	before := time.Now().Add(-olderThan)

	c.logger.Info("cleaning up old snapshots",
		zap.Time("before", before),
		zap.Duration("older_than", olderThan),
	)

	if err := c.statsRepo.DeleteOldSnapshots(ctx, before); err != nil {
		return fmt.Errorf("failed to delete old snapshots: %w", err)
	}

	return nil
}

package pebble

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
)

const (
	statsPrefix          = "stats:"
	chainStatsPrefix     = "stats:chain:"
	globalStatsKey       = "stats:global"
	snapshotPrefix       = "stats:snapshot:"
	timeSeriesPrefix     = "stats:timeseries:"
)

// Ensure PebbleStorage implements StatisticsRepository
var _ repository.StatisticsRepository = (*PebbleStorage)(nil)

// In-memory cache for statistics to reduce DB reads
type statsCache struct {
	mu          sync.RWMutex
	chainStats  map[string]*models.ChainStatistics
	globalStats *models.GlobalStatistics
	lastUpdate  time.Time
	ttl         time.Duration
}

func newStatsCache(ttl time.Duration) *statsCache {
	return &statsCache{
		chainStats: make(map[string]*models.ChainStatistics),
		ttl:        ttl,
	}
}

func (c *statsCache) getChainStats(chainID string) (*models.ChainStatistics, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if time.Since(c.lastUpdate) > c.ttl {
		return nil, false
	}

	stats, ok := c.chainStats[chainID]
	return stats, ok
}

func (c *statsCache) setChainStats(chainID string, stats *models.ChainStatistics) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.chainStats[chainID] = stats
	c.lastUpdate = time.Now()
}

func (c *statsCache) getGlobalStats() (*models.GlobalStatistics, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if time.Since(c.lastUpdate) > c.ttl || c.globalStats == nil {
		return nil, false
	}

	return c.globalStats, true
}

func (c *statsCache) setGlobalStats(stats *models.GlobalStatistics) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.globalStats = stats
	c.lastUpdate = time.Now()
}

func (c *statsCache) invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.chainStats = make(map[string]*models.ChainStatistics)
	c.globalStats = nil
	c.lastUpdate = time.Time{}
}

// SaveChainStatistics saves chain statistics
func (s *PebbleStorage) SaveChainStatistics(ctx context.Context, stats *models.ChainStatistics) error {
	if stats == nil {
		return fmt.Errorf("statistics is nil")
	}

	if err := stats.Validate(); err != nil {
		return fmt.Errorf("invalid statistics: %w", err)
	}

	stats.LastUpdated = time.Now()
	stats.CalculateAverages()

	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal statistics: %w", err)
	}

	key := []byte(chainStatsPrefix + stats.ChainID)
	if err := s.db.Set(key, data, pebble.Sync); err != nil {
		return fmt.Errorf("failed to save statistics: %w", err)
	}

	// Update cache
	if s.statsCache != nil {
		s.statsCache.setChainStats(stats.ChainID, stats)
	}

	return nil
}

// GetChainStatistics retrieves chain statistics by chain ID
func (s *PebbleStorage) GetChainStatistics(ctx context.Context, chainID string) (*models.ChainStatistics, error) {
	if chainID == "" {
		return nil, fmt.Errorf("chain ID is required")
	}

	// Check cache first
	if s.statsCache != nil {
		if stats, ok := s.statsCache.getChainStats(chainID); ok {
			return stats, nil
		}
	}

	key := []byte(chainStatsPrefix + chainID)
	value, closer, err := s.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	defer closer.Close()

	var stats models.ChainStatistics
	if err := json.Unmarshal(value, &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal statistics: %w", err)
	}

	// Update cache
	if s.statsCache != nil {
		s.statsCache.setChainStats(chainID, &stats)
	}

	return &stats, nil
}

// GetAllChainStatistics retrieves statistics for all chains
func (s *PebbleStorage) GetAllChainStatistics(ctx context.Context) ([]*models.ChainStatistics, error) {
	prefix := []byte(chainStatsPrefix)
	// Create upper bound by incrementing last byte
	upperBound := make([]byte, len(prefix))
	copy(upperBound, prefix)
	upperBound[len(upperBound)-1]++

	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: upperBound,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Close()

	var statsList []*models.ChainStatistics
	for iter.First(); iter.Valid(); iter.Next() {
		var stats models.ChainStatistics
		if err := json.Unmarshal(iter.Value(), &stats); err != nil {
			continue // Skip invalid entries
		}
		statsList = append(statsList, &stats)
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	return statsList, nil
}

// SaveGlobalStatistics saves global statistics
func (s *PebbleStorage) SaveGlobalStatistics(ctx context.Context, stats *models.GlobalStatistics) error {
	if stats == nil {
		return fmt.Errorf("statistics is nil")
	}

	if err := stats.Validate(); err != nil {
		return fmt.Errorf("invalid statistics: %w", err)
	}

	stats.LastUpdated = time.Now()
	stats.CalculateAverages()

	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal statistics: %w", err)
	}

	key := []byte(globalStatsKey)
	if err := s.db.Set(key, data, pebble.Sync); err != nil {
		return fmt.Errorf("failed to save statistics: %w", err)
	}

	// Update cache
	if s.statsCache != nil {
		s.statsCache.setGlobalStats(stats)
	}

	return nil
}

// GetGlobalStatistics retrieves global statistics
func (s *PebbleStorage) GetGlobalStatistics(ctx context.Context) (*models.GlobalStatistics, error) {
	// Check cache first
	if s.statsCache != nil {
		if stats, ok := s.statsCache.getGlobalStats(); ok {
			return stats, nil
		}
	}

	key := []byte(globalStatsKey)
	value, closer, err := s.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	defer closer.Close()

	var stats models.GlobalStatistics
	if err := json.Unmarshal(value, &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal statistics: %w", err)
	}

	// Update cache
	if s.statsCache != nil {
		s.statsCache.setGlobalStats(&stats)
	}

	return &stats, nil
}

// SaveStatisticsSnapshot saves a statistics snapshot
func (s *PebbleStorage) SaveStatisticsSnapshot(ctx context.Context, snapshot *models.StatisticsSnapshot) error {
	if snapshot == nil {
		return fmt.Errorf("snapshot is nil")
	}

	timestamp := snapshot.Timestamp.Unix()
	chainID := "global"
	if snapshot.ChainStats != nil {
		chainID = snapshot.ChainStats.ChainID
	}

	key := []byte(fmt.Sprintf("%s%s:%d", snapshotPrefix, chainID, timestamp))
	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if err := s.db.Set(key, data, pebble.NoSync); err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}

	return nil
}

// GetStatisticsSnapshots retrieves statistics snapshots within a time range
func (s *PebbleStorage) GetStatisticsSnapshots(ctx context.Context, chainID string, startTime, endTime time.Time) ([]*models.StatisticsSnapshot, error) {
	if chainID == "" {
		chainID = "global"
	}

	startKey := []byte(fmt.Sprintf("%s%s:%d", snapshotPrefix, chainID, startTime.Unix()))
	endKey := []byte(fmt.Sprintf("%s%s:%d", snapshotPrefix, chainID, endTime.Unix()))

	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: startKey,
		UpperBound: endKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Close()

	var snapshots []*models.StatisticsSnapshot
	for iter.First(); iter.Valid(); iter.Next() {
		var snapshot models.StatisticsSnapshot
		if err := json.Unmarshal(iter.Value(), &snapshot); err != nil {
			continue // Skip invalid entries
		}
		snapshots = append(snapshots, &snapshot)
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	return snapshots, nil
}

// SaveTimeSeriesData saves time series data
func (s *PebbleStorage) SaveTimeSeriesData(ctx context.Context, data *models.TimeSeriesData) error {
	if data == nil {
		return fmt.Errorf("time series data is nil")
	}

	chainID := data.ChainID
	if chainID == "" {
		chainID = "global"
	}

	key := []byte(fmt.Sprintf("%s%s:%s:%d", timeSeriesPrefix, chainID, data.Metric, data.StartTime.Unix()))
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal time series data: %w", err)
	}

	if err := s.db.Set(key, jsonData, pebble.NoSync); err != nil {
		return fmt.Errorf("failed to save time series data: %w", err)
	}

	return nil
}

// GetTimeSeriesData retrieves time series data
func (s *PebbleStorage) GetTimeSeriesData(ctx context.Context, metric, chainID string, startTime, endTime time.Time) (*models.TimeSeriesData, error) {
	if chainID == "" {
		chainID = "global"
	}

	startKey := []byte(fmt.Sprintf("%s%s:%s:%d", timeSeriesPrefix, chainID, metric, startTime.Unix()))
	endKey := []byte(fmt.Sprintf("%s%s:%s:%d", timeSeriesPrefix, chainID, metric, endTime.Unix()))

	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: startKey,
		UpperBound: endKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Close()

	result := &models.TimeSeriesData{
		Metric:     metric,
		ChainID:    chainID,
		DataPoints: make([]*models.TimeSeriesDataPoint, 0),
		StartTime:  startTime,
		EndTime:    endTime,
	}

	for iter.First(); iter.Valid(); iter.Next() {
		var data models.TimeSeriesData
		if err := json.Unmarshal(iter.Value(), &data); err != nil {
			continue
		}
		result.DataPoints = append(result.DataPoints, data.DataPoints...)
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	return result, nil
}

// DeleteOldSnapshots deletes snapshots older than the specified time
func (s *PebbleStorage) DeleteOldSnapshots(ctx context.Context, before time.Time) error {
	prefix := []byte(snapshotPrefix)
	endKey := []byte(fmt.Sprintf("%s%d", snapshotPrefix, before.Unix()))

	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: endKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Close()

	batch := s.db.NewBatch()
	count := 0

	for iter.First(); iter.Valid(); iter.Next() {
		if err := batch.Delete(iter.Key(), pebble.NoSync); err != nil {
			return fmt.Errorf("failed to delete snapshot: %w", err)
		}
		count++

		// Commit in batches to avoid memory issues
		if count >= 1000 {
			if err := batch.Commit(pebble.Sync); err != nil {
				return fmt.Errorf("failed to commit batch: %w", err)
			}
			batch = s.db.NewBatch()
			count = 0
		}
	}

	if count > 0 {
		if err := batch.Commit(pebble.Sync); err != nil {
			return fmt.Errorf("failed to commit final batch: %w", err)
		}
	}

	return iter.Error()
}

// UpdateChainStatistic updates a specific statistic field for a chain
func (s *PebbleStorage) UpdateChainStatistic(ctx context.Context, chainID string, field string, value interface{}) error {
	stats, err := s.GetChainStatistics(ctx, chainID)
	if err != nil {
		if err == repository.ErrNotFound {
			// Create new statistics
			stats = &models.ChainStatistics{
				ChainID:     chainID,
				LastUpdated: time.Now(),
			}
		} else {
			return err
		}
	}

	// Update the specified field
	switch field {
	case "total_blocks":
		if v, ok := value.(uint64); ok {
			stats.TotalBlocks = v
		}
	case "total_transactions":
		if v, ok := value.(uint64); ok {
			stats.TotalTransactions = v
		}
	case "latest_block_number":
		if v, ok := value.(uint64); ok {
			stats.LatestBlockNumber = v
		}
	case "total_errors":
		if v, ok := value.(uint64); ok {
			stats.TotalErrors = v
		}
	default:
		return fmt.Errorf("unknown field: %s", field)
	}

	return s.SaveChainStatistics(ctx, stats)
}

// IncrementChainCounter increments a counter statistic for a chain
func (s *PebbleStorage) IncrementChainCounter(ctx context.Context, chainID string, counter string, delta uint64) error {
	stats, err := s.GetChainStatistics(ctx, chainID)
	if err != nil {
		if err == repository.ErrNotFound {
			stats = &models.ChainStatistics{
				ChainID:     chainID,
				LastUpdated: time.Now(),
			}
		} else {
			return err
		}
	}

	// Increment the specified counter
	switch counter {
	case "total_blocks":
		stats.TotalBlocks += delta
	case "total_transactions":
		stats.TotalTransactions += delta
	case "blocks_indexed":
		stats.BlocksIndexed += delta
	case "total_errors":
		stats.TotalErrors += delta
	default:
		return fmt.Errorf("unknown counter: %s", counter)
	}

	return s.SaveChainStatistics(ctx, stats)
}

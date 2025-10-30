package models

import (
	"fmt"
	"time"
)

// ChainStatistics represents statistics for a single blockchain
type ChainStatistics struct {
	ChainID           string    `json:"chain_id"`
	ChainType         string    `json:"chain_type"`
	ChainName         string    `json:"chain_name"`
	TotalBlocks       uint64    `json:"total_blocks"`
	TotalTransactions uint64    `json:"total_transactions"`
	AverageBlockTime  float64   `json:"average_block_time"`  // seconds
	AverageTxPerBlock float64   `json:"average_tx_per_block"`
	LatestBlockNumber uint64    `json:"latest_block_number"`
	OldestBlockNumber uint64    `json:"oldest_block_number"`
	FirstBlockTime    time.Time `json:"first_block_time"`
	LatestBlockTime   time.Time `json:"latest_block_time"`
	LastUpdated       time.Time `json:"last_updated"`

	// Indexing statistics
	IndexingStartTime time.Time `json:"indexing_start_time"`
	BlocksIndexed     uint64    `json:"blocks_indexed"`
	BlocksBehind      uint64    `json:"blocks_behind"`
	SyncProgress      float64   `json:"sync_progress"` // percentage
	IndexingRate      float64   `json:"indexing_rate"` // blocks per second

	// Error statistics
	TotalErrors       uint64    `json:"total_errors"`
	LastError         string    `json:"last_error,omitempty"`
	LastErrorTime     time.Time `json:"last_error_time,omitempty"`
}

// GlobalStatistics represents aggregated statistics across all chains
type GlobalStatistics struct {
	TotalChains          int       `json:"total_chains"`
	ActiveChains         int       `json:"active_chains"`
	TotalBlocks          uint64    `json:"total_blocks"`
	TotalTransactions    uint64    `json:"total_transactions"`
	AverageBlockTime     float64   `json:"average_block_time"`
	AverageTxPerBlock    float64   `json:"average_tx_per_block"`
	ChainsIndexed        []string  `json:"chains_indexed"`
	IndexingStartTime    time.Time `json:"indexing_start_time"`
	LastUpdated          time.Time `json:"last_updated"`

	// Overall system health
	TotalErrors          uint64    `json:"total_errors"`
	OverallSyncProgress  float64   `json:"overall_sync_progress"` // percentage
	AverageIndexingRate  float64   `json:"average_indexing_rate"` // blocks per second

	// Per-chain breakdown
	ChainStatistics      []*ChainStatistics `json:"chain_statistics,omitempty"`
}

// TimeSeriesDataPoint represents a single data point in a time series
type TimeSeriesDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	ChainID   string    `json:"chain_id,omitempty"`
}

// TimeSeriesData represents a collection of time series data points
type TimeSeriesData struct {
	Metric     string                  `json:"metric"`
	ChainID    string                  `json:"chain_id,omitempty"`
	DataPoints []*TimeSeriesDataPoint  `json:"data_points"`
	StartTime  time.Time               `json:"start_time"`
	EndTime    time.Time               `json:"end_time"`
}

// StatisticsSnapshot represents a point-in-time snapshot of statistics
type StatisticsSnapshot struct {
	Timestamp   time.Time         `json:"timestamp"`
	ChainStats  *ChainStatistics  `json:"chain_stats,omitempty"`
	GlobalStats *GlobalStatistics `json:"global_stats,omitempty"`
}

// Validate validates chain statistics
func (cs *ChainStatistics) Validate() error {
	if cs.ChainID == "" {
		return fmt.Errorf("chain ID is required")
	}
	if cs.ChainType == "" {
		return fmt.Errorf("chain type is required")
	}
	if cs.TotalBlocks > 0 && cs.TotalTransactions > 0 {
		if cs.AverageTxPerBlock < 0 {
			return fmt.Errorf("average tx per block cannot be negative")
		}
	}
	if cs.AverageBlockTime < 0 {
		return fmt.Errorf("average block time cannot be negative")
	}
	if cs.LatestBlockNumber < cs.OldestBlockNumber {
		return fmt.Errorf("latest block cannot be less than oldest block")
	}
	return nil
}

// CalculateAverages calculates average statistics
func (cs *ChainStatistics) CalculateAverages() {
	if cs.TotalBlocks > 0 {
		cs.AverageTxPerBlock = float64(cs.TotalTransactions) / float64(cs.TotalBlocks)
	}

	if !cs.FirstBlockTime.IsZero() && !cs.LatestBlockTime.IsZero() && cs.TotalBlocks > 1 {
		duration := cs.LatestBlockTime.Sub(cs.FirstBlockTime).Seconds()
		if duration > 0 {
			cs.AverageBlockTime = duration / float64(cs.TotalBlocks-1)
		}
	}

	if !cs.IndexingStartTime.IsZero() && cs.BlocksIndexed > 0 {
		duration := time.Since(cs.IndexingStartTime).Seconds()
		if duration > 0 {
			cs.IndexingRate = float64(cs.BlocksIndexed) / duration
		}
	}
}

// String returns a string representation of chain statistics
func (cs *ChainStatistics) String() string {
	return fmt.Sprintf(
		"ChainStats[%s: blocks=%d, txs=%d, avg_block_time=%.2fs, avg_tx=%d, progress=%.2f%%]",
		cs.ChainID,
		cs.TotalBlocks,
		cs.TotalTransactions,
		cs.AverageBlockTime,
		int(cs.AverageTxPerBlock),
		cs.SyncProgress,
	)
}

// Validate validates global statistics
func (gs *GlobalStatistics) Validate() error {
	if gs.TotalChains < 0 {
		return fmt.Errorf("total chains cannot be negative")
	}
	if gs.ActiveChains < 0 || gs.ActiveChains > gs.TotalChains {
		return fmt.Errorf("active chains must be between 0 and total chains")
	}
	if gs.AverageBlockTime < 0 {
		return fmt.Errorf("average block time cannot be negative")
	}
	if gs.AverageTxPerBlock < 0 {
		return fmt.Errorf("average tx per block cannot be negative")
	}
	return nil
}

// CalculateAverages calculates global average statistics
func (gs *GlobalStatistics) CalculateAverages() {
	if gs.TotalBlocks > 0 {
		gs.AverageTxPerBlock = float64(gs.TotalTransactions) / float64(gs.TotalBlocks)
	}

	// Calculate weighted average block time and indexing rate
	if len(gs.ChainStatistics) > 0 {
		totalBlockTime := 0.0
		totalIndexingRate := 0.0
		totalBlocks := uint64(0)
		totalSyncProgress := 0.0
		activeChains := 0

		for _, cs := range gs.ChainStatistics {
			if cs.TotalBlocks > 0 {
				totalBlockTime += cs.AverageBlockTime * float64(cs.TotalBlocks)
				totalBlocks += cs.TotalBlocks
				totalIndexingRate += cs.IndexingRate
				totalSyncProgress += cs.SyncProgress
				activeChains++
			}
		}

		if totalBlocks > 0 {
			gs.AverageBlockTime = totalBlockTime / float64(totalBlocks)
		}

		if activeChains > 0 {
			gs.AverageIndexingRate = totalIndexingRate / float64(activeChains)
			gs.OverallSyncProgress = totalSyncProgress / float64(activeChains)
			gs.ActiveChains = activeChains
		}
	}
}

// String returns a string representation of global statistics
func (gs *GlobalStatistics) String() string {
	return fmt.Sprintf(
		"GlobalStats[chains=%d/%d, blocks=%d, txs=%d, avg_block_time=%.2fs, progress=%.2f%%]",
		gs.ActiveChains,
		gs.TotalChains,
		gs.TotalBlocks,
		gs.TotalTransactions,
		gs.AverageBlockTime,
		gs.OverallSyncProgress,
	)
}

// MergeChainStatistics merges chain statistics into global statistics
func (gs *GlobalStatistics) MergeChainStatistics(chainStats []*ChainStatistics) {
	gs.ChainStatistics = chainStats
	gs.TotalChains = len(chainStats)
	gs.TotalBlocks = 0
	gs.TotalTransactions = 0
	gs.TotalErrors = 0
	gs.ChainsIndexed = make([]string, 0, len(chainStats))

	for _, cs := range chainStats {
		gs.TotalBlocks += cs.TotalBlocks
		gs.TotalTransactions += cs.TotalTransactions
		gs.TotalErrors += cs.TotalErrors
		gs.ChainsIndexed = append(gs.ChainsIndexed, cs.ChainID)
	}

	gs.CalculateAverages()
	gs.LastUpdated = time.Now()
}

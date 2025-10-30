package repository

import (
	"context"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// StatisticsRepository defines the interface for statistics storage operations
type StatisticsRepository interface {
	// SaveChainStatistics saves chain statistics
	SaveChainStatistics(ctx context.Context, stats *models.ChainStatistics) error

	// GetChainStatistics retrieves chain statistics by chain ID
	GetChainStatistics(ctx context.Context, chainID string) (*models.ChainStatistics, error)

	// GetAllChainStatistics retrieves statistics for all chains
	GetAllChainStatistics(ctx context.Context) ([]*models.ChainStatistics, error)

	// SaveGlobalStatistics saves global statistics
	SaveGlobalStatistics(ctx context.Context, stats *models.GlobalStatistics) error

	// GetGlobalStatistics retrieves global statistics
	GetGlobalStatistics(ctx context.Context) (*models.GlobalStatistics, error)

	// SaveStatisticsSnapshot saves a statistics snapshot
	SaveStatisticsSnapshot(ctx context.Context, snapshot *models.StatisticsSnapshot) error

	// GetStatisticsSnapshots retrieves statistics snapshots within a time range
	GetStatisticsSnapshots(ctx context.Context, chainID string, startTime, endTime time.Time) ([]*models.StatisticsSnapshot, error)

	// SaveTimeSeriesData saves time series data
	SaveTimeSeriesData(ctx context.Context, data *models.TimeSeriesData) error

	// GetTimeSeriesData retrieves time series data
	GetTimeSeriesData(ctx context.Context, metric, chainID string, startTime, endTime time.Time) (*models.TimeSeriesData, error)

	// DeleteOldSnapshots deletes snapshots older than the specified time
	DeleteOldSnapshots(ctx context.Context, before time.Time) error

	// UpdateChainStatistic updates a specific statistic field for a chain
	UpdateChainStatistic(ctx context.Context, chainID string, field string, value interface{}) error

	// IncrementChainCounter increments a counter statistic for a chain
	IncrementChainCounter(ctx context.Context, chainID string, counter string, delta uint64) error
}

package pebble

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
)

// PebbleStorage implements the Storage interface using PebbleDB
// This is the main storage implementation that composes all repositories
type PebbleStorage struct {
	db         *pebble.DB
	encoder    *Encoder
	path       string
	statsCache *statsCache

	// Embedded repository implementations
	*BlockRepo
	*TransactionRepo
	*ChainRepo
}

// Config holds PebbleDB configuration
type Config struct {
	Path string // Path to the database directory

	// PebbleDB options
	CacheSize         int64 // Cache size in bytes (default: 64MB)
	MaxOpenFiles      int   // Maximum number of open files (default: 1000)
	WriteBufferSize   int   // Write buffer size in bytes (default: 64MB)
	MaxConcurrentMem  int   // Maximum concurrent memtables (default: 2)
	DisableWAL        bool  // Disable write-ahead log (default: false)
	BytesPerSync      int   // Bytes to write before syncing (default: 512KB)

	// Custom logger
	Logger pebble.Logger
}

// DefaultConfig returns the default configuration
func DefaultConfig(path string) *Config {
	return &Config{
		Path:             path,
		CacheSize:        64 << 20, // 64MB
		MaxOpenFiles:     1000,
		WriteBufferSize:  64 << 20, // 64MB
		MaxConcurrentMem: 2,
		DisableWAL:       false,
		BytesPerSync:     512 << 10, // 512KB
	}
}

// HighPerformanceConfig returns a configuration optimized for high performance
// Recommended for systems with ample memory and fast storage
func HighPerformanceConfig(path string) *Config {
	return &Config{
		Path:             path,
		CacheSize:        256 << 20, // 256MB - larger cache for better read performance
		MaxOpenFiles:     5000,      // More open files for concurrent access
		WriteBufferSize:  128 << 20, // 128MB - larger buffer for write throughput
		MaxConcurrentMem: 4,         // More memtables for write concurrency
		DisableWAL:       false,
		BytesPerSync:     1 << 20, // 1MB - less frequent sync for better write performance
	}
}

// LowMemoryConfig returns a configuration for memory-constrained environments
func LowMemoryConfig(path string) *Config {
	return &Config{
		Path:             path,
		CacheSize:        16 << 20, // 16MB
		MaxOpenFiles:     100,
		WriteBufferSize:  16 << 20, // 16MB
		MaxConcurrentMem: 1,
		DisableWAL:       false,
		BytesPerSync:     256 << 10, // 256KB
	}
}

// NewStorage creates a new PebbleDB storage instance
func NewStorage(config *Config) (*PebbleStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.Path == "" {
		return nil, fmt.Errorf("database path cannot be empty")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(config.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Configure PebbleDB options
	opts := &pebble.Options{
		Cache:                       pebble.NewCache(config.CacheSize),
		MaxOpenFiles:                config.MaxOpenFiles,
		MemTableSize:                uint64(config.WriteBufferSize),
		MemTableStopWritesThreshold: config.MaxConcurrentMem,
		BytesPerSync:                config.BytesPerSync,
		DisableWAL:                  config.DisableWAL,
		Logger:                      config.Logger,
	}

	// Open the database
	db, err := pebble.Open(config.Path, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create encoder
	encoder := NewEncoder()

	// Create storage instance
	storage := &PebbleStorage{
		db:         db,
		encoder:    encoder,
		path:       config.Path,
		statsCache: newStatsCache(5 * time.Minute), // 5 minute cache TTL
	}

	// Initialize repositories
	storage.BlockRepo = NewBlockRepo(db, encoder)
	storage.TransactionRepo = NewTransactionRepo(db, encoder)
	storage.ChainRepo = NewChainRepo(db, encoder)

	return storage, nil
}

// Close closes the database connection
func (s *PebbleStorage) Close() error {
	if s.db == nil {
		return nil
	}

	// Flush any pending writes
	if err := s.db.Flush(); err != nil {
		return fmt.Errorf("failed to flush database: %w", err)
	}

	// Close the database
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	s.db = nil
	return nil
}

// NewBatch creates a new batch for atomic operations
func (s *PebbleStorage) NewBatch() repository.Batch {
	return NewBatch(s.db, s.encoder)
}

// GetStats returns storage statistics
func (s *PebbleStorage) GetStats(ctx context.Context) (*repository.StorageStats, error) {
	stats := &repository.StorageStats{
		ChainStats: make(map[string]*repository.ChainStorageStats),
	}

	// Get all chains
	chains, err := s.GetAllChains(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chains: %w", err)
	}

	// Calculate stats for each chain
	for _, chain := range chains {
		chainStats, err := s.GetChainStats(ctx, chain.ChainID)
		if err != nil {
			return nil, fmt.Errorf("failed to get chain stats for %s: %w", chain.ChainID, err)
		}

		stats.TotalBlocks += chainStats.TotalBlocks
		stats.TotalTransactions += chainStats.TotalTransactions

		stats.ChainStats[chain.ChainID] = &repository.ChainStorageStats{
			ChainID:          chain.ChainID,
			BlockCount:       chainStats.TotalBlocks,
			TransactionCount: chainStats.TotalTransactions,
			LatestBlock:      chainStats.LatestIndexedBlock,
		}
	}

	// Get disk usage
	if metrics := s.db.Metrics(); metrics != nil {
		// Calculate approximate disk usage from LSM levels
		for i := 0; i < len(metrics.Levels); i++ {
			stats.DiskUsage += uint64(metrics.Levels[i].Size)
		}
	}

	return stats, nil
}

// Compact performs manual compaction on a key range
// This is useful for optimizing storage after bulk operations
func (s *PebbleStorage) Compact(start, end []byte) error {
	if s.db == nil {
		return fmt.Errorf("database is closed")
	}

	return s.db.Compact(start, end, true)
}

// CompactAll performs full database compaction
func (s *PebbleStorage) CompactAll() error {
	// Use empty byte slices for full range compaction
	return s.Compact([]byte{}, []byte{0xff, 0xff, 0xff, 0xff})
}

// Flush flushes pending writes to disk
func (s *PebbleStorage) Flush() error {
	if s.db == nil {
		return fmt.Errorf("database is closed")
	}

	return s.db.Flush()
}

// Metrics returns database metrics
func (s *PebbleStorage) Metrics() *pebble.Metrics {
	if s.db == nil {
		return nil
	}

	return s.db.Metrics()
}

// Path returns the database path
func (s *PebbleStorage) Path() string {
	return s.path
}

package pebble

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// Helper function to create a temporary test database
func setupTestDB(t *testing.T) (*PebbleStorage, string) {
	t.Helper()

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "pebble-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create storage configuration
	config := DefaultConfig(tmpDir)
	config.DisableWAL = false // Keep WAL enabled for proper testing

	// Create storage
	storage, err := NewStorage(config)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	return storage, tmpDir
}

// Helper function to cleanup test database
func cleanupTestDB(t *testing.T, storage *PebbleStorage, tmpDir string) {
	t.Helper()

	if storage != nil {
		if err := storage.Close(); err != nil {
			t.Errorf("failed to close storage: %v", err)
		}
	}

	if tmpDir != "" {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("failed to remove temp dir: %v", err)
		}
	}
}

func TestNewStorage(t *testing.T) {
	t.Run("create storage with default config", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "pebble-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		config := DefaultConfig(tmpDir)
		storage, err := NewStorage(config)
		if err != nil {
			t.Fatalf("NewStorage() error = %v", err)
		}
		defer storage.Close()

		if storage == nil {
			t.Error("NewStorage() returned nil storage")
		}

		if storage.db == nil {
			t.Error("storage.db is nil")
		}

		if storage.encoder == nil {
			t.Error("storage.encoder is nil")
		}

		if storage.BlockRepo == nil {
			t.Error("storage.BlockRepo is nil")
		}

		if storage.TransactionRepo == nil {
			t.Error("storage.TransactionRepo is nil")
		}

		if storage.ChainRepo == nil {
			t.Error("storage.ChainRepo is nil")
		}
	})

	t.Run("create storage with nil config", func(t *testing.T) {
		_, err := NewStorage(nil)
		if err == nil {
			t.Error("NewStorage(nil) should return error")
		}
	})

	t.Run("create storage with empty path", func(t *testing.T) {
		config := &Config{Path: ""}
		_, err := NewStorage(config)
		if err == nil {
			t.Error("NewStorage() with empty path should return error")
		}
	})

	t.Run("create storage in non-existent directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "pebble-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		dbPath := filepath.Join(tmpDir, "nested", "path", "db")
		config := DefaultConfig(dbPath)
		storage, err := NewStorage(config)
		if err != nil {
			t.Fatalf("NewStorage() should create nested directories, got error: %v", err)
		}
		defer storage.Close()

		// Verify directory was created
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Error("NewStorage() should create nested directories")
		}
	})
}

func TestStorage_Close(t *testing.T) {
	t.Run("close storage successfully", func(t *testing.T) {
		storage, tmpDir := setupTestDB(t)
		defer os.RemoveAll(tmpDir)

		err := storage.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}

		// Verify db is nil after close
		if storage.db != nil {
			t.Error("storage.db should be nil after Close()")
		}
	})

	t.Run("close storage twice", func(t *testing.T) {
		storage, tmpDir := setupTestDB(t)
		defer os.RemoveAll(tmpDir)

		// First close
		if err := storage.Close(); err != nil {
			t.Errorf("first Close() error = %v", err)
		}

		// Second close should not error
		if err := storage.Close(); err != nil {
			t.Errorf("second Close() error = %v", err)
		}
	})
}

func TestStorage_NewBatch(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	batch := storage.NewBatch()
	if batch == nil {
		t.Error("NewBatch() returned nil")
	}

	// Verify batch can be closed
	if err := batch.Close(); err != nil {
		t.Errorf("batch.Close() error = %v", err)
	}
}

func TestStorage_GetStats(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	t.Run("get stats for empty storage", func(t *testing.T) {
		stats, err := storage.GetStats(ctx)
		if err != nil {
			t.Fatalf("GetStats() error = %v", err)
		}

		if stats == nil {
			t.Fatal("GetStats() returned nil stats")
		}

		if stats.TotalBlocks != 0 {
			t.Errorf("TotalBlocks = %v, want 0", stats.TotalBlocks)
		}

		if stats.TotalTransactions != 0 {
			t.Errorf("TotalTransactions = %v, want 0", stats.TotalTransactions)
		}

		if stats.ChainStats == nil {
			t.Error("ChainStats should not be nil")
		}
	})

	t.Run("get stats after adding data", func(t *testing.T) {
		// Add a chain
		chain := models.NewChain(models.ChainTypeEVM, "ethereum", "Ethereum")
		chain.RPCEndpoints = []string{"http://localhost:8545"}
		if err := storage.SaveChain(ctx, chain); err != nil {
			t.Fatalf("SaveChain() error = %v", err)
		}

		// Add a block
		block := models.NewBlock(models.ChainTypeEVM, "ethereum", 1, "0xabc")
		if err := storage.SaveBlock(ctx, block); err != nil {
			t.Fatalf("SaveBlock() error = %v", err)
		}

		// Get stats
		stats, err := storage.GetStats(ctx)
		if err != nil {
			t.Fatalf("GetStats() error = %v", err)
		}

		if len(stats.ChainStats) == 0 {
			t.Error("ChainStats should contain ethereum chain")
		}

		if chainStats, exists := stats.ChainStats["ethereum"]; exists {
			if chainStats.ChainID != "ethereum" {
				t.Errorf("ChainID = %v, want ethereum", chainStats.ChainID)
			}
		} else {
			t.Error("ethereum chain not found in stats")
		}
	})
}

func TestStorage_Flush(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	err := storage.Flush()
	if err != nil {
		t.Errorf("Flush() error = %v", err)
	}
}

func TestStorage_Compact(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	t.Run("compact specific range", func(t *testing.T) {
		start := []byte("block:ethereum:0")
		end := []byte("block:ethereum:1000")

		err := storage.Compact(start, end)
		if err != nil {
			t.Errorf("Compact() error = %v", err)
		}
	})

	t.Run("compact all", func(t *testing.T) {
		err := storage.CompactAll()
		if err != nil {
			t.Errorf("CompactAll() error = %v", err)
		}
	})
}

func TestStorage_Metrics(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	metrics := storage.Metrics()
	if metrics == nil {
		t.Error("Metrics() returned nil")
	}
}

func TestStorage_Path(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultConfig(tmpDir)
	storage, err := NewStorage(config)
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}
	defer storage.Close()

	path := storage.Path()
	if path != tmpDir {
		t.Errorf("Path() = %v, want %v", path, tmpDir)
	}
}

func TestDefaultConfig(t *testing.T) {
	path := "/tmp/test-db"
	config := DefaultConfig(path)

	if config.Path != path {
		t.Errorf("Path = %v, want %v", config.Path, path)
	}

	if config.CacheSize != 64<<20 {
		t.Errorf("CacheSize = %v, want %v", config.CacheSize, 64<<20)
	}

	if config.MaxOpenFiles != 1000 {
		t.Errorf("MaxOpenFiles = %v, want %v", config.MaxOpenFiles, 1000)
	}

	if config.WriteBufferSize != 64<<20 {
		t.Errorf("WriteBufferSize = %v, want %v", config.WriteBufferSize, 64<<20)
	}

	if config.MaxConcurrentMem != 2 {
		t.Errorf("MaxConcurrentMem = %v, want %v", config.MaxConcurrentMem, 2)
	}

	if config.DisableWAL != false {
		t.Errorf("DisableWAL = %v, want false", config.DisableWAL)
	}

	if config.BytesPerSync != 512<<10 {
		t.Errorf("BytesPerSync = %v, want %v", config.BytesPerSync, 512<<10)
	}
}

// Benchmark tests
func BenchmarkStorage_SaveBlock(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "pebble-bench-*")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultConfig(tmpDir)
	storage, err := NewStorage(config)
	if err != nil {
		b.Fatalf("NewStorage() error = %v", err)
	}
	defer storage.Close()

	ctx := context.Background()
	block := models.NewBlock(models.ChainTypeEVM, "ethereum", 1, "0xabc")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block.Number = uint64(i)
		_ = storage.SaveBlock(ctx, block)
	}
}

func BenchmarkStorage_GetBlock(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "pebble-bench-*")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultConfig(tmpDir)
	storage, err := NewStorage(config)
	if err != nil {
		b.Fatalf("NewStorage() error = %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Setup: save some blocks
	for i := 0; i < 1000; i++ {
		block := models.NewBlock(models.ChainTypeEVM, "ethereum", uint64(i), "0xabc")
		if err := storage.SaveBlock(ctx, block); err != nil {
			b.Fatalf("SaveBlock() error = %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetBlock(ctx, "ethereum", uint64(i%1000))
	}
}

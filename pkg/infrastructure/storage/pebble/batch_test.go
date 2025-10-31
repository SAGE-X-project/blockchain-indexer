package pebble

import (
	"context"
	"testing"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

func TestBatch_SetBlock(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()
	batch := storage.NewBatch()
	defer batch.Close()

	t.Run("add block to batch", func(t *testing.T) {
		block := models.NewBlock(models.ChainTypeEVM, "ethereum", 1, "0xabc")

		err := batch.SetBlock(ctx, block)
		if err != nil {
			t.Fatalf("SetBlock() error = %v", err)
		}

		if batch.Count() == 0 {
			t.Error("batch count should be > 0 after SetBlock")
		}
	})

	t.Run("add nil block to batch", func(t *testing.T) {
		err := batch.SetBlock(ctx, nil)
		if err == nil {
			t.Error("SetBlock(nil) should return error")
		}
	})

	t.Run("add invalid block to batch", func(t *testing.T) {
		block := &models.Block{
			ChainType: "invalid",
			ChainID:   "ethereum",
		}

		err := batch.SetBlock(ctx, block)
		if err == nil {
			t.Error("SetBlock() with invalid block should return error")
		}
	})
}

func TestBatch_SetBlocks(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()
	batch := storage.NewBatch()
	defer batch.Close()

	t.Run("add multiple blocks to batch", func(t *testing.T) {
		blocks := make([]*models.Block, 5)
		for i := 0; i < 5; i++ {
			blocks[i] = models.NewBlock(models.ChainTypeEVM, "ethereum", uint64(i), "0xhash")
		}

		err := batch.SetBlocks(ctx, blocks)
		if err != nil {
			t.Fatalf("SetBlocks() error = %v", err)
		}

		if batch.Count() == 0 {
			t.Error("batch count should be > 0 after SetBlocks")
		}
	})

	t.Run("add empty blocks slice", func(t *testing.T) {
		err := batch.SetBlocks(ctx, []*models.Block{})
		if err != nil {
			t.Errorf("SetBlocks() with empty slice should not error, got: %v", err)
		}
	})
}

func TestBatch_SetTransaction(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()
	batch := storage.NewBatch()
	defer batch.Close()

	t.Run("add transaction to batch", func(t *testing.T) {
		tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx123")
		tx.From = "0xfrom"
		tx.To = "0xto"

		err := batch.SetTransaction(ctx, tx)
		if err != nil {
			t.Fatalf("SetTransaction() error = %v", err)
		}

		if batch.Count() == 0 {
			t.Error("batch count should be > 0 after SetTransaction")
		}
	})

	t.Run("add nil transaction to batch", func(t *testing.T) {
		err := batch.SetTransaction(ctx, nil)
		if err == nil {
			t.Error("SetTransaction(nil) should return error")
		}
	})

	t.Run("add invalid transaction to batch", func(t *testing.T) {
		tx := &models.Transaction{
			ChainType: "invalid",
			ChainID:   "ethereum",
			Hash:      "0xtx",
		}

		err := batch.SetTransaction(ctx, tx)
		if err == nil {
			t.Error("SetTransaction() with invalid transaction should return error")
		}
	})
}

func TestBatch_SetTransactions(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()
	batch := storage.NewBatch()
	defer batch.Close()

	t.Run("add multiple transactions to batch", func(t *testing.T) {
		txs := make([]*models.Transaction, 5)
		for i := 0; i < 5; i++ {
			tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx"+string(rune('0'+i)))
			tx.From = "0xfrom"
			tx.Index = uint64(i)
			txs[i] = tx
		}

		err := batch.SetTransactions(ctx, txs)
		if err != nil {
			t.Fatalf("SetTransactions() error = %v", err)
		}

		if batch.Count() == 0 {
			t.Error("batch count should be > 0 after SetTransactions")
		}
	})

	t.Run("add empty transactions slice", func(t *testing.T) {
		err := batch.SetTransactions(ctx, []*models.Transaction{})
		if err != nil {
			t.Errorf("SetTransactions() with empty slice should not error, got: %v", err)
		}
	})
}

func TestBatch_Commit(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	t.Run("commit batch with blocks", func(t *testing.T) {
		batch := storage.NewBatch()
		defer batch.Close()

		// Add blocks to batch
		for i := 0; i < 10; i++ {
			block := models.NewBlock(models.ChainTypeEVM, "ethereum", uint64(i), "0xhash")
			if err := batch.SetBlock(ctx, block); err != nil {
				t.Fatalf("SetBlock() error = %v", err)
			}
		}

		// Commit
		err := batch.Commit()
		if err != nil {
			t.Fatalf("Commit() error = %v", err)
		}

		// Verify blocks were committed
		for i := 0; i < 10; i++ {
			exists, err := storage.HasBlock(ctx, "ethereum", uint64(i))
			if err != nil {
				t.Fatalf("HasBlock() error = %v", err)
			}
			if !exists {
				t.Errorf("block %d not found after commit", i)
			}
		}

		// Verify batch count is reset after commit
		if batch.Count() != 0 {
			t.Errorf("batch count = %v, want 0 after commit", batch.Count())
		}
	})

	t.Run("commit batch with transactions", func(t *testing.T) {
		batch := storage.NewBatch()
		defer batch.Close()

		// Add transactions to batch
		for i := 0; i < 10; i++ {
			tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx"+string(rune('0'+i)))
			tx.From = "0xfrom"
			tx.Index = uint64(i)
			if err := batch.SetTransaction(ctx, tx); err != nil {
				t.Fatalf("SetTransaction() error = %v", err)
			}
		}

		// Commit
		err := batch.Commit()
		if err != nil {
			t.Fatalf("Commit() error = %v", err)
		}

		// Verify transactions were committed
		for i := 0; i < 10; i++ {
			exists, err := storage.HasTransaction(ctx, "ethereum", "0xtx"+string(rune('0'+i)))
			if err != nil {
				t.Fatalf("HasTransaction() error = %v", err)
			}
			if !exists {
				t.Errorf("transaction %d not found after commit", i)
			}
		}
	})

	t.Run("commit empty batch", func(t *testing.T) {
		batch := storage.NewBatch()
		defer batch.Close()

		err := batch.Commit()
		if err != nil {
			t.Errorf("Commit() on empty batch should not error, got: %v", err)
		}
	})
}

func TestBatch_Reset(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()
	batch := storage.NewBatch()
	defer batch.Close()

	// Add some operations
	block := models.NewBlock(models.ChainTypeEVM, "ethereum", 1, "0xhash")
	if err := batch.SetBlock(ctx, block); err != nil {
		t.Fatalf("SetBlock() error = %v", err)
	}

	initialCount := batch.Count()
	if initialCount == 0 {
		t.Fatal("batch should have operations before Reset")
	}

	// Reset
	batch.Reset()

	// Verify batch is empty
	if batch.Count() != 0 {
		t.Errorf("batch count = %v, want 0 after Reset", batch.Count())
	}
}

func TestBatch_Count(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()
	batch := storage.NewBatch()
	defer batch.Close()

	t.Run("count starts at zero", func(t *testing.T) {
		if batch.Count() != 0 {
			t.Errorf("Count() = %v, want 0", batch.Count())
		}
	})

	t.Run("count increases with operations", func(t *testing.T) {
		block := models.NewBlock(models.ChainTypeEVM, "ethereum", 1, "0xhash")
		if err := batch.SetBlock(ctx, block); err != nil {
			t.Fatalf("SetBlock() error = %v", err)
		}

		if batch.Count() == 0 {
			t.Error("Count() should be > 0 after SetBlock")
		}

		tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx")
		tx.From = "0xfrom"
		if err := batch.SetTransaction(ctx, tx); err != nil {
			t.Fatalf("SetTransaction() error = %v", err)
		}

		if batch.Count() == 0 {
			t.Error("Count() should be > 0 after SetTransaction")
		}
	})
}

func TestBatch_Close(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	t.Run("close batch without commit", func(t *testing.T) {
		batch := storage.NewBatch()

		// Add operations
		block := models.NewBlock(models.ChainTypeEVM, "ethereum", 1, "0xhash")
		if err := batch.SetBlock(ctx, block); err != nil {
			t.Fatalf("SetBlock() error = %v", err)
		}

		// Close without commit
		err := batch.Close()
		if err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		// Verify block was not committed
		exists, err := storage.HasBlock(ctx, "ethereum", 1)
		if err != nil {
			t.Fatalf("HasBlock() error = %v", err)
		}
		if exists {
			t.Error("block should not exist after Close without Commit")
		}
	})

	t.Run("close batch twice", func(t *testing.T) {
		batch := storage.NewBatch()

		// First close
		if err := batch.Close(); err != nil {
			t.Fatalf("first Close() error = %v", err)
		}

		// Second close
		if err := batch.Close(); err != nil {
			t.Errorf("second Close() should not error, got: %v", err)
		}
	})
}

func TestBatch_AtomicCommit(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	t.Run("all operations commit atomically", func(t *testing.T) {
		batch := storage.NewBatch()
		defer batch.Close()

		// Add multiple blocks
		for i := 0; i < 100; i++ {
			block := models.NewBlock(models.ChainTypeEVM, "ethereum", uint64(i), "0xhash")
			if err := batch.SetBlock(ctx, block); err != nil {
				t.Fatalf("SetBlock() error = %v", err)
			}
		}

		// Add multiple transactions
		for i := 0; i < 100; i++ {
			tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx"+string(rune(i)))
			tx.From = "0xfrom"
			tx.Index = uint64(i)
			if err := batch.SetTransaction(ctx, tx); err != nil {
				t.Fatalf("SetTransaction() error = %v", err)
			}
		}

		// Commit all at once
		err := batch.Commit()
		if err != nil {
			t.Fatalf("Commit() error = %v", err)
		}

		// Verify all blocks exist
		for i := 0; i < 100; i++ {
			exists, err := storage.HasBlock(ctx, "ethereum", uint64(i))
			if err != nil {
				t.Fatalf("HasBlock() error = %v", err)
			}
			if !exists {
				t.Errorf("block %d not found", i)
			}
		}

		// Verify all transactions exist
		for i := 0; i < 100; i++ {
			exists, err := storage.HasTransaction(ctx, "ethereum", "0xtx"+string(rune(i)))
			if err != nil {
				t.Fatalf("HasTransaction() error = %v", err)
			}
			if !exists {
				t.Errorf("transaction %d not found", i)
			}
		}
	})
}

// Benchmark tests
func BenchmarkBatch_SetBlock(b *testing.B) {
	storage, tmpDir := setupTestDB(&testing.T{})
	defer cleanupTestDB(&testing.T{}, storage, tmpDir)

	ctx := context.Background()
	batch := storage.NewBatch()
	defer batch.Close()

	block := models.NewBlock(models.ChainTypeEVM, "ethereum", 1, "0xhash")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block.Number = uint64(i)
		_ = batch.SetBlock(ctx, block)
	}
}

func BenchmarkBatch_Commit(b *testing.B) {
	storage, tmpDir := setupTestDB(&testing.T{})
	defer cleanupTestDB(&testing.T{}, storage, tmpDir)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		batch := storage.NewBatch()

		// Add 100 blocks
		for j := 0; j < 100; j++ {
			block := models.NewBlock(models.ChainTypeEVM, "ethereum", uint64(i*100+j), "0xhash")
			_ = batch.SetBlock(ctx, block)
		}

		b.StartTimer()
		_ = batch.Commit()
		batch.Close()
	}
}

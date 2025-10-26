package pebble

import (
	"context"
	"testing"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
)

func TestBlockRepo_SaveBlock(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	t.Run("save valid block", func(t *testing.T) {
		block := models.NewBlock(models.ChainTypeEVM, "ethereum", 1, "0xabc123")
		block.ParentHash = "0xparent"
		block.Proposer = "0xminer"

		err := storage.SaveBlock(ctx, block)
		if err != nil {
			t.Fatalf("SaveBlock() error = %v", err)
		}

		// Verify block was saved
		retrieved, err := storage.GetBlock(ctx, "ethereum", 1)
		if err != nil {
			t.Fatalf("GetBlock() error = %v", err)
		}

		if retrieved.Hash != block.Hash {
			t.Errorf("Hash = %v, want %v", retrieved.Hash, block.Hash)
		}

		if retrieved.ParentHash != block.ParentHash {
			t.Errorf("ParentHash = %v, want %v", retrieved.ParentHash, block.ParentHash)
		}
	})

	t.Run("save nil block", func(t *testing.T) {
		err := storage.SaveBlock(ctx, nil)
		if err == nil {
			t.Error("SaveBlock(nil) should return error")
		}
	})

	t.Run("save invalid block", func(t *testing.T) {
		block := &models.Block{
			ChainType: "invalid",
			ChainID:   "ethereum",
			Number:    1,
			Hash:      "0xabc",
		}

		err := storage.SaveBlock(ctx, block)
		if err == nil {
			t.Error("SaveBlock() with invalid block should return error")
		}
	})

	t.Run("update latest height", func(t *testing.T) {
		// Save block 100
		block := models.NewBlock(models.ChainTypeEVM, "ethereum", 100, "0x100")
		if err := storage.SaveBlock(ctx, block); err != nil {
			t.Fatalf("SaveBlock() error = %v", err)
		}

		// Verify latest height is 100
		height, err := storage.GetLatestHeight(ctx, "ethereum")
		if err != nil {
			t.Fatalf("GetLatestHeight() error = %v", err)
		}

		if height != 100 {
			t.Errorf("LatestHeight = %v, want 100", height)
		}
	})
}

func TestBlockRepo_GetBlock(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save a block
	block := models.NewBlock(models.ChainTypeEVM, "ethereum", 12345, "0xabc123")
	block.ParentHash = "0xparent"
	if err := storage.SaveBlock(ctx, block); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("get existing block", func(t *testing.T) {
		retrieved, err := storage.GetBlock(ctx, "ethereum", 12345)
		if err != nil {
			t.Fatalf("GetBlock() error = %v", err)
		}

		if retrieved == nil {
			t.Fatal("GetBlock() returned nil")
		}

		if retrieved.Number != 12345 {
			t.Errorf("Number = %v, want 12345", retrieved.Number)
		}

		if retrieved.Hash != "0xabc123" {
			t.Errorf("Hash = %v, want 0xabc123", retrieved.Hash)
		}
	})

	t.Run("get non-existent block", func(t *testing.T) {
		_, err := storage.GetBlock(ctx, "ethereum", 99999)
		if err != repository.ErrBlockNotFound {
			t.Errorf("GetBlock() error = %v, want ErrBlockNotFound", err)
		}
	})

	t.Run("get block from non-existent chain", func(t *testing.T) {
		_, err := storage.GetBlock(ctx, "nonexistent", 1)
		if err != repository.ErrBlockNotFound {
			t.Errorf("GetBlock() error = %v, want ErrBlockNotFound", err)
		}
	})
}

func TestBlockRepo_GetBlockByHash(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save a block
	block := models.NewBlock(models.ChainTypeEVM, "ethereum", 12345, "0xhash123")
	if err := storage.SaveBlock(ctx, block); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("get block by existing hash", func(t *testing.T) {
		retrieved, err := storage.GetBlockByHash(ctx, "ethereum", "0xhash123")
		if err != nil {
			t.Fatalf("GetBlockByHash() error = %v", err)
		}

		if retrieved.Number != 12345 {
			t.Errorf("Number = %v, want 12345", retrieved.Number)
		}

		if retrieved.Hash != "0xhash123" {
			t.Errorf("Hash = %v, want 0xhash123", retrieved.Hash)
		}
	})

	t.Run("get block by non-existent hash", func(t *testing.T) {
		_, err := storage.GetBlockByHash(ctx, "ethereum", "0xnonexistent")
		if err != repository.ErrBlockNotFound {
			t.Errorf("GetBlockByHash() error = %v, want ErrBlockNotFound", err)
		}
	})
}

func TestBlockRepo_GetBlocks(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save blocks 10-20
	for i := 10; i <= 20; i++ {
		block := models.NewBlock(models.ChainTypeEVM, "ethereum", uint64(i), "0xhash")
		if err := storage.SaveBlock(ctx, block); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	t.Run("get blocks in range", func(t *testing.T) {
		blocks, err := storage.GetBlocks(ctx, "ethereum", 10, 15)
		if err != nil {
			t.Fatalf("GetBlocks() error = %v", err)
		}

		if len(blocks) != 6 {
			t.Errorf("len(blocks) = %v, want 6", len(blocks))
		}

		// Verify blocks are in order
		for i, block := range blocks {
			expectedNum := uint64(10 + i)
			if block.Number != expectedNum {
				t.Errorf("blocks[%d].Number = %v, want %v", i, block.Number, expectedNum)
			}
		}
	})

	t.Run("get blocks with start > end", func(t *testing.T) {
		_, err := storage.GetBlocks(ctx, "ethereum", 20, 10)
		if err == nil {
			t.Error("GetBlocks() with start > end should return error")
		}
	})

	t.Run("get blocks with no results", func(t *testing.T) {
		blocks, err := storage.GetBlocks(ctx, "nonexistent-chain", 1000, 2000)
		if err != nil {
			t.Fatalf("GetBlocks() error = %v", err)
		}

		if len(blocks) != 0 {
			t.Errorf("len(blocks) = %v, want 0", len(blocks))
		}
	})
}

func TestBlockRepo_GetLatestBlock(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	t.Run("get latest block when no blocks exist", func(t *testing.T) {
		_, err := storage.GetLatestBlock(ctx, "ethereum")
		if err != repository.ErrBlockNotFound {
			t.Errorf("GetLatestBlock() error = %v, want ErrBlockNotFound", err)
		}
	})

	t.Run("get latest block", func(t *testing.T) {
		// Save blocks
		for i := 1; i <= 10; i++ {
			block := models.NewBlock(models.ChainTypeEVM, "ethereum", uint64(i), "0xhash")
			if err := storage.SaveBlock(ctx, block); err != nil {
				t.Fatalf("SaveBlock() error = %v", err)
			}
		}

		latest, err := storage.GetLatestBlock(ctx, "ethereum")
		if err != nil {
			t.Fatalf("GetLatestBlock() error = %v", err)
		}

		if latest.Number != 10 {
			t.Errorf("Number = %v, want 10", latest.Number)
		}
	})
}

func TestBlockRepo_HasBlock(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save a block
	block := models.NewBlock(models.ChainTypeEVM, "ethereum", 100, "0xhash")
	if err := storage.SaveBlock(ctx, block); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("has existing block", func(t *testing.T) {
		exists, err := storage.HasBlock(ctx, "ethereum", 100)
		if err != nil {
			t.Fatalf("HasBlock() error = %v", err)
		}

		if !exists {
			t.Error("HasBlock() = false, want true")
		}
	})

	t.Run("has non-existent block", func(t *testing.T) {
		exists, err := storage.HasBlock(ctx, "ethereum", 999)
		if err != nil {
			t.Fatalf("HasBlock() error = %v", err)
		}

		if exists {
			t.Error("HasBlock() = true, want false")
		}
	})
}

func TestBlockRepo_SaveBlocks(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	t.Run("save multiple blocks", func(t *testing.T) {
		blocks := make([]*models.Block, 10)
		for i := 0; i < 10; i++ {
			blocks[i] = models.NewBlock(models.ChainTypeEVM, "ethereum", uint64(i), "0xhash")
		}

		err := storage.SaveBlocks(ctx, blocks)
		if err != nil {
			t.Fatalf("SaveBlocks() error = %v", err)
		}

		// Verify all blocks were saved
		for i := 0; i < 10; i++ {
			exists, err := storage.HasBlock(ctx, "ethereum", uint64(i))
			if err != nil {
				t.Fatalf("HasBlock() error = %v", err)
			}
			if !exists {
				t.Errorf("block %d not found", i)
			}
		}

		// Verify latest height
		height, err := storage.GetLatestHeight(ctx, "ethereum")
		if err != nil {
			t.Fatalf("GetLatestHeight() error = %v", err)
		}

		if height != 9 {
			t.Errorf("LatestHeight = %v, want 9", height)
		}
	})

	t.Run("save empty blocks slice", func(t *testing.T) {
		err := storage.SaveBlocks(ctx, []*models.Block{})
		if err != nil {
			t.Errorf("SaveBlocks() with empty slice should not error, got: %v", err)
		}
	})
}

func TestBlockRepo_DeleteBlock(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save a block
	block := models.NewBlock(models.ChainTypeEVM, "ethereum", 100, "0xhash123")
	if err := storage.SaveBlock(ctx, block); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("delete existing block", func(t *testing.T) {
		err := storage.DeleteBlock(ctx, "ethereum", 100)
		if err != nil {
			t.Fatalf("DeleteBlock() error = %v", err)
		}

		// Verify block was deleted
		exists, err := storage.HasBlock(ctx, "ethereum", 100)
		if err != nil {
			t.Fatalf("HasBlock() error = %v", err)
		}

		if exists {
			t.Error("block should be deleted")
		}

		// Verify hash index was also deleted
		_, err = storage.GetBlockByHash(ctx, "ethereum", "0xhash123")
		if err != repository.ErrBlockNotFound {
			t.Errorf("GetBlockByHash() error = %v, want ErrBlockNotFound", err)
		}
	})

	t.Run("delete non-existent block", func(t *testing.T) {
		err := storage.DeleteBlock(ctx, "ethereum", 999)
		if err != repository.ErrBlockNotFound {
			t.Errorf("DeleteBlock() error = %v, want ErrBlockNotFound", err)
		}
	})
}

func TestBlockRepo_QueryBlocks(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save blocks
	for i := 1; i <= 50; i++ {
		block := models.NewBlock(models.ChainTypeEVM, "ethereum", uint64(i), "0xhash")
		if err := storage.SaveBlock(ctx, block); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	t.Run("query with pagination", func(t *testing.T) {
		chainID := "ethereum"
		min := uint64(10)
		max := uint64(30)

		filter := &models.BlockFilter{
			ChainID:      &chainID,
			NumberMin:    &min,
			NumberMax:    &max,
		}

		pagination := &models.PaginationOptions{
			Limit:  5,
			Offset: 0,
		}

		blocks, err := storage.QueryBlocks(ctx, filter, pagination)
		if err != nil {
			t.Fatalf("QueryBlocks() error = %v", err)
		}

		if len(blocks) != 5 {
			t.Errorf("len(blocks) = %v, want 5", len(blocks))
		}

		// Verify first block
		if blocks[0].Number != 10 {
			t.Errorf("blocks[0].Number = %v, want 10", blocks[0].Number)
		}
	})

	t.Run("query with nil filter", func(t *testing.T) {
		_, err := storage.QueryBlocks(ctx, nil, nil)
		if err == nil {
			t.Error("QueryBlocks() with nil filter should return error")
		}
	})
}

// Benchmark tests
func BenchmarkBlockRepo_SaveBlock(b *testing.B) {
	storage, tmpDir := setupTestDB(&testing.T{})
	defer cleanupTestDB(&testing.T{}, storage, tmpDir)

	ctx := context.Background()
	block := models.NewBlock(models.ChainTypeEVM, "ethereum", 1, "0xabc")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block.Number = uint64(i)
		_ = storage.SaveBlock(ctx, block)
	}
}

func BenchmarkBlockRepo_GetBlock(b *testing.B) {
	storage, tmpDir := setupTestDB(&testing.T{})
	defer cleanupTestDB(&testing.T{}, storage, tmpDir)

	ctx := context.Background()

	// Setup
	for i := 0; i < 1000; i++ {
		block := models.NewBlock(models.ChainTypeEVM, "ethereum", uint64(i), "0xabc")
		_ = storage.SaveBlock(ctx, block)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetBlock(ctx, "ethereum", uint64(i%1000))
	}
}

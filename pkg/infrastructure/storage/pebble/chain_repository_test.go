package pebble

import (
	"context"
	"testing"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
)

func TestChainRepo_SaveChain(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	t.Run("save valid chain", func(t *testing.T) {
		chain := models.NewChain(models.ChainTypeEVM, "ethereum", "Ethereum Mainnet")
		chain.Network = "mainnet"
		chain.RPCEndpoints = []string{"http://localhost:8545"}

		err := storage.SaveChain(ctx, chain)
		if err != nil {
			t.Fatalf("SaveChain() error = %v", err)
		}

		// Verify chain was saved
		retrieved, err := storage.GetChain(ctx, "ethereum")
		if err != nil {
			t.Fatalf("GetChain() error = %v", err)
		}

		if retrieved.ChainID != chain.ChainID {
			t.Errorf("ChainID = %v, want %v", retrieved.ChainID, chain.ChainID)
		}

		if retrieved.Name != chain.Name {
			t.Errorf("Name = %v, want %v", retrieved.Name, chain.Name)
		}
	})

	t.Run("save nil chain", func(t *testing.T) {
		err := storage.SaveChain(ctx, nil)
		if err == nil {
			t.Error("SaveChain(nil) should return error")
		}
	})

	t.Run("save invalid chain", func(t *testing.T) {
		chain := &models.Chain{
			ChainType: "invalid",
			ChainID:   "test",
			Name:      "Test",
		}

		err := storage.SaveChain(ctx, chain)
		if err == nil {
			t.Error("SaveChain() with invalid chain should return error")
		}
	})
}

func TestChainRepo_GetChain(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save a chain
	chain := models.NewChain(models.ChainTypeEVM, "ethereum", "Ethereum")
	chain.RPCEndpoints = []string{"http://localhost:8545"}
	if err := storage.SaveChain(ctx, chain); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("get existing chain", func(t *testing.T) {
		retrieved, err := storage.GetChain(ctx, "ethereum")
		if err != nil {
			t.Fatalf("GetChain() error = %v", err)
		}

		if retrieved.ChainID != "ethereum" {
			t.Errorf("ChainID = %v, want ethereum", retrieved.ChainID)
		}
	})

	t.Run("get non-existent chain", func(t *testing.T) {
		_, err := storage.GetChain(ctx, "nonexistent")
		if err != repository.ErrChainNotFound {
			t.Errorf("GetChain() error = %v, want ErrChainNotFound", err)
		}
	})
}

func TestChainRepo_GetAllChains(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	t.Run("get all chains when empty", func(t *testing.T) {
		chains, err := storage.GetAllChains(ctx)
		if err != nil {
			t.Fatalf("GetAllChains() error = %v", err)
		}

		if len(chains) != 0 {
			t.Errorf("len(chains) = %v, want 0", len(chains))
		}
	})

	t.Run("get all chains", func(t *testing.T) {
		// Setup: save multiple chains
		chainIDs := []string{"ethereum", "bsc", "polygon"}
		for _, id := range chainIDs {
			chain := models.NewChain(models.ChainTypeEVM, id, id)
			chain.RPCEndpoints = []string{"http://localhost:8545"}
			if err := storage.SaveChain(ctx, chain); err != nil {
				t.Fatalf("SaveChain() error = %v", err)
			}
		}

		chains, err := storage.GetAllChains(ctx)
		if err != nil {
			t.Fatalf("GetAllChains() error = %v", err)
		}

		if len(chains) != 3 {
			t.Errorf("len(chains) = %v, want 3", len(chains))
		}
	})
}

func TestChainRepo_GetChainsByType(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save chains of different types
	evmChain := models.NewChain(models.ChainTypeEVM, "ethereum", "Ethereum")
	evmChain.RPCEndpoints = []string{"http://localhost:8545"}
	if err := storage.SaveChain(ctx, evmChain); err != nil {
		t.Fatalf("SaveChain() error = %v", err)
	}

	solanaChain := models.NewChain(models.ChainTypeSolana, "solana", "Solana")
	solanaChain.RPCEndpoints = []string{"http://localhost:8899"}
	if err := storage.SaveChain(ctx, solanaChain); err != nil {
		t.Fatalf("SaveChain() error = %v", err)
	}

	t.Run("get EVM chains", func(t *testing.T) {
		chains, err := storage.GetChainsByType(ctx, models.ChainTypeEVM)
		if err != nil {
			t.Fatalf("GetChainsByType() error = %v", err)
		}

		if len(chains) != 1 {
			t.Errorf("len(chains) = %v, want 1", len(chains))
		}

		if chains[0].ChainType != models.ChainTypeEVM {
			t.Errorf("ChainType = %v, want %v", chains[0].ChainType, models.ChainTypeEVM)
		}
	})

	t.Run("get Solana chains", func(t *testing.T) {
		chains, err := storage.GetChainsByType(ctx, models.ChainTypeSolana)
		if err != nil {
			t.Fatalf("GetChainsByType() error = %v", err)
		}

		if len(chains) != 1 {
			t.Errorf("len(chains) = %v, want 1", len(chains))
		}
	})
}

func TestChainRepo_GetEnabledChains(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save enabled and disabled chains
	enabledChain := models.NewChain(models.ChainTypeEVM, "ethereum", "Ethereum")
	enabledChain.RPCEndpoints = []string{"http://localhost:8545"}
	enabledChain.Enabled = true
	if err := storage.SaveChain(ctx, enabledChain); err != nil {
		t.Fatalf("SaveChain() error = %v", err)
	}

	disabledChain := models.NewChain(models.ChainTypeEVM, "bsc", "BSC")
	disabledChain.RPCEndpoints = []string{"http://localhost:8545"}
	disabledChain.Enabled = false
	if err := storage.SaveChain(ctx, disabledChain); err != nil {
		t.Fatalf("SaveChain() error = %v", err)
	}

	t.Run("get enabled chains", func(t *testing.T) {
		chains, err := storage.GetEnabledChains(ctx)
		if err != nil {
			t.Fatalf("GetEnabledChains() error = %v", err)
		}

		if len(chains) != 1 {
			t.Errorf("len(chains) = %v, want 1", len(chains))
		}

		if chains[0].ChainID != "ethereum" {
			t.Errorf("ChainID = %v, want ethereum", chains[0].ChainID)
		}
	})
}

func TestChainRepo_HasChain(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save a chain
	chain := models.NewChain(models.ChainTypeEVM, "ethereum", "Ethereum")
	chain.RPCEndpoints = []string{"http://localhost:8545"}
	if err := storage.SaveChain(ctx, chain); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("has existing chain", func(t *testing.T) {
		exists, err := storage.HasChain(ctx, "ethereum")
		if err != nil {
			t.Fatalf("HasChain() error = %v", err)
		}

		if !exists {
			t.Error("HasChain() = false, want true")
		}
	})

	t.Run("has non-existent chain", func(t *testing.T) {
		exists, err := storage.HasChain(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("HasChain() error = %v", err)
		}

		if exists {
			t.Error("HasChain() = true, want false")
		}
	})
}

func TestChainRepo_UpdateChain(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save a chain
	chain := models.NewChain(models.ChainTypeEVM, "ethereum", "Ethereum")
	chain.RPCEndpoints = []string{"http://localhost:8545"}
	if err := storage.SaveChain(ctx, chain); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("update existing chain", func(t *testing.T) {
		// Modify chain
		chain.Name = "Ethereum Mainnet Updated"
		chain.BatchSize = 200

		err := storage.UpdateChain(ctx, chain)
		if err != nil {
			t.Fatalf("UpdateChain() error = %v", err)
		}

		// Verify update
		retrieved, err := storage.GetChain(ctx, "ethereum")
		if err != nil {
			t.Fatalf("GetChain() error = %v", err)
		}

		if retrieved.Name != "Ethereum Mainnet Updated" {
			t.Errorf("Name = %v, want Ethereum Mainnet Updated", retrieved.Name)
		}

		if retrieved.BatchSize != 200 {
			t.Errorf("BatchSize = %v, want 200", retrieved.BatchSize)
		}
	})
}

func TestChainRepo_DeleteChain(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save a chain
	chain := models.NewChain(models.ChainTypeEVM, "ethereum", "Ethereum")
	chain.RPCEndpoints = []string{"http://localhost:8545"}
	if err := storage.SaveChain(ctx, chain); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("delete existing chain", func(t *testing.T) {
		err := storage.DeleteChain(ctx, "ethereum")
		if err != nil {
			t.Fatalf("DeleteChain() error = %v", err)
		}

		// Verify chain was deleted
		exists, err := storage.HasChain(ctx, "ethereum")
		if err != nil {
			t.Fatalf("HasChain() error = %v", err)
		}

		if exists {
			t.Error("chain should be deleted")
		}
	})

	t.Run("delete non-existent chain", func(t *testing.T) {
		err := storage.DeleteChain(ctx, "nonexistent")
		if err != nil {
			t.Errorf("DeleteChain() on non-existent chain should not error, got: %v", err)
		}
	})
}

func TestChainRepo_UpdateChainStatus(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save a chain
	chain := models.NewChain(models.ChainTypeEVM, "ethereum", "Ethereum")
	chain.RPCEndpoints = []string{"http://localhost:8545"}
	chain.Status = models.ChainStatusIdle
	if err := storage.SaveChain(ctx, chain); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("update chain status", func(t *testing.T) {
		err := storage.UpdateChainStatus(ctx, "ethereum", models.ChainStatusSyncing)
		if err != nil {
			t.Fatalf("UpdateChainStatus() error = %v", err)
		}

		// Verify status was updated
		retrieved, err := storage.GetChain(ctx, "ethereum")
		if err != nil {
			t.Fatalf("GetChain() error = %v", err)
		}

		if retrieved.Status != models.ChainStatusSyncing {
			t.Errorf("Status = %v, want %v", retrieved.Status, models.ChainStatusSyncing)
		}
	})
}

func TestChainRepo_UpdateLatestBlock(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save a chain
	chain := models.NewChain(models.ChainTypeEVM, "ethereum", "Ethereum")
	chain.RPCEndpoints = []string{"http://localhost:8545"}
	if err := storage.SaveChain(ctx, chain); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("update latest block", func(t *testing.T) {
		err := storage.UpdateLatestBlock(ctx, "ethereum", 1000, 1100)
		if err != nil {
			t.Fatalf("UpdateLatestBlock() error = %v", err)
		}

		// Verify update
		retrieved, err := storage.GetChain(ctx, "ethereum")
		if err != nil {
			t.Fatalf("GetChain() error = %v", err)
		}

		if retrieved.LatestIndexedBlock != 1000 {
			t.Errorf("LatestIndexedBlock = %v, want 1000", retrieved.LatestIndexedBlock)
		}

		if retrieved.LatestChainBlock != 1100 {
			t.Errorf("LatestChainBlock = %v, want 1100", retrieved.LatestChainBlock)
		}
	})
}

func TestChainRepo_GetChainStats(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	t.Run("get stats for non-existent chain", func(t *testing.T) {
		stats, err := storage.GetChainStats(ctx, "ethereum")
		if err != nil {
			t.Fatalf("GetChainStats() error = %v", err)
		}

		// Should return empty stats, not error
		if stats.ChainID != "ethereum" {
			t.Errorf("ChainID = %v, want ethereum", stats.ChainID)
		}

		if stats.TotalBlocks != 0 {
			t.Errorf("TotalBlocks = %v, want 0", stats.TotalBlocks)
		}
	})

	t.Run("update and get chain stats", func(t *testing.T) {
		stats := &models.ChainStats{
			ChainID:            "ethereum",
			ChainType:          models.ChainTypeEVM,
			LatestIndexedBlock: 1000,
			LatestChainBlock:   1100,
			TotalBlocks:        1000,
			TotalTransactions:  5000,
		}

		err := storage.UpdateChainStats(ctx, stats)
		if err != nil {
			t.Fatalf("UpdateChainStats() error = %v", err)
		}

		// Verify stats
		retrieved, err := storage.GetChainStats(ctx, "ethereum")
		if err != nil {
			t.Fatalf("GetChainStats() error = %v", err)
		}

		if retrieved.TotalBlocks != 1000 {
			t.Errorf("TotalBlocks = %v, want 1000", retrieved.TotalBlocks)
		}

		if retrieved.TotalTransactions != 5000 {
			t.Errorf("TotalTransactions = %v, want 5000", retrieved.TotalTransactions)
		}
	})
}

func TestChainRepo_GetAllChainStats(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save multiple chains
	for i, id := range []string{"ethereum", "bsc", "polygon"} {
		chain := models.NewChain(models.ChainTypeEVM, id, id)
		chain.RPCEndpoints = []string{"http://localhost:8545"}
		if err := storage.SaveChain(ctx, chain); err != nil {
			t.Fatalf("SaveChain() error = %v", err)
		}

		// Add stats
		stats := &models.ChainStats{
			ChainID:       id,
			ChainType:     models.ChainTypeEVM,
			TotalBlocks:   uint64((i + 1) * 1000),
			TotalTransactions: uint64((i + 1) * 5000),
		}
		if err := storage.UpdateChainStats(ctx, stats); err != nil {
			t.Fatalf("UpdateChainStats() error = %v", err)
		}
	}

	t.Run("get all chain stats", func(t *testing.T) {
		allStats, err := storage.GetAllChainStats(ctx)
		if err != nil {
			t.Fatalf("GetAllChainStats() error = %v", err)
		}

		if len(allStats) != 3 {
			t.Errorf("len(allStats) = %v, want 3", len(allStats))
		}
	})
}

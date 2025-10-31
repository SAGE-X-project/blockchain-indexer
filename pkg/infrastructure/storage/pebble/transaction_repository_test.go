package pebble

import (
	"context"
	"testing"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
)

func TestTransactionRepo_SaveTransaction(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	t.Run("save valid transaction", func(t *testing.T) {
		tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx123")
		tx.BlockNumber = 100
		tx.BlockHash = "0xblock"
		tx.From = "0xfrom"
		tx.To = "0xto"
		tx.Value = "1000000"
		tx.Status = models.TxStatusSuccess

		err := storage.SaveTransaction(ctx, tx)
		if err != nil {
			t.Fatalf("SaveTransaction() error = %v", err)
		}

		// Verify transaction was saved
		retrieved, err := storage.GetTransaction(ctx, "ethereum", "0xtx123")
		if err != nil {
			t.Fatalf("GetTransaction() error = %v", err)
		}

		if retrieved.Hash != tx.Hash {
			t.Errorf("Hash = %v, want %v", retrieved.Hash, tx.Hash)
		}

		if retrieved.From != tx.From {
			t.Errorf("From = %v, want %v", retrieved.From, tx.From)
		}

		if retrieved.To != tx.To {
			t.Errorf("To = %v, want %v", retrieved.To, tx.To)
		}
	})

	t.Run("save nil transaction", func(t *testing.T) {
		err := storage.SaveTransaction(ctx, nil)
		if err == nil {
			t.Error("SaveTransaction(nil) should return error")
		}
	})

	t.Run("save transaction with empty from address", func(t *testing.T) {
		tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx123")
		tx.From = "" // Invalid

		err := storage.SaveTransaction(ctx, tx)
		if err == nil {
			t.Error("SaveTransaction() with empty from should return error")
		}
	})
}

func TestTransactionRepo_GetTransaction(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save a transaction
	tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx123")
	tx.BlockNumber = 100
	tx.From = "0xfrom"
	tx.To = "0xto"
	if err := storage.SaveTransaction(ctx, tx); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("get existing transaction", func(t *testing.T) {
		retrieved, err := storage.GetTransaction(ctx, "ethereum", "0xtx123")
		if err != nil {
			t.Fatalf("GetTransaction() error = %v", err)
		}

		if retrieved.Hash != "0xtx123" {
			t.Errorf("Hash = %v, want 0xtx123", retrieved.Hash)
		}
	})

	t.Run("get non-existent transaction", func(t *testing.T) {
		_, err := storage.GetTransaction(ctx, "ethereum", "0xnonexistent")
		if err != repository.ErrTransactionNotFound {
			t.Errorf("GetTransaction() error = %v, want ErrTransactionNotFound", err)
		}
	})
}

func TestTransactionRepo_GetTransactionsByBlock(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save transactions for block 100
	for i := 0; i < 5; i++ {
		tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx"+string(rune('0'+i)))
		tx.BlockNumber = 100
		tx.Index = uint64(i)
		tx.From = "0xfrom"
		tx.To = "0xto"

		if err := storage.SaveTransaction(ctx, tx); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	t.Run("get transactions by block", func(t *testing.T) {
		txs, err := storage.GetTransactionsByBlock(ctx, "ethereum", 100)
		if err != nil {
			t.Fatalf("GetTransactionsByBlock() error = %v", err)
		}

		if len(txs) != 5 {
			t.Errorf("len(txs) = %v, want 5", len(txs))
		}
	})

	t.Run("get transactions for empty block", func(t *testing.T) {
		txs, err := storage.GetTransactionsByBlock(ctx, "ethereum", 999)
		if err != nil {
			t.Fatalf("GetTransactionsByBlock() error = %v", err)
		}

		if len(txs) != 0 {
			t.Errorf("len(txs) = %v, want 0", len(txs))
		}
	})
}

func TestTransactionRepo_GetTransactionsByAddress(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	address := "0xaddress123"

	// Setup: save transactions involving the address
	for i := 0; i < 3; i++ {
		tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx"+string(rune('A'+i)))
		tx.BlockNumber = uint64(100 + i)
		tx.Index = 0
		tx.From = address // From address
		tx.To = "0xother"

		if err := storage.SaveTransaction(ctx, tx); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	for i := 0; i < 2; i++ {
		tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx"+string(rune('D'+i)))
		tx.BlockNumber = uint64(200 + i)
		tx.Index = 0
		tx.From = "0xother"
		tx.To = address // To address

		if err := storage.SaveTransaction(ctx, tx); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	t.Run("get transactions by address", func(t *testing.T) {
		pagination := &models.PaginationOptions{
			Limit:  10,
			Offset: 0,
		}

		txs, err := storage.GetTransactionsByAddress(ctx, "ethereum", address, pagination)
		if err != nil {
			t.Fatalf("GetTransactionsByAddress() error = %v", err)
		}

		if len(txs) != 5 {
			t.Errorf("len(txs) = %v, want 5", len(txs))
		}
	})

	t.Run("get transactions with pagination", func(t *testing.T) {
		pagination := &models.PaginationOptions{
			Limit:  2,
			Offset: 0,
		}

		txs, err := storage.GetTransactionsByAddress(ctx, "ethereum", address, pagination)
		if err != nil {
			t.Fatalf("GetTransactionsByAddress() error = %v", err)
		}

		if len(txs) != 2 {
			t.Errorf("len(txs) = %v, want 2", len(txs))
		}
	})
}

func TestTransactionRepo_HasTransaction(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save a transaction
	tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx123")
	tx.From = "0xfrom"
	if err := storage.SaveTransaction(ctx, tx); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("has existing transaction", func(t *testing.T) {
		exists, err := storage.HasTransaction(ctx, "ethereum", "0xtx123")
		if err != nil {
			t.Fatalf("HasTransaction() error = %v", err)
		}

		if !exists {
			t.Error("HasTransaction() = false, want true")
		}
	})

	t.Run("has non-existent transaction", func(t *testing.T) {
		exists, err := storage.HasTransaction(ctx, "ethereum", "0xnonexistent")
		if err != nil {
			t.Fatalf("HasTransaction() error = %v", err)
		}

		if exists {
			t.Error("HasTransaction() = true, want false")
		}
	})
}

func TestTransactionRepo_SaveTransactions(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	t.Run("save multiple transactions", func(t *testing.T) {
		txs := make([]*models.Transaction, 10)
		for i := 0; i < 10; i++ {
			tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx"+string(rune('0'+i)))
			tx.BlockNumber = 100
			tx.Index = uint64(i)
			tx.From = "0xfrom"
			tx.To = "0xto"
			txs[i] = tx
		}

		err := storage.SaveTransactions(ctx, txs)
		if err != nil {
			t.Fatalf("SaveTransactions() error = %v", err)
		}

		// Verify all transactions were saved
		for i := 0; i < 10; i++ {
			exists, err := storage.HasTransaction(ctx, "ethereum", "0xtx"+string(rune('0'+i)))
			if err != nil {
				t.Fatalf("HasTransaction() error = %v", err)
			}
			if !exists {
				t.Errorf("transaction %d not found", i)
			}
		}
	})

	t.Run("save empty transactions slice", func(t *testing.T) {
		err := storage.SaveTransactions(ctx, []*models.Transaction{})
		if err != nil {
			t.Errorf("SaveTransactions() with empty slice should not error, got: %v", err)
		}
	})
}

func TestTransactionRepo_DeleteTransaction(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Setup: save a transaction
	tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx123")
	tx.BlockNumber = 100
	tx.Index = 0
	tx.From = "0xfrom"
	tx.To = "0xto"
	if err := storage.SaveTransaction(ctx, tx); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("delete existing transaction", func(t *testing.T) {
		err := storage.DeleteTransaction(ctx, "ethereum", "0xtx123")
		if err != nil {
			t.Fatalf("DeleteTransaction() error = %v", err)
		}

		// Verify transaction was deleted
		exists, err := storage.HasTransaction(ctx, "ethereum", "0xtx123")
		if err != nil {
			t.Fatalf("HasTransaction() error = %v", err)
		}

		if exists {
			t.Error("transaction should be deleted")
		}
	})

	t.Run("delete non-existent transaction", func(t *testing.T) {
		err := storage.DeleteTransaction(ctx, "ethereum", "0xnonexistent")
		if err != repository.ErrTransactionNotFound {
			t.Errorf("DeleteTransaction() error = %v, want ErrTransactionNotFound", err)
		}
	})
}

func TestTransactionRepo_AddressIndex(t *testing.T) {
	storage, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, storage, tmpDir)

	ctx := context.Background()

	// Save a transaction
	tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx123")
	tx.BlockNumber = 100
	tx.Index = 0
	tx.From = "0xfrom"
	tx.To = "0xto"
	if err := storage.SaveTransaction(ctx, tx); err != nil {
		t.Fatalf("SaveTransaction() error = %v", err)
	}

	t.Run("get address transactions", func(t *testing.T) {
		pagination := &models.PaginationOptions{
			Limit:  10,
			Offset: 0,
		}

		hashes, err := storage.GetAddressTransactions(ctx, "ethereum", "0xfrom", pagination)
		if err != nil {
			t.Fatalf("GetAddressTransactions() error = %v", err)
		}

		if len(hashes) == 0 {
			t.Error("GetAddressTransactions() returned empty result")
		}

		if hashes[0] != "0xtx123" {
			t.Errorf("hash = %v, want 0xtx123", hashes[0])
		}
	})
}

// Benchmark tests
func BenchmarkTransactionRepo_SaveTransaction(b *testing.B) {
	storage, tmpDir := setupTestDB(&testing.T{})
	defer cleanupTestDB(&testing.T{}, storage, tmpDir)

	ctx := context.Background()
	tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx")
	tx.From = "0xfrom"
	tx.To = "0xto"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx.Hash = "0xtx" + string(rune(i))
		tx.Index = uint64(i)
		_ = storage.SaveTransaction(ctx, tx)
	}
}

func BenchmarkTransactionRepo_GetTransaction(b *testing.B) {
	storage, tmpDir := setupTestDB(&testing.T{})
	defer cleanupTestDB(&testing.T{}, storage, tmpDir)

	ctx := context.Background()

	// Setup
	for i := 0; i < 1000; i++ {
		tx := models.NewTransaction(models.ChainTypeEVM, "ethereum", "0xtx"+string(rune(i)))
		tx.From = "0xfrom"
		tx.Index = uint64(i)
		_ = storage.SaveTransaction(ctx, tx)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetTransaction(ctx, "ethereum", "0xtx"+string(rune(i%1000)))
	}
}

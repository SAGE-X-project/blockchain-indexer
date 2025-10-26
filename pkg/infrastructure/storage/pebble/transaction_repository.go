package pebble

import (
	"context"
	"fmt"

	"github.com/cockroachdb/pebble"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
)

// TransactionRepo implements the TransactionRepository interface using PebbleDB
type TransactionRepo struct {
	db      *pebble.DB
	encoder *Encoder
}

// NewTransactionRepo creates a new transaction repository
func NewTransactionRepo(db *pebble.DB, encoder *Encoder) *TransactionRepo {
	return &TransactionRepo{
		db:      db,
		encoder: encoder,
	}
}

// GetTransaction retrieves a transaction by chain ID and hash
func (r *TransactionRepo) GetTransaction(ctx context.Context, chainID string, hash string) (*models.Transaction, error) {
	key := TransactionKey(chainID, hash)

	value, closer, err := r.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, repository.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	defer closer.Close()

	// Decode the transaction
	tx, err := r.encoder.DecodeTransaction(value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %w", err)
	}

	return tx, nil
}

// GetTransactionsByBlock retrieves all transactions in a block
func (r *TransactionRepo) GetTransactionsByBlock(ctx context.Context, chainID string, blockNumber uint64) ([]*models.Transaction, error) {
	// Create iterator for transaction-by-block prefix
	prefix := TransactionByBlockPrefix(chainID, blockNumber)

	iter, err := r.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: keyUpperBound(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Close()

	transactions := make([]*models.Transaction, 0)

	// Iterate through transaction hash references
	for iter.First(); iter.Valid(); iter.Next() {
		// The value is the transaction hash
		txHash := r.encoder.DecodeString(iter.Value())

		// Get the actual transaction
		tx, err := r.GetTransaction(ctx, chainID, txHash)
		if err != nil {
			return nil, fmt.Errorf("failed to get transaction %s: %w", txHash, err)
		}

		transactions = append(transactions, tx)
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	return transactions, nil
}

// GetTransactionsByAddress retrieves transactions for an address
func (r *TransactionRepo) GetTransactionsByAddress(ctx context.Context, chainID string, address string, pagination *models.PaginationOptions) ([]*models.Transaction, error) {
	// Get transaction hashes first
	txHashes, err := r.GetAddressTransactions(ctx, chainID, address, pagination)
	if err != nil {
		return nil, err
	}

	// Fetch full transactions
	transactions := make([]*models.Transaction, 0, len(txHashes))
	for _, txHash := range txHashes {
		tx, err := r.GetTransaction(ctx, chainID, txHash)
		if err != nil {
			if err == repository.ErrTransactionNotFound {
				// Skip missing transactions
				continue
			}
			return nil, fmt.Errorf("failed to get transaction %s: %w", txHash, err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// HasTransaction checks if a transaction exists
func (r *TransactionRepo) HasTransaction(ctx context.Context, chainID string, hash string) (bool, error) {
	key := TransactionKey(chainID, hash)

	_, closer, err := r.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check transaction existence: %w", err)
	}
	defer closer.Close()

	return true, nil
}

// QueryTransactions queries transactions with filtering and pagination
func (r *TransactionRepo) QueryTransactions(ctx context.Context, filter *models.TransactionFilter, pagination *models.PaginationOptions) ([]*models.Transaction, error) {
	if filter == nil {
		return nil, fmt.Errorf("filter cannot be nil")
	}

	// If filtering by address, use address index
	if filter.From != nil {
		return r.GetTransactionsByAddress(ctx, *filter.ChainID, *filter.From, pagination)
	}

	if filter.To != nil {
		return r.GetTransactionsByAddress(ctx, *filter.ChainID, *filter.To, pagination)
	}

	// If filtering by block range, iterate through blocks
	if filter.BlockNumberMin != nil && filter.BlockNumberMax != nil && filter.ChainID != nil {
		transactions := make([]*models.Transaction, 0)

		for blockNum := *filter.BlockNumberMin; blockNum <= *filter.BlockNumberMax; blockNum++ {
			txs, err := r.GetTransactionsByBlock(ctx, *filter.ChainID, blockNum)
			if err != nil {
				return nil, fmt.Errorf("failed to get transactions for block %d: %w", blockNum, err)
			}
			transactions = append(transactions, txs...)
		}

		// Apply pagination
		if pagination != nil {
			if err := pagination.Validate(); err != nil {
				return nil, err
			}

			offset := pagination.Offset
			limit := pagination.Limit

			if offset >= len(transactions) {
				return []*models.Transaction{}, nil
			}

			end := offset + limit
			if end > len(transactions) {
				end = len(transactions)
			}

			transactions = transactions[offset:end]
		}

		return transactions, nil
	}

	return nil, fmt.Errorf("unsupported query filter combination")
}

// QueryTransactionSummaries queries transaction summaries with filtering and pagination
func (r *TransactionRepo) QueryTransactionSummaries(ctx context.Context, filter *models.TransactionFilter, pagination *models.PaginationOptions) ([]*models.TransactionSummary, error) {
	transactions, err := r.QueryTransactions(ctx, filter, pagination)
	if err != nil {
		return nil, err
	}

	summaries := make([]*models.TransactionSummary, len(transactions))
	for i, tx := range transactions {
		summaries[i] = tx.ToSummary()
	}

	return summaries, nil
}

// CountTransactions counts transactions matching the filter
func (r *TransactionRepo) CountTransactions(ctx context.Context, filter *models.TransactionFilter) (uint64, error) {
	transactions, err := r.QueryTransactions(ctx, filter, nil)
	if err != nil {
		return 0, err
	}

	return uint64(len(transactions)), nil
}

// SaveTransaction saves a single transaction
func (r *TransactionRepo) SaveTransaction(ctx context.Context, tx *models.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	// Encode the transaction
	data, err := r.encoder.EncodeTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to encode transaction: %w", err)
	}

	// Save the transaction by hash
	txKey := TransactionKey(tx.ChainID, tx.Hash)
	if err := r.db.Set(txKey, data, pebble.Sync); err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	// Save transaction-by-block index
	txBlockKey := TransactionByBlockKey(tx.ChainID, tx.BlockNumber, tx.Index)
	txHashData := r.encoder.EncodeString(tx.Hash)
	if err := r.db.Set(txBlockKey, txHashData, pebble.Sync); err != nil {
		return fmt.Errorf("failed to save transaction-by-block index: %w", err)
	}

	// Save address indexes
	if err := r.AddAddressIndex(ctx, tx.ChainID, tx.From, tx.Hash); err != nil {
		return fmt.Errorf("failed to save from address index: %w", err)
	}

	if tx.To != "" {
		if err := r.AddAddressIndex(ctx, tx.ChainID, tx.To, tx.Hash); err != nil {
			return fmt.Errorf("failed to save to address index: %w", err)
		}
	}

	return nil
}

// SaveTransactions saves multiple transactions
func (r *TransactionRepo) SaveTransactions(ctx context.Context, txs []*models.Transaction) error {
	if len(txs) == 0 {
		return nil
	}

	// Use a batch for efficiency
	batch := r.db.NewBatch()
	defer batch.Close()

	for _, tx := range txs {
		if tx == nil {
			continue
		}

		if err := tx.Validate(); err != nil {
			return fmt.Errorf("invalid transaction %s: %w", tx.Hash, err)
		}

		// Encode the transaction
		data, err := r.encoder.EncodeTransaction(tx)
		if err != nil {
			return fmt.Errorf("failed to encode transaction %s: %w", tx.Hash, err)
		}

		// Save the transaction by hash
		txKey := TransactionKey(tx.ChainID, tx.Hash)
		if err := batch.Set(txKey, data, pebble.Sync); err != nil {
			return fmt.Errorf("failed to batch set transaction %s: %w", tx.Hash, err)
		}

		// Save transaction-by-block index
		txBlockKey := TransactionByBlockKey(tx.ChainID, tx.BlockNumber, tx.Index)
		txHashData := r.encoder.EncodeString(tx.Hash)
		if err := batch.Set(txBlockKey, txHashData, pebble.Sync); err != nil {
			return fmt.Errorf("failed to batch set transaction-by-block index: %w", err)
		}

		// Save address indexes
		fromAddrKey := AddressTxKey(tx.ChainID, tx.From, tx.BlockNumber, tx.Index)
		if err := batch.Set(fromAddrKey, txHashData, pebble.Sync); err != nil {
			return fmt.Errorf("failed to batch set from address index: %w", err)
		}

		if tx.To != "" {
			toAddrKey := AddressTxKey(tx.ChainID, tx.To, tx.BlockNumber, tx.Index)
			if err := batch.Set(toAddrKey, txHashData, pebble.Sync); err != nil {
				return fmt.Errorf("failed to batch set to address index: %w", err)
			}
		}
	}

	// Commit the batch
	if err := batch.Commit(pebble.Sync); err != nil {
		return fmt.Errorf("failed to commit batch: %w", err)
	}

	return nil
}

// UpdateTransaction updates an existing transaction
func (r *TransactionRepo) UpdateTransaction(ctx context.Context, tx *models.Transaction) error {
	// For PebbleDB, update is the same as save (overwrite)
	return r.SaveTransaction(ctx, tx)
}

// DeleteTransaction deletes a transaction
func (r *TransactionRepo) DeleteTransaction(ctx context.Context, chainID string, hash string) error {
	// Get the transaction first to get its metadata
	tx, err := r.GetTransaction(ctx, chainID, hash)
	if err != nil {
		return err
	}

	// Delete the transaction
	txKey := TransactionKey(chainID, hash)
	if err := r.db.Delete(txKey, pebble.Sync); err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	// Delete transaction-by-block index
	txBlockKey := TransactionByBlockKey(chainID, tx.BlockNumber, tx.Index)
	if err := r.db.Delete(txBlockKey, pebble.Sync); err != nil {
		return fmt.Errorf("failed to delete transaction-by-block index: %w", err)
	}

	// Delete address indexes
	fromAddrKey := AddressTxKey(chainID, tx.From, tx.BlockNumber, tx.Index)
	if err := r.db.Delete(fromAddrKey, pebble.Sync); err != nil {
		return fmt.Errorf("failed to delete from address index: %w", err)
	}

	if tx.To != "" {
		toAddrKey := AddressTxKey(chainID, tx.To, tx.BlockNumber, tx.Index)
		if err := r.db.Delete(toAddrKey, pebble.Sync); err != nil {
			return fmt.Errorf("failed to delete to address index: %w", err)
		}
	}

	return nil
}

// AddAddressIndex adds an address index for a transaction
func (r *TransactionRepo) AddAddressIndex(ctx context.Context, chainID string, address string, txHash string) error {
	// Get the transaction to get block number and index
	tx, err := r.GetTransaction(ctx, chainID, txHash)
	if err != nil {
		return err
	}

	// Create address index key
	key := AddressTxKey(chainID, address, tx.BlockNumber, tx.Index)
	data := r.encoder.EncodeString(txHash)

	if err := r.db.Set(key, data, pebble.Sync); err != nil {
		return fmt.Errorf("failed to save address index: %w", err)
	}

	return nil
}

// GetAddressTransactions retrieves transaction hashes for an address
func (r *TransactionRepo) GetAddressTransactions(ctx context.Context, chainID string, address string, pagination *models.PaginationOptions) ([]string, error) {
	// Create iterator for address prefix
	prefix := AddressTxPrefix(chainID, address)

	iter, err := r.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: keyUpperBound(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Close()

	txHashes := make([]string, 0)

	// Iterate through transaction hash references
	for iter.First(); iter.Valid(); iter.Next() {
		txHash := r.encoder.DecodeString(iter.Value())
		txHashes = append(txHashes, txHash)
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	// Apply pagination
	if pagination != nil {
		if err := pagination.Validate(); err != nil {
			return nil, err
		}

		offset := pagination.Offset
		limit := pagination.Limit

		if offset >= len(txHashes) {
			return []string{}, nil
		}

		end := offset + limit
		if end > len(txHashes) {
			end = len(txHashes)
		}

		txHashes = txHashes[offset:end]
	}

	return txHashes, nil
}

// SaveTransactionsBatch saves transactions in batches for better performance
func (r *TransactionRepo) SaveTransactionsBatch(ctx context.Context, txs []*models.Transaction, batchSize int) error {
	if len(txs) == 0 {
		return nil
	}

	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	// Process transactions in batches
	for i := 0; i < len(txs); i += batchSize {
		end := i + batchSize
		if end > len(txs) {
			end = len(txs)
		}

		batch := txs[i:end]
		if err := r.SaveTransactions(ctx, batch); err != nil {
			return fmt.Errorf("failed to save batch starting at index %d: %w", i, err)
		}
	}

	return nil
}

// keyUpperBound returns the upper bound for a prefix scan
func keyUpperBound(prefix []byte) []byte {
	end := make([]byte, len(prefix))
	copy(end, prefix)
	for i := len(end) - 1; i >= 0; i-- {
		end[i]++
		if end[i] != 0 {
			return end
		}
	}
	return nil // No upper bound if we reach here
}

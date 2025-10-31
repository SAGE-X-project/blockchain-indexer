package pebble

import (
	"context"
	"fmt"

	"github.com/cockroachdb/pebble"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// PebbleBatch implements the repository.Batch interface using PebbleDB
type PebbleBatch struct {
	db      *pebble.DB
	batch   *pebble.Batch
	encoder *Encoder
	count   int
}

// NewBatch creates a new batch instance
func NewBatch(db *pebble.DB, encoder *Encoder) *PebbleBatch {
	return &PebbleBatch{
		db:      db,
		batch:   db.NewBatch(),
		encoder: encoder,
		count:   0,
	}
}

// SetBlock adds a block to the batch
func (b *PebbleBatch) SetBlock(ctx context.Context, block *models.Block) error {
	if block == nil {
		return fmt.Errorf("block cannot be nil")
	}

	if err := block.Validate(); err != nil {
		return fmt.Errorf("invalid block: %w", err)
	}

	// Encode the block
	data, err := b.encoder.EncodeBlock(block)
	if err != nil {
		return fmt.Errorf("failed to encode block: %w", err)
	}

	// Add block to batch
	blockKey := BlockKey(block.ChainID, block.Number)
	if err := b.batch.Set(blockKey, data, pebble.Sync); err != nil {
		return fmt.Errorf("failed to batch set block: %w", err)
	}
	b.count++

	// Add block hash index to batch
	hashKey := BlockHashKey(block.ChainID, block.Hash)
	numberData := b.encoder.EncodeUint64(block.Number)
	if err := b.batch.Set(hashKey, numberData, pebble.Sync); err != nil {
		return fmt.Errorf("failed to batch set block hash index: %w", err)
	}
	b.count++

	return nil
}

// SetBlocks adds multiple blocks to the batch
func (b *PebbleBatch) SetBlocks(ctx context.Context, blocks []*models.Block) error {
	for _, block := range blocks {
		if err := b.SetBlock(ctx, block); err != nil {
			return fmt.Errorf("failed to batch set block %d: %w", block.Number, err)
		}
	}

	// Update latest heights for all chains
	latestHeights := make(map[string]uint64)
	for _, block := range blocks {
		if currentLatest, exists := latestHeights[block.ChainID]; !exists || block.Number > currentLatest {
			latestHeights[block.ChainID] = block.Number
		}
	}

	// Add latest height updates to batch
	for chainID, height := range latestHeights {
		heightKey := LatestHeightKey(chainID)
		heightData := b.encoder.EncodeUint64(height)
		if err := b.batch.Set(heightKey, heightData, pebble.Sync); err != nil {
			return fmt.Errorf("failed to batch set latest height: %w", err)
		}
		b.count++
	}

	return nil
}

// SetTransaction adds a transaction to the batch
func (b *PebbleBatch) SetTransaction(ctx context.Context, tx *models.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	// Encode the transaction
	data, err := b.encoder.EncodeTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to encode transaction: %w", err)
	}

	// Add transaction to batch
	txKey := TransactionKey(tx.ChainID, tx.Hash)
	if err := b.batch.Set(txKey, data, pebble.Sync); err != nil {
		return fmt.Errorf("failed to batch set transaction: %w", err)
	}
	b.count++

	// Add transaction-by-block index to batch
	txBlockKey := TransactionByBlockKey(tx.ChainID, tx.BlockNumber, tx.Index)
	txHashData := b.encoder.EncodeString(tx.Hash)
	if err := b.batch.Set(txBlockKey, txHashData, pebble.Sync); err != nil {
		return fmt.Errorf("failed to batch set transaction-by-block index: %w", err)
	}
	b.count++

	// Add from address index to batch
	fromAddrKey := AddressTxKey(tx.ChainID, tx.From, tx.BlockNumber, tx.Index)
	if err := b.batch.Set(fromAddrKey, txHashData, pebble.Sync); err != nil {
		return fmt.Errorf("failed to batch set from address index: %w", err)
	}
	b.count++

	// Add to address index to batch (if not contract creation)
	if tx.To != "" {
		toAddrKey := AddressTxKey(tx.ChainID, tx.To, tx.BlockNumber, tx.Index)
		if err := b.batch.Set(toAddrKey, txHashData, pebble.Sync); err != nil {
			return fmt.Errorf("failed to batch set to address index: %w", err)
		}
		b.count++
	}

	return nil
}

// SetTransactions adds multiple transactions to the batch
func (b *PebbleBatch) SetTransactions(ctx context.Context, txs []*models.Transaction) error {
	for _, tx := range txs {
		if err := b.SetTransaction(ctx, tx); err != nil {
			return fmt.Errorf("failed to batch set transaction %s: %w", tx.Hash, err)
		}
	}
	return nil
}

// Commit writes all batched operations atomically
func (b *PebbleBatch) Commit() error {
	if b.batch == nil {
		return fmt.Errorf("batch is nil")
	}

	if err := b.batch.Commit(pebble.Sync); err != nil {
		return fmt.Errorf("failed to commit batch: %w", err)
	}

	// Reset the batch after successful commit
	b.batch = b.db.NewBatch()
	b.count = 0

	return nil
}

// Reset clears all operations in the batch without committing
func (b *PebbleBatch) Reset() {
	if b.batch != nil {
		b.batch.Reset()
	}
	b.count = 0
}

// Count returns the number of operations in the batch
func (b *PebbleBatch) Count() int {
	return b.count
}

// Close releases batch resources without committing
func (b *PebbleBatch) Close() error {
	if b.batch != nil {
		if err := b.batch.Close(); err != nil {
			return fmt.Errorf("failed to close batch: %w", err)
		}
		b.batch = nil
	}
	b.count = 0
	return nil
}

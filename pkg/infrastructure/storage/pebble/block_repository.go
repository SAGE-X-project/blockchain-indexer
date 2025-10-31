package pebble

import (
	"context"
	"fmt"

	"github.com/cockroachdb/pebble"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
)

// BlockRepo implements the BlockRepository interface using PebbleDB
type BlockRepo struct {
	db      *pebble.DB
	encoder *Encoder
}

// NewBlockRepo creates a new block repository
func NewBlockRepo(db *pebble.DB, encoder *Encoder) *BlockRepo {
	return &BlockRepo{
		db:      db,
		encoder: encoder,
	}
}

// GetBlock retrieves a block by chain ID and block number
func (r *BlockRepo) GetBlock(ctx context.Context, chainID string, number uint64) (*models.Block, error) {
	key := BlockKey(chainID, number)

	value, closer, err := r.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, repository.ErrBlockNotFound
		}
		return nil, fmt.Errorf("failed to get block: %w", err)
	}
	defer closer.Close()

	// Decode the block
	block, err := r.encoder.DecodeBlock(value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode block: %w", err)
	}

	return block, nil
}

// GetBlockByHash retrieves a block by chain ID and block hash
func (r *BlockRepo) GetBlockByHash(ctx context.Context, chainID string, hash string) (*models.Block, error) {
	// First get the block number from the hash index
	hashKey := BlockHashKey(chainID, hash)

	value, closer, err := r.db.Get(hashKey)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, repository.ErrBlockNotFound
		}
		return nil, fmt.Errorf("failed to get block hash index: %w", err)
	}
	defer closer.Close()

	// Decode the block number
	blockNumber, err := r.encoder.DecodeUint64(value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode block number: %w", err)
	}

	// Get the actual block
	return r.GetBlock(ctx, chainID, blockNumber)
}

// GetBlocks retrieves blocks in a range
func (r *BlockRepo) GetBlocks(ctx context.Context, chainID string, start, end uint64) ([]*models.Block, error) {
	if start > end {
		return nil, fmt.Errorf("start block number must be less than or equal to end")
	}

	blocks := make([]*models.Block, 0, end-start+1)

	// Create iterator for the range
	startKey := BlockKey(chainID, start)
	endKey := BlockKey(chainID, end+1) // Exclusive upper bound

	iter, err := r.db.NewIter(&pebble.IterOptions{
		LowerBound: startKey,
		UpperBound: endKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Close()

	// Iterate through blocks
	for iter.First(); iter.Valid(); iter.Next() {
		block, err := r.encoder.DecodeBlock(iter.Value())
		if err != nil {
			return nil, fmt.Errorf("failed to decode block: %w", err)
		}
		blocks = append(blocks, block)
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	return blocks, nil
}

// GetLatestBlock retrieves the latest block for a chain
func (r *BlockRepo) GetLatestBlock(ctx context.Context, chainID string) (*models.Block, error) {
	height, err := r.GetLatestHeight(ctx, chainID)
	if err != nil {
		return nil, err
	}

	return r.GetBlock(ctx, chainID, height)
}

// GetLatestHeight retrieves the latest block height for a chain
func (r *BlockRepo) GetLatestHeight(ctx context.Context, chainID string) (uint64, error) {
	key := LatestHeightKey(chainID)

	value, closer, err := r.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return 0, repository.ErrBlockNotFound
		}
		return 0, fmt.Errorf("failed to get latest height: %w", err)
	}
	defer closer.Close()

	height, err := r.encoder.DecodeUint64(value)
	if err != nil {
		return 0, fmt.Errorf("failed to decode height: %w", err)
	}

	return height, nil
}

// HasBlock checks if a block exists
func (r *BlockRepo) HasBlock(ctx context.Context, chainID string, number uint64) (bool, error) {
	key := BlockKey(chainID, number)

	_, closer, err := r.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check block existence: %w", err)
	}
	defer closer.Close()

	return true, nil
}

// QueryBlocks queries blocks with filtering and pagination
func (r *BlockRepo) QueryBlocks(ctx context.Context, filter *models.BlockFilter, pagination *models.PaginationOptions) ([]*models.Block, error) {
	// For now, implement basic filtering by chain and block range
	// More advanced filtering can be added later
	if filter == nil {
		return nil, fmt.Errorf("filter cannot be nil")
	}

	if filter.ChainID == nil {
		return nil, fmt.Errorf("chain ID is required for query")
	}

	// Determine range
	start := uint64(0)
	if filter.NumberMin != nil {
		start = *filter.NumberMin
	}

	end := uint64(1<<63 - 1) // Max value
	if filter.NumberMax != nil {
		end = *filter.NumberMax
	}

	// Get blocks in range
	blocks, err := r.GetBlocks(ctx, *filter.ChainID, start, end)
	if err != nil {
		return nil, err
	}

	// Apply pagination
	if pagination != nil {
		if err := pagination.Validate(); err != nil {
			return nil, err
		}

		offset := pagination.Offset
		limit := pagination.Limit

		if offset >= len(blocks) {
			return []*models.Block{}, nil
		}

		end := offset + limit
		if end > len(blocks) {
			end = len(blocks)
		}

		blocks = blocks[offset:end]
	}

	return blocks, nil
}

// QueryBlockSummaries queries block summaries with filtering and pagination
func (r *BlockRepo) QueryBlockSummaries(ctx context.Context, filter *models.BlockFilter, pagination *models.PaginationOptions) ([]*models.BlockSummary, error) {
	blocks, err := r.QueryBlocks(ctx, filter, pagination)
	if err != nil {
		return nil, err
	}

	summaries := make([]*models.BlockSummary, len(blocks))
	for i, block := range blocks {
		summaries[i] = block.ToSummary()
	}

	return summaries, nil
}

// CountBlocks counts blocks matching the filter
func (r *BlockRepo) CountBlocks(ctx context.Context, filter *models.BlockFilter) (uint64, error) {
	blocks, err := r.QueryBlocks(ctx, filter, nil)
	if err != nil {
		return 0, err
	}

	return uint64(len(blocks)), nil
}

// SaveBlock saves a single block
func (r *BlockRepo) SaveBlock(ctx context.Context, block *models.Block) error {
	if block == nil {
		return fmt.Errorf("block cannot be nil")
	}

	if err := block.Validate(); err != nil {
		return fmt.Errorf("invalid block: %w", err)
	}

	// Encode the block
	data, err := r.encoder.EncodeBlock(block)
	if err != nil {
		return fmt.Errorf("failed to encode block: %w", err)
	}

	// Save the block by number
	blockKey := BlockKey(block.ChainID, block.Number)
	if err := r.db.Set(blockKey, data, pebble.Sync); err != nil {
		return fmt.Errorf("failed to save block: %w", err)
	}

	// Save the hash index
	hashKey := BlockHashKey(block.ChainID, block.Hash)
	numberData := r.encoder.EncodeUint64(block.Number)
	if err := r.db.Set(hashKey, numberData, pebble.Sync); err != nil {
		return fmt.Errorf("failed to save block hash index: %w", err)
	}

	// Update latest height if this is the latest block
	currentHeight, err := r.GetLatestHeight(ctx, block.ChainID)
	if err != nil && err != repository.ErrBlockNotFound {
		return fmt.Errorf("failed to get current height: %w", err)
	}

	if err == repository.ErrBlockNotFound || block.Number > currentHeight {
		heightKey := LatestHeightKey(block.ChainID)
		heightData := r.encoder.EncodeUint64(block.Number)
		if err := r.db.Set(heightKey, heightData, pebble.Sync); err != nil {
			return fmt.Errorf("failed to update latest height: %w", err)
		}
	}

	return nil
}

// SaveBlocks saves multiple blocks
func (r *BlockRepo) SaveBlocks(ctx context.Context, blocks []*models.Block) error {
	if len(blocks) == 0 {
		return nil
	}

	// Use a batch for efficiency
	batch := r.db.NewBatch()
	defer batch.Close()

	latestHeights := make(map[string]uint64) // Track latest height per chain

	for _, block := range blocks {
		if block == nil {
			continue
		}

		if err := block.Validate(); err != nil {
			return fmt.Errorf("invalid block %d: %w", block.Number, err)
		}

		// Encode the block
		data, err := r.encoder.EncodeBlock(block)
		if err != nil {
			return fmt.Errorf("failed to encode block %d: %w", block.Number, err)
		}

		// Save the block by number
		blockKey := BlockKey(block.ChainID, block.Number)
		if err := batch.Set(blockKey, data, pebble.Sync); err != nil {
			return fmt.Errorf("failed to batch set block %d: %w", block.Number, err)
		}

		// Save the hash index
		hashKey := BlockHashKey(block.ChainID, block.Hash)
		numberData := r.encoder.EncodeUint64(block.Number)
		if err := batch.Set(hashKey, numberData, pebble.Sync); err != nil {
			return fmt.Errorf("failed to batch set block hash index: %w", err)
		}

		// Track latest height
		if currentLatest, exists := latestHeights[block.ChainID]; !exists || block.Number > currentLatest {
			latestHeights[block.ChainID] = block.Number
		}
	}

	// Update latest heights for all chains
	for chainID, height := range latestHeights {
		heightKey := LatestHeightKey(chainID)
		heightData := r.encoder.EncodeUint64(height)
		if err := batch.Set(heightKey, heightData, pebble.Sync); err != nil {
			return fmt.Errorf("failed to batch set latest height: %w", err)
		}
	}

	// Commit the batch
	if err := batch.Commit(pebble.Sync); err != nil {
		return fmt.Errorf("failed to commit batch: %w", err)
	}

	return nil
}

// UpdateBlock updates an existing block
func (r *BlockRepo) UpdateBlock(ctx context.Context, block *models.Block) error {
	// For PebbleDB, update is the same as save (overwrite)
	return r.SaveBlock(ctx, block)
}

// DeleteBlock deletes a block
func (r *BlockRepo) DeleteBlock(ctx context.Context, chainID string, number uint64) error {
	// Get the block first to get its hash
	block, err := r.GetBlock(ctx, chainID, number)
	if err != nil {
		return err
	}

	// Delete the block
	blockKey := BlockKey(chainID, number)
	if err := r.db.Delete(blockKey, pebble.Sync); err != nil {
		return fmt.Errorf("failed to delete block: %w", err)
	}

	// Delete the hash index
	hashKey := BlockHashKey(chainID, block.Hash)
	if err := r.db.Delete(hashKey, pebble.Sync); err != nil {
		return fmt.Errorf("failed to delete block hash index: %w", err)
	}

	return nil
}

// SaveBlocksBatch saves blocks in batches for better performance
func (r *BlockRepo) SaveBlocksBatch(ctx context.Context, blocks []*models.Block, batchSize int) error {
	if len(blocks) == 0 {
		return nil
	}

	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	// Process blocks in batches
	for i := 0; i < len(blocks); i += batchSize {
		end := i + batchSize
		if end > len(blocks) {
			end = len(blocks)
		}

		batch := blocks[i:end]
		if err := r.SaveBlocks(ctx, batch); err != nil {
			return fmt.Errorf("failed to save batch starting at index %d: %w", i, err)
		}
	}

	return nil
}

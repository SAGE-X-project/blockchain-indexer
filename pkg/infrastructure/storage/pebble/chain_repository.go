package pebble

import (
	"context"
	"fmt"

	"github.com/cockroachdb/pebble"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
)

// ChainRepo implements the ChainRepository interface using PebbleDB
type ChainRepo struct {
	db      *pebble.DB
	encoder *Encoder
}

// NewChainRepo creates a new chain repository
func NewChainRepo(db *pebble.DB, encoder *Encoder) *ChainRepo {
	return &ChainRepo{
		db:      db,
		encoder: encoder,
	}
}

// GetChain retrieves a chain by ID
func (r *ChainRepo) GetChain(ctx context.Context, chainID string) (*models.Chain, error) {
	key := ChainKey(chainID)

	value, closer, err := r.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, repository.ErrChainNotFound
		}
		return nil, fmt.Errorf("failed to get chain: %w", err)
	}
	defer closer.Close()

	// Decode the chain
	chain, err := r.encoder.DecodeChain(value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode chain: %w", err)
	}

	return chain, nil
}

// GetAllChains retrieves all chains
func (r *ChainRepo) GetAllChains(ctx context.Context) ([]*models.Chain, error) {
	prefix := []byte(PrefixChain)

	iter, err := r.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: keyUpperBound(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Close()

	chains := make([]*models.Chain, 0)

	for iter.First(); iter.Valid(); iter.Next() {
		chain, err := r.encoder.DecodeChain(iter.Value())
		if err != nil {
			return nil, fmt.Errorf("failed to decode chain: %w", err)
		}
		chains = append(chains, chain)
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	return chains, nil
}

// GetChainsByType retrieves all chains of a specific type
func (r *ChainRepo) GetChainsByType(ctx context.Context, chainType models.ChainType) ([]*models.Chain, error) {
	chains, err := r.GetAllChains(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]*models.Chain, 0)
	for _, chain := range chains {
		if chain.ChainType == chainType {
			filtered = append(filtered, chain)
		}
	}

	return filtered, nil
}

// GetEnabledChains retrieves all enabled chains
func (r *ChainRepo) GetEnabledChains(ctx context.Context) ([]*models.Chain, error) {
	chains, err := r.GetAllChains(ctx)
	if err != nil {
		return nil, err
	}

	enabled := make([]*models.Chain, 0)
	for _, chain := range chains {
		if chain.Enabled {
			enabled = append(enabled, chain)
		}
	}

	return enabled, nil
}

// HasChain checks if a chain exists
func (r *ChainRepo) HasChain(ctx context.Context, chainID string) (bool, error) {
	key := ChainKey(chainID)

	_, closer, err := r.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check chain existence: %w", err)
	}
	defer closer.Close()

	return true, nil
}

// SaveChain saves a chain configuration
func (r *ChainRepo) SaveChain(ctx context.Context, chain *models.Chain) error {
	if chain == nil {
		return fmt.Errorf("chain cannot be nil")
	}

	if err := chain.Validate(); err != nil {
		return fmt.Errorf("invalid chain: %w", err)
	}

	// Encode the chain
	data, err := r.encoder.EncodeChain(chain)
	if err != nil {
		return fmt.Errorf("failed to encode chain: %w", err)
	}

	// Save the chain
	key := ChainKey(chain.ChainID)
	if err := r.db.Set(key, data, pebble.Sync); err != nil {
		return fmt.Errorf("failed to save chain: %w", err)
	}

	return nil
}

// UpdateChain updates an existing chain
func (r *ChainRepo) UpdateChain(ctx context.Context, chain *models.Chain) error {
	// For PebbleDB, update is the same as save (overwrite)
	return r.SaveChain(ctx, chain)
}

// DeleteChain deletes a chain
func (r *ChainRepo) DeleteChain(ctx context.Context, chainID string) error {
	key := ChainKey(chainID)

	if err := r.db.Delete(key, pebble.Sync); err != nil {
		return fmt.Errorf("failed to delete chain: %w", err)
	}

	// Also delete chain stats
	statsKey := ChainStatsKey(chainID)
	if err := r.db.Delete(statsKey, pebble.Sync); err != nil {
		// Don't fail if stats don't exist
		if err != pebble.ErrNotFound {
			return fmt.Errorf("failed to delete chain stats: %w", err)
		}
	}

	return nil
}

// UpdateChainStatus updates the status of a chain
func (r *ChainRepo) UpdateChainStatus(ctx context.Context, chainID string, status models.ChainStatus) error {
	chain, err := r.GetChain(ctx, chainID)
	if err != nil {
		return err
	}

	chain.Status = status
	return r.UpdateChain(ctx, chain)
}

// UpdateLatestBlock updates the latest indexed block for a chain
func (r *ChainRepo) UpdateLatestBlock(ctx context.Context, chainID string, indexedBlock, chainBlock uint64) error {
	chain, err := r.GetChain(ctx, chainID)
	if err != nil {
		return err
	}

	chain.LatestIndexedBlock = indexedBlock
	chain.LatestChainBlock = chainBlock

	return r.UpdateChain(ctx, chain)
}

// GetChainStats retrieves statistics for a chain
func (r *ChainRepo) GetChainStats(ctx context.Context, chainID string) (*models.ChainStats, error) {
	key := ChainStatsKey(chainID)

	value, closer, err := r.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			// Return empty stats if not found
			return &models.ChainStats{
				ChainID:            chainID,
				TotalBlocks:        0,
				TotalTransactions:  0,
				LatestIndexedBlock: 0,
				LatestChainBlock:   0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get chain stats: %w", err)
	}
	defer closer.Close()

	// Decode the stats
	stats, err := r.encoder.DecodeChainStats(value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode chain stats: %w", err)
	}

	return stats, nil
}

// GetAllChainStats retrieves statistics for all chains
func (r *ChainRepo) GetAllChainStats(ctx context.Context) ([]*models.ChainStats, error) {
	chains, err := r.GetAllChains(ctx)
	if err != nil {
		return nil, err
	}

	allStats := make([]*models.ChainStats, 0, len(chains))
	for _, chain := range chains {
		stats, err := r.GetChainStats(ctx, chain.ChainID)
		if err != nil {
			return nil, fmt.Errorf("failed to get stats for chain %s: %w", chain.ChainID, err)
		}
		allStats = append(allStats, stats)
	}

	return allStats, nil
}

// UpdateChainStats updates statistics for a chain
func (r *ChainRepo) UpdateChainStats(ctx context.Context, stats *models.ChainStats) error {
	if stats == nil {
		return fmt.Errorf("stats cannot be nil")
	}

	// Encode the stats
	data, err := r.encoder.EncodeChainStats(stats)
	if err != nil {
		return fmt.Errorf("failed to encode chain stats: %w", err)
	}

	// Save the stats
	key := ChainStatsKey(stats.ChainID)
	if err := r.db.Set(key, data, pebble.Sync); err != nil {
		return fmt.Errorf("failed to save chain stats: %w", err)
	}

	return nil
}

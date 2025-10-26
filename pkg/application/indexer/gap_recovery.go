package indexer

import (
	"context"
	"fmt"
	"sort"

	"github.com/sage-x-project/blockchain-indexer/pkg/application/processor"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/service"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"go.uber.org/zap"
)

// GapRecovery handles detection and recovery of missing blocks
type GapRecovery struct {
	adapter   service.ChainAdapter
	blockRepo repository.BlockRepository
	processor *processor.BlockProcessor
	logger    *logger.Logger
}

// NewGapRecovery creates a new gap recovery instance
func NewGapRecovery(
	adapter service.ChainAdapter,
	blockRepo repository.BlockRepository,
	processor *processor.BlockProcessor,
	logger *logger.Logger,
) *GapRecovery {
	return &GapRecovery{
		adapter:   adapter,
		blockRepo: blockRepo,
		processor: processor,
		logger:    logger,
	}
}

// Gap represents a range of missing blocks
type Gap struct {
	ChainID    string
	StartBlock uint64
	EndBlock   uint64
	Size       uint64
}

// String returns a string representation of the gap
func (g *Gap) String() string {
	return fmt.Sprintf("Gap[%s: %d-%d, size=%d]", g.ChainID, g.StartBlock, g.EndBlock, g.Size)
}

// DetectGaps detects gaps in indexed blocks
func (g *GapRecovery) DetectGaps(ctx context.Context, chainID string) ([]*Gap, error) {
	g.logger.Debug("detecting gaps", zap.String("chain_id", chainID))

	// Get latest indexed block
	latestBlock, err := g.blockRepo.GetLatestBlock(ctx, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	if latestBlock == nil {
		// No blocks indexed yet
		return nil, nil
	}

	// Sample blocks to find gaps
	// In production, you might want to use a more sophisticated approach
	// such as maintaining a bitmap or using database queries
	gaps := make([]*Gap, 0)

	// Check blocks in ranges (sampling approach)
	// This is simplified - in production you'd check more thoroughly
	const sampleSize = 1000
	endBlock := latestBlock.Number

	for start := uint64(0); start < endBlock; start += sampleSize {
		end := start + sampleSize
		if end > endBlock {
			end = endBlock
		}

		// Get blocks in range
		blocks, err := g.blockRepo.GetBlocks(ctx, chainID, start, end)
		if err != nil {
			g.logger.Error("failed to get blocks for gap detection",
				zap.String("chain_id", chainID),
				zap.Uint64("start", start),
				zap.Uint64("end", end),
				zap.Error(err),
			)
			continue
		}

		// Find gaps in this range
		rangeGaps := g.findGapsInRange(chainID, blocks, start, end)
		gaps = append(gaps, rangeGaps...)
	}

	if len(gaps) > 0 {
		g.logger.Info("detected gaps",
			zap.String("chain_id", chainID),
			zap.Int("gap_count", len(gaps)),
		)
	}

	return gaps, nil
}

// findGapsInRange finds gaps in a sorted list of blocks
func (g *GapRecovery) findGapsInRange(chainID string, blocks []*models.Block, start, end uint64) []*Gap {
	if len(blocks) == 0 {
		// Entire range is a gap
		return []*Gap{{
			ChainID:    chainID,
			StartBlock: start,
			EndBlock:   end,
			Size:       end - start + 1,
		}}
	}

	// Sort blocks by number
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Number < blocks[j].Number
	})

	gaps := make([]*Gap, 0)
	expectedBlock := start

	for _, block := range blocks {
		if block.Number > expectedBlock {
			// Found a gap
			gaps = append(gaps, &Gap{
				ChainID:    chainID,
				StartBlock: expectedBlock,
				EndBlock:   block.Number - 1,
				Size:       block.Number - expectedBlock,
			})
		}
		expectedBlock = block.Number + 1
	}

	// Check if there's a gap at the end
	if expectedBlock <= end {
		gaps = append(gaps, &Gap{
			ChainID:    chainID,
			StartBlock: expectedBlock,
			EndBlock:   end,
			Size:       end - expectedBlock + 1,
		})
	}

	return gaps
}

// RecoverGap recovers a gap by fetching and processing missing blocks
func (g *GapRecovery) RecoverGap(ctx context.Context, gap *Gap) error {
	g.logger.Info("recovering gap",
		zap.String("chain_id", gap.ChainID),
		zap.Uint64("start", gap.StartBlock),
		zap.Uint64("end", gap.EndBlock),
		zap.Uint64("size", gap.Size),
	)

	// Fetch missing blocks
	blocks, err := g.adapter.GetBlocks(ctx, gap.StartBlock, gap.EndBlock)
	if err != nil {
		return fmt.Errorf("failed to fetch blocks for gap recovery: %w", err)
	}

	// Process blocks
	if err := g.processor.ProcessBlocks(ctx, blocks); err != nil {
		return fmt.Errorf("failed to process blocks for gap recovery: %w", err)
	}

	g.logger.Info("gap recovered",
		zap.String("chain_id", gap.ChainID),
		zap.Uint64("start", gap.StartBlock),
		zap.Uint64("end", gap.EndBlock),
		zap.Int("blocks_recovered", len(blocks)),
	)

	return nil
}

// RecoverAllGaps detects and recovers all gaps
func (g *GapRecovery) RecoverAllGaps(ctx context.Context, chainID string) error {
	gaps, err := g.DetectGaps(ctx, chainID)
	if err != nil {
		return fmt.Errorf("failed to detect gaps: %w", err)
	}

	if len(gaps) == 0 {
		g.logger.Info("no gaps found", zap.String("chain_id", chainID))
		return nil
	}

	successCount := 0
	errorCount := 0

	for _, gap := range gaps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := g.RecoverGap(ctx, gap); err != nil {
			g.logger.Error("failed to recover gap",
				zap.String("chain_id", chainID),
				zap.Uint64("start", gap.StartBlock),
				zap.Uint64("end", gap.EndBlock),
				zap.Error(err),
			)
			errorCount++
		} else {
			successCount++
		}
	}

	g.logger.Info("gap recovery completed",
		zap.String("chain_id", chainID),
		zap.Int("total_gaps", len(gaps)),
		zap.Int("recovered", successCount),
		zap.Int("failed", errorCount),
	)

	if errorCount > 0 {
		return fmt.Errorf("recovered %d/%d gaps", successCount, len(gaps))
	}

	return nil
}

// VerifyBlockContinuity verifies that blocks are continuous within a range
func (g *GapRecovery) VerifyBlockContinuity(ctx context.Context, chainID string, start, end uint64) (bool, error) {
	blocks, err := g.blockRepo.GetBlocks(ctx, chainID, start, end)
	if err != nil {
		return false, fmt.Errorf("failed to get blocks: %w", err)
	}

	expectedCount := end - start + 1
	if uint64(len(blocks)) != expectedCount {
		return false, nil
	}

	// Sort blocks by number
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Number < blocks[j].Number
	})

	// Verify continuity
	for i, block := range blocks {
		expectedNumber := start + uint64(i)
		if block.Number != expectedNumber {
			return false, nil
		}
	}

	return true, nil
}

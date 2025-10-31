package resolver

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/application/indexer"
	"github.com/sage-x-project/blockchain-indexer/pkg/application/statistics"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
	gql "github.com/sage-x-project/blockchain-indexer/pkg/presentation/graphql"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/event"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"go.uber.org/zap"
)

// Resolver is the root GraphQL resolver
type Resolver struct {
	blockRepo       repository.BlockRepository
	txRepo          repository.TransactionRepository
	chainRepo       repository.ChainRepository
	progressTracker *indexer.ProgressTracker
	statsCollector  *statistics.Collector
	gapRecovery     map[string]*indexer.GapRecovery // chainID -> GapRecovery
	eventBus        event.EventBus
	logger          *logger.Logger
}

// NewResolver creates a new GraphQL resolver
func NewResolver(
	blockRepo repository.BlockRepository,
	txRepo repository.TransactionRepository,
	chainRepo repository.ChainRepository,
	progressTracker *indexer.ProgressTracker,
	statsCollector *statistics.Collector,
	gapRecovery map[string]*indexer.GapRecovery,
	eventBus event.EventBus,
	logger *logger.Logger,
) *Resolver {
	return &Resolver{
		blockRepo:       blockRepo,
		txRepo:          txRepo,
		chainRepo:       chainRepo,
		progressTracker: progressTracker,
		statsCollector:  statsCollector,
		gapRecovery:     gapRecovery,
		eventBus:        eventBus,
		logger:          logger,
	}
}

// Query Resolvers

// Chain resolves a single chain by ID
func (r *Resolver) Chain(ctx context.Context, chainID string) (*gql.Chain, error) {
	chain, err := r.chainRepo.GetChain(ctx, chainID)
	if err != nil {
		r.logger.Error("failed to get chain",
			zap.String("chain_id", chainID),
			zap.Error(err),
		)
		return nil, err
	}

	if chain == nil {
		return nil, nil
	}

	return &gql.Chain{
		ChainID:            chain.ChainID,
		ChainType:          gql.ToGraphQLChainType(chain.ChainType),
		Name:               chain.Name,
		Network:            chain.Network,
		Status:             gql.ChainStatusActive,
		StartBlock:         gql.BigInt(fmt.Sprintf("%d", chain.StartBlock)),
		LatestIndexedBlock: gql.BigInt(fmt.Sprintf("%d", chain.LatestIndexedBlock)),
		LatestChainBlock:   gql.BigInt(fmt.Sprintf("%d", chain.LatestChainBlock)),
		LastUpdated:        gql.Time(chain.LastUpdated),
	}, nil
}

// Chains resolves all chains
func (r *Resolver) Chains(ctx context.Context) ([]*gql.Chain, error) {
	// For now, return empty list as we don't have ListChains method
	// TODO: Implement ListChains in repository
	return []*gql.Chain{}, nil
}

// Block resolves a single block by number
func (r *Resolver) Block(ctx context.Context, chainID string, number gql.BigInt) (*gql.Block, error) {
	blockNum, err := strconv.ParseUint(string(number), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid block number: %w", err)
	}

	block, err := r.blockRepo.GetBlock(ctx, chainID, blockNum)
	if err != nil {
		r.logger.Error("failed to get block",
			zap.String("chain_id", chainID),
			zap.Uint64("block_number", blockNum),
			zap.Error(err),
		)
		return nil, err
	}

	return gql.ToGraphQLBlock(block), nil
}

// BlockByHash resolves a block by hash
func (r *Resolver) BlockByHash(ctx context.Context, chainID string, hash string) (*gql.Block, error) {
	block, err := r.blockRepo.GetBlockByHash(ctx, chainID, hash)
	if err != nil {
		r.logger.Error("failed to get block by hash",
			zap.String("chain_id", chainID),
			zap.String("hash", hash),
			zap.Error(err),
		)
		return nil, err
	}

	return gql.ToGraphQLBlock(block), nil
}

// Blocks resolves a paginated list of blocks
func (r *Resolver) Blocks(ctx context.Context, args BlocksArgs) (*gql.BlockConnection, error) {
	// Default pagination
	first := 10
	if args.First != nil && *args.First > 0 {
		first = *args.First
		if first > 100 {
			first = 100 // Max limit
		}
	}

	// Parse cursor if provided
	var startBlock uint64
	if args.After != nil {
		decoded, err := base64.StdEncoding.DecodeString(*args.After)
		if err == nil {
			startBlock, _ = strconv.ParseUint(string(decoded), 10, 64)
		}
	}

	// Get blocks
	endBlock := startBlock + uint64(first)
	blocks, err := r.blockRepo.GetBlocks(ctx, args.ChainID, startBlock, endBlock)
	if err != nil {
		return nil, err
	}

	// Build connection
	edges := make([]*gql.BlockEdge, 0, len(blocks))
	for _, block := range blocks {
		cursor := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", block.Number)))
		edges = append(edges, &gql.BlockEdge{
			Node:   gql.ToGraphQLBlock(block),
			Cursor: cursor,
		})
	}

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &gql.BlockConnection{
		Edges: edges,
		PageInfo: &gql.PageInfo{
			HasNextPage:     len(edges) == first,
			HasPreviousPage: startBlock > 0,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
			TotalCount:      len(edges),
		},
	}, nil
}

// BlockRange resolves a range of blocks
func (r *Resolver) BlockRange(ctx context.Context, chainID string, startBlock gql.BigInt, endBlock gql.BigInt) ([]*gql.Block, error) {
	start, err := strconv.ParseUint(string(startBlock), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid start block: %w", err)
	}

	end, err := strconv.ParseUint(string(endBlock), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid end block: %w", err)
	}

	blocks, err := r.blockRepo.GetBlocks(ctx, chainID, start, end)
	if err != nil {
		return nil, err
	}

	result := make([]*gql.Block, 0, len(blocks))
	for _, block := range blocks {
		result = append(result, gql.ToGraphQLBlock(block))
	}

	return result, nil
}

// LatestBlock resolves the latest indexed block
func (r *Resolver) LatestBlock(ctx context.Context, chainID string) (*gql.Block, error) {
	block, err := r.blockRepo.GetLatestBlock(ctx, chainID)
	if err != nil {
		return nil, err
	}

	return gql.ToGraphQLBlock(block), nil
}

// Transaction resolves a transaction by hash
func (r *Resolver) Transaction(ctx context.Context, chainID string, hash string) (*gql.Transaction, error) {
	tx, err := r.txRepo.GetTransaction(ctx, chainID, hash)
	if err != nil {
		r.logger.Error("failed to get transaction",
			zap.String("chain_id", chainID),
			zap.String("hash", hash),
			zap.Error(err),
		)
		return nil, err
	}

	return gql.ToGraphQLTransaction(tx), nil
}

// Transactions resolves a paginated list of transactions
func (r *Resolver) Transactions(ctx context.Context, args TransactionsArgs) (*gql.TransactionConnection, error) {
	// Default pagination
	first := 10
	if args.First != nil && *args.First > 0 {
		first = *args.First
		if first > 100 {
			first = 100 // Max limit
		}
	}

	// For now, return empty connection
	// TODO: Implement pagination properly
	return &gql.TransactionConnection{
		Edges: []*gql.TransactionEdge{},
		PageInfo: &gql.PageInfo{
			HasNextPage:     false,
			HasPreviousPage: false,
			TotalCount:      0,
		},
	}, nil
}

// TransactionsByBlock resolves transactions by block number
func (r *Resolver) TransactionsByBlock(ctx context.Context, chainID string, blockNumber gql.BigInt) ([]*gql.Transaction, error) {
	blockNum, err := strconv.ParseUint(string(blockNumber), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid block number: %w", err)
	}

	txs, err := r.txRepo.GetTransactionsByBlock(ctx, chainID, blockNum)
	if err != nil {
		return nil, err
	}

	result := make([]*gql.Transaction, 0, len(txs))
	for _, tx := range txs {
		result = append(result, gql.ToGraphQLTransaction(tx))
	}

	return result, nil
}

// TransactionsByAddress resolves transactions by address
func (r *Resolver) TransactionsByAddress(ctx context.Context, args TransactionsByAddressArgs) (*gql.TransactionConnection, error) {
	// Default pagination
	first := 10
	if args.First != nil && *args.First > 0 {
		first = *args.First
		if first > 100 {
			first = 100
		}
	}

	paginationOpts := &models.PaginationOptions{
		Limit:  first,
		Offset: 0,
	}
	txs, err := r.txRepo.GetTransactionsByAddress(ctx, args.ChainID, args.Address, paginationOpts)
	if err != nil {
		return nil, err
	}

	edges := make([]*gql.TransactionEdge, 0, len(txs))
	for _, tx := range txs {
		cursor := base64.StdEncoding.EncodeToString([]byte(tx.Hash))
		edges = append(edges, &gql.TransactionEdge{
			Node:   gql.ToGraphQLTransaction(tx),
			Cursor: cursor,
		})
	}

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &gql.TransactionConnection{
		Edges: edges,
		PageInfo: &gql.PageInfo{
			HasNextPage:     len(edges) == first,
			HasPreviousPage: false,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
			TotalCount:      len(edges),
		},
	}, nil
}

// Progress resolves indexing progress for a chain
func (r *Resolver) Progress(ctx context.Context, chainID string) (*gql.Progress, error) {
	if r.progressTracker == nil {
		return nil, fmt.Errorf("progress tracker not available")
	}

	progress, err := r.progressTracker.GetProgress(ctx, chainID)
	if err != nil {
		return nil, err
	}

	return &gql.Progress{
		ChainID:            progress.ChainID,
		ChainType:          progress.ChainType,
		LatestIndexedBlock: gql.BigInt(fmt.Sprintf("%d", progress.LatestIndexedBlock)),
		LatestChainBlock:   gql.BigInt(fmt.Sprintf("%d", progress.LatestChainBlock)),
		TargetBlock:        gql.BigInt(fmt.Sprintf("%d", progress.TargetBlock)),
		StartBlock:         gql.BigInt(fmt.Sprintf("%d", progress.StartBlock)),
		BlocksBehind:       gql.BigInt(fmt.Sprintf("%d", progress.BlocksBehind)),
		ProgressPercentage: progress.ProgressPercentage,
		BlocksPerSecond:    progress.BlocksPerSecond,
		EstimatedTimeLeft:  progress.EstimatedTimeLeft.String(),
		LastUpdated:        gql.Time(progress.LastUpdated),
		Status:             progress.Status,
	}, nil
}

// AllProgress resolves progress for all chains
func (r *Resolver) AllProgress(ctx context.Context) ([]*gql.Progress, error) {
	// For now, return empty list
	// TODO: Implement when we have chain listing
	return []*gql.Progress{}, nil
}

// Gaps resolves gaps for a chain
func (r *Resolver) Gaps(ctx context.Context, chainID string) ([]*gql.Gap, error) {
	// Check if gap recovery is available for this chain
	if r.gapRecovery == nil {
		r.logger.Warn("gap recovery not initialized")
		return []*gql.Gap{}, nil
	}

	recovery, ok := r.gapRecovery[chainID]
	if !ok {
		r.logger.Warn("gap recovery not found for chain", zap.String("chain_id", chainID))
		return []*gql.Gap{}, nil
	}

	// Detect gaps
	gaps, err := recovery.DetectGaps(ctx, chainID)
	if err != nil {
		r.logger.Error("failed to detect gaps",
			zap.String("chain_id", chainID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to detect gaps: %w", err)
	}

	// Convert to GraphQL type
	result := make([]*gql.Gap, len(gaps))
	for i, gap := range gaps {
		result[i] = &gql.Gap{
			ChainID:    gap.ChainID,
			StartBlock: gql.BigInt(fmt.Sprintf("%d", gap.StartBlock)),
			EndBlock:   gql.BigInt(fmt.Sprintf("%d", gap.EndBlock)),
			Size:       gql.BigInt(fmt.Sprintf("%d", gap.Size)),
		}
	}

	r.logger.Debug("resolved gaps query",
		zap.String("chain_id", chainID),
		zap.Int("gap_count", len(result)),
	)

	return result, nil
}

// Stats resolves statistics for a chain
func (r *Resolver) Stats(ctx context.Context, chainID string) (*gql.Stats, error) {
	if r.statsCollector == nil {
		return &gql.Stats{
			TotalBlocks:       "0",
			TotalTransactions: "0",
			ChainsIndexed:     0,
			AverageBlockTime:  0,
			AverageTxPerBlock: 0,
		}, nil
	}

	stats, err := r.statsCollector.GetChainStatistics(ctx, chainID)
	if err != nil {
		if err == repository.ErrNotFound {
			return &gql.Stats{
				TotalBlocks:       "0",
				TotalTransactions: "0",
				ChainsIndexed:     0,
				AverageBlockTime:  0,
				AverageTxPerBlock: 0,
			}, nil
		}
		r.logger.Error("failed to get chain statistics",
			zap.String("chain_id", chainID),
			zap.Error(err),
		)
		return nil, err
	}

	return &gql.Stats{
		TotalBlocks:       gql.BigInt(fmt.Sprintf("%d", stats.TotalBlocks)),
		TotalTransactions: gql.BigInt(fmt.Sprintf("%d", stats.TotalTransactions)),
		ChainsIndexed:     1,
		AverageBlockTime:  stats.AverageBlockTime,
		AverageTxPerBlock: stats.AverageTxPerBlock,
	}, nil
}

// GlobalStats resolves global statistics
func (r *Resolver) GlobalStats(ctx context.Context) (*gql.Stats, error) {
	if r.statsCollector == nil {
		return &gql.Stats{
			TotalBlocks:       "0",
			TotalTransactions: "0",
			ChainsIndexed:     0,
			AverageBlockTime:  0,
			AverageTxPerBlock: 0,
		}, nil
	}

	stats, err := r.statsCollector.GetGlobalStatistics(ctx)
	if err != nil {
		if err == repository.ErrNotFound {
			return &gql.Stats{
				TotalBlocks:       "0",
				TotalTransactions: "0",
				ChainsIndexed:     0,
				AverageBlockTime:  0,
				AverageTxPerBlock: 0,
			}, nil
		}
		r.logger.Error("failed to get global statistics", zap.Error(err))
		return nil, err
	}

	return &gql.Stats{
		TotalBlocks:       gql.BigInt(fmt.Sprintf("%d", stats.TotalBlocks)),
		TotalTransactions: gql.BigInt(fmt.Sprintf("%d", stats.TotalTransactions)),
		ChainsIndexed:     stats.TotalChains,
		AverageBlockTime:  stats.AverageBlockTime,
		AverageTxPerBlock: stats.AverageTxPerBlock,
	}, nil
}

// Argument types for resolvers

// BlocksArgs represents arguments for the blocks query
type BlocksArgs struct {
	ChainID string
	First   *int
	After   *string
	Last    *int
	Before  *string
	OrderBy *string
}

// TransactionsArgs represents arguments for the transactions query
type TransactionsArgs struct {
	ChainID     string
	First       *int
	After       *string
	Last        *int
	Before      *string
	BlockNumber *gql.BigInt
	From        *string
	To          *string
}

// TransactionsByAddressArgs represents arguments for transactions by address query
type TransactionsByAddressArgs struct {
	ChainID string
	Address string
	First   *int
	After   *string
}

// Subscription Resolvers

// BlockIndexed subscribes to new blocks
func (r *Resolver) BlockIndexed(ctx context.Context, chainID string) (<-chan *gql.Block, error) {
	if r.eventBus == nil {
		return nil, fmt.Errorf("event bus not available")
	}

	ch := make(chan *gql.Block, 10)

	handler := func(evt *event.Event) {
		if payload, ok := evt.Payload.(*event.BlockIndexedPayload); ok {
			select {
			case ch <- gql.ToGraphQLBlock(payload.Block):
			case <-ctx.Done():
				return
			default:
				// Drop if channel is full
			}
		}
	}

	filter := func(evt *event.Event) bool {
		return evt.Type == event.EventTypeBlockIndexed && evt.ChainID == chainID
	}

	subID, err := r.eventBus.Subscribe(filter, handler)
	if err != nil {
		close(ch)
		return nil, err
	}

	// Cleanup on context cancellation
	go func() {
		<-ctx.Done()
		r.eventBus.Unsubscribe(subID)
		close(ch)
	}()

	return ch, nil
}

// TransactionIndexed subscribes to new transactions
func (r *Resolver) TransactionIndexed(ctx context.Context, chainID string) (<-chan *gql.Transaction, error) {
	if r.eventBus == nil {
		return nil, fmt.Errorf("event bus not available")
	}

	ch := make(chan *gql.Transaction, 10)

	handler := func(evt *event.Event) {
		if payload, ok := evt.Payload.(*event.TransactionIndexedPayload); ok {
			select {
			case ch <- gql.ToGraphQLTransaction(payload.Transaction):
			case <-ctx.Done():
				return
			default:
				// Drop if channel is full
			}
		}
	}

	filter := func(evt *event.Event) bool {
		return evt.Type == event.EventTypeTransactionIndexed && evt.ChainID == chainID
	}

	subID, err := r.eventBus.Subscribe(filter, handler)
	if err != nil {
		close(ch)
		return nil, err
	}

	// Cleanup on context cancellation
	go func() {
		<-ctx.Done()
		r.eventBus.Unsubscribe(subID)
		close(ch)
	}()

	return ch, nil
}

// SyncProgress subscribes to sync progress updates
func (r *Resolver) SyncProgress(ctx context.Context, chainID string) (<-chan *gql.Progress, error) {
	// Create a channel that emits progress every 5 seconds
	ch := make(chan *gql.Progress, 1)

	go func() {
		defer close(ch)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if r.progressTracker != nil {
					progress, err := r.progressTracker.GetProgress(ctx, chainID)
					if err == nil {
						gqlProgress := &gql.Progress{
							ChainID:            progress.ChainID,
							ChainType:          progress.ChainType,
							LatestIndexedBlock: gql.BigInt(fmt.Sprintf("%d", progress.LatestIndexedBlock)),
							LatestChainBlock:   gql.BigInt(fmt.Sprintf("%d", progress.LatestChainBlock)),
							TargetBlock:        gql.BigInt(fmt.Sprintf("%d", progress.TargetBlock)),
							StartBlock:         gql.BigInt(fmt.Sprintf("%d", progress.StartBlock)),
							BlocksBehind:       gql.BigInt(fmt.Sprintf("%d", progress.BlocksBehind)),
							ProgressPercentage: progress.ProgressPercentage,
							BlocksPerSecond:    progress.BlocksPerSecond,
							EstimatedTimeLeft:  progress.EstimatedTimeLeft.String(),
							LastUpdated:        gql.Time(progress.LastUpdated),
							Status:             progress.Status,
						}

						select {
						case ch <- gqlProgress:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()

	return ch, nil
}

// GapDetected subscribes to gap detection events
func (r *Resolver) GapDetected(ctx context.Context, chainID string) (<-chan *gql.Gap, error) {
	if r.eventBus == nil {
		return nil, fmt.Errorf("event bus not available")
	}

	ch := make(chan *gql.Gap, 10)

	handler := func(evt *event.Event) {
		if payload, ok := evt.Payload.(*event.GapPayload); ok {
			select {
			case ch <- &gql.Gap{
				ChainID:    payload.ChainID,
				StartBlock: gql.BigInt(fmt.Sprintf("%d", payload.StartBlock)),
				EndBlock:   gql.BigInt(fmt.Sprintf("%d", payload.EndBlock)),
				Size:       gql.BigInt(fmt.Sprintf("%d", payload.Size)),
			}:
			case <-ctx.Done():
				return
			default:
				// Drop if channel is full
			}
		}
	}

	filter := func(evt *event.Event) bool {
		return evt.Type == event.EventTypeGapDetected && evt.ChainID == chainID
	}

	subID, err := r.eventBus.Subscribe(filter, handler)
	if err != nil {
		close(ch)
		return nil, err
	}

	// Cleanup on context cancellation
	go func() {
		<-ctx.Done()
		r.eventBus.Unsubscribe(subID)
		close(ch)
	}()

	return ch, nil
}

// GapRecovered subscribes to gap recovery events
func (r *Resolver) GapRecovered(ctx context.Context, chainID string) (<-chan *gql.Gap, error) {
	if r.eventBus == nil {
		return nil, fmt.Errorf("event bus not available")
	}

	ch := make(chan *gql.Gap, 10)

	handler := func(evt *event.Event) {
		if payload, ok := evt.Payload.(*event.GapPayload); ok {
			select {
			case ch <- &gql.Gap{
				ChainID:    payload.ChainID,
				StartBlock: gql.BigInt(fmt.Sprintf("%d", payload.StartBlock)),
				EndBlock:   gql.BigInt(fmt.Sprintf("%d", payload.EndBlock)),
				Size:       gql.BigInt(fmt.Sprintf("%d", payload.Size)),
			}:
			case <-ctx.Done():
				return
			default:
				// Drop if channel is full
			}
		}
	}

	filter := func(evt *event.Event) bool {
		return evt.Type == event.EventTypeGapRecovered && evt.ChainID == chainID
	}

	subID, err := r.eventBus.Subscribe(filter, handler)
	if err != nil {
		close(ch)
		return nil, err
	}

	// Cleanup on context cancellation
	go func() {
		<-ctx.Done()
		r.eventBus.Unsubscribe(subID)
		close(ch)
	}()

	return ch, nil
}

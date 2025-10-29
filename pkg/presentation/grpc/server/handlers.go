package server

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	indexerv1 "github.com/sage-x-project/blockchain-indexer/api/proto/indexer/v1"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/event"
)

// GetChain retrieves chain information
func (s *Server) GetChain(ctx context.Context, req *indexerv1.GetChainRequest) (*indexerv1.GetChainResponse, error) {
	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain_id is required")
	}

	chain, err := s.chainRepo.GetChain(ctx, req.ChainId)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "chain not found: %s", req.ChainId)
		}
		return nil, status.Errorf(codes.Internal, "failed to get chain: %v", err)
	}

	return &indexerv1.GetChainResponse{
		Chain: convertChainToProto(chain),
	}, nil
}

// ListChains lists all indexed chains
func (s *Server) ListChains(ctx context.Context, req *indexerv1.ListChainsRequest) (*indexerv1.ListChainsResponse, error) {
	chains, err := s.chainRepo.GetAllChains(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list chains: %v", err)
	}

	protoChains := make([]*indexerv1.Chain, len(chains))
	for i, chain := range chains {
		protoChains[i] = convertChainToProto(chain)
	}

	return &indexerv1.ListChainsResponse{
		Chains: protoChains,
	}, nil
}

// GetBlock retrieves a block by number
func (s *Server) GetBlock(ctx context.Context, req *indexerv1.GetBlockRequest) (*indexerv1.GetBlockResponse, error) {
	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain_id is required")
	}

	block, err := s.blockRepo.GetBlock(ctx, req.ChainId, req.Number)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "block not found: %d", req.Number)
		}
		return nil, status.Errorf(codes.Internal, "failed to get block: %v", err)
	}

	return &indexerv1.GetBlockResponse{
		Block: convertBlockToProto(block),
	}, nil
}

// GetBlockByHash retrieves a block by hash
func (s *Server) GetBlockByHash(ctx context.Context, req *indexerv1.GetBlockByHashRequest) (*indexerv1.GetBlockByHashResponse, error) {
	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain_id is required")
	}
	if req.Hash == "" {
		return nil, status.Error(codes.InvalidArgument, "hash is required")
	}

	block, err := s.blockRepo.GetBlockByHash(ctx, req.ChainId, req.Hash)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "block not found: %s", req.Hash)
		}
		return nil, status.Errorf(codes.Internal, "failed to get block: %v", err)
	}

	return &indexerv1.GetBlockByHashResponse{
		Block: convertBlockToProto(block),
	}, nil
}

// ListBlocks lists blocks with pagination
func (s *Server) ListBlocks(ctx context.Context, req *indexerv1.ListBlocksRequest) (*indexerv1.ListBlocksResponse, error) {
	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain_id is required")
	}

	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 100
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	var blocks []*models.Block
	var err error

	if req.StartBlock > 0 && req.EndBlock > 0 {
		blocks, err = s.blockRepo.GetBlocks(ctx, req.ChainId, req.StartBlock, req.EndBlock)
	} else {
		// For now, implement simple pagination without cursor
		// In production, you would implement proper cursor-based pagination
		blocks, err = s.blockRepo.GetBlocks(ctx, req.ChainId, 0, uint64(pageSize))
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list blocks: %v", err)
	}

	protoBlocks := make([]*indexerv1.Block, len(blocks))
	for i, block := range blocks {
		protoBlocks[i] = convertBlockToProto(block)
	}

	return &indexerv1.ListBlocksResponse{
		Blocks:    protoBlocks,
		TotalCount: int32(len(protoBlocks)),
	}, nil
}

// GetLatestBlock retrieves the latest indexed block
func (s *Server) GetLatestBlock(ctx context.Context, req *indexerv1.GetLatestBlockRequest) (*indexerv1.GetLatestBlockResponse, error) {
	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain_id is required")
	}

	latestHeight, err := s.blockRepo.GetLatestHeight(ctx, req.ChainId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get latest height: %v", err)
	}

	block, err := s.blockRepo.GetBlock(ctx, req.ChainId, latestHeight)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Error(codes.NotFound, "no blocks indexed yet")
		}
		return nil, status.Errorf(codes.Internal, "failed to get block: %v", err)
	}

	return &indexerv1.GetLatestBlockResponse{
		Block: convertBlockToProto(block),
	}, nil
}

// GetTransaction retrieves a transaction by hash
func (s *Server) GetTransaction(ctx context.Context, req *indexerv1.GetTransactionRequest) (*indexerv1.GetTransactionResponse, error) {
	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain_id is required")
	}
	if req.Hash == "" {
		return nil, status.Error(codes.InvalidArgument, "hash is required")
	}

	tx, err := s.transactionRepo.GetTransaction(ctx, req.ChainId, req.Hash)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "transaction not found: %s", req.Hash)
		}
		return nil, status.Errorf(codes.Internal, "failed to get transaction: %v", err)
	}

	return &indexerv1.GetTransactionResponse{
		Transaction: convertTransactionToProto(tx),
	}, nil
}

// ListTransactionsByBlock retrieves all transactions in a block
func (s *Server) ListTransactionsByBlock(ctx context.Context, req *indexerv1.ListTransactionsByBlockRequest) (*indexerv1.ListTransactionsByBlockResponse, error) {
	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain_id is required")
	}

	txs, err := s.transactionRepo.GetTransactionsByBlock(ctx, req.ChainId, req.BlockNumber)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transactions: %v", err)
	}

	protoTxs := make([]*indexerv1.Transaction, len(txs))
	for i, tx := range txs {
		protoTxs[i] = convertTransactionToProto(tx)
	}

	return &indexerv1.ListTransactionsByBlockResponse{
		Transactions: protoTxs,
	}, nil
}

// ListTransactionsByAddress retrieves transactions for an address
func (s *Server) ListTransactionsByAddress(ctx context.Context, req *indexerv1.ListTransactionsByAddressRequest) (*indexerv1.ListTransactionsByAddressResponse, error) {
	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain_id is required")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 100
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	pagination := &models.PaginationOptions{
		Limit: int(pageSize),
		Offset: 0,
	}

	txs, err := s.transactionRepo.GetTransactionsByAddress(ctx, req.ChainId, req.Address, pagination)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transactions: %v", err)
	}

	protoTxs := make([]*indexerv1.Transaction, len(txs))
	for i, tx := range txs {
		protoTxs[i] = convertTransactionToProto(tx)
	}

	return &indexerv1.ListTransactionsByAddressResponse{
		Transactions: protoTxs,
		TotalCount:   int32(len(protoTxs)),
	}, nil
}

// GetProgress retrieves indexing progress for a chain
func (s *Server) GetProgress(ctx context.Context, req *indexerv1.GetProgressRequest) (*indexerv1.GetProgressResponse, error) {
	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain_id is required")
	}

	chain, err := s.chainRepo.GetChain(ctx, req.ChainId)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "chain not found: %s", req.ChainId)
		}
		return nil, status.Errorf(codes.Internal, "failed to get chain: %v", err)
	}

	blocksBehind := int64(0)
	if chain.LatestChainBlock > chain.LatestIndexedBlock {
		blocksBehind = int64(chain.LatestChainBlock - chain.LatestIndexedBlock)
	}

	progressPercentage := 0.0
	if chain.LatestChainBlock > 0 {
		progressPercentage = float64(chain.LatestIndexedBlock) / float64(chain.LatestChainBlock) * 100
	}

	progress := &indexerv1.Progress{
		ChainId:             chain.ChainID,
		ChainType:           chain.ChainType.String(),
		LatestIndexedBlock:  chain.LatestIndexedBlock,
		LatestChainBlock:    chain.LatestChainBlock,
		TargetBlock:         chain.LatestChainBlock,
		StartBlock:          chain.StartBlock,
		BlocksBehind:        uint64(blocksBehind),
		ProgressPercentage:  progressPercentage,
		BlocksPerSecond:     0, // Would need to track this separately
		EstimatedTimeLeftSeconds: 0, // Would need to calculate based on rate
		LastUpdated:         timestamppb.New(chain.LastUpdated),
		Status:              chain.Status.String(),
	}

	return &indexerv1.GetProgressResponse{
		Progress: progress,
	}, nil
}

// ListGaps lists gaps in indexed blocks (not implemented yet)
func (s *Server) ListGaps(ctx context.Context, req *indexerv1.ListGapsRequest) (*indexerv1.ListGapsResponse, error) {
	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain_id is required")
	}

	// TODO: Implement gap detection
	return &indexerv1.ListGapsResponse{
		Gaps: []*indexerv1.Gap{},
	}, nil
}

// GetStats retrieves statistics for a chain (not implemented yet)
func (s *Server) GetStats(ctx context.Context, req *indexerv1.GetStatsRequest) (*indexerv1.GetStatsResponse, error) {
	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain_id is required")
	}

	// TODO: Implement statistics calculation
	return &indexerv1.GetStatsResponse{
		Stats: &indexerv1.Stats{
			TotalBlocks:       0,
			TotalTransactions: 0,
			ChainsIndexed:     0,
			AverageBlockTime:  0,
			AverageTxPerBlock: 0,
		},
	}, nil
}

// StreamBlocks streams newly indexed blocks
func (s *Server) StreamBlocks(req *indexerv1.StreamBlocksRequest, stream indexerv1.IndexerService_StreamBlocksServer) error {
	if req.ChainId == "" {
		return status.Error(codes.InvalidArgument, "chain_id is required")
	}

	if s.eventBus == nil {
		return status.Error(codes.Unimplemented, "event bus not configured")
	}

	// Create a channel to receive events
	eventChan := make(chan *models.Block, 100)
	defer close(eventChan)

	// Create a subscriber for block events
	handler := func(e *event.Event) {
		if e.ChainID == req.ChainId && e.Type == event.EventTypeBlockIndexed {
			if payload, ok := e.Payload.(*event.BlockIndexedPayload); ok {
				select {
				case eventChan <- payload.Block:
				default:
					// Drop event if channel is full
				}
			}
		}
	}

	subID, err := s.eventBus.SubscribeChain(req.ChainId, handler)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to subscribe: %v", err)
	}
	defer s.eventBus.Unsubscribe(subID)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case block := <-eventChan:
			if err := stream.Send(convertBlockToProto(block)); err != nil {
				return status.Errorf(codes.Internal, "failed to send block: %v", err)
			}
		}
	}
}

// StreamTransactions streams newly indexed transactions
func (s *Server) StreamTransactions(req *indexerv1.StreamTransactionsRequest, stream indexerv1.IndexerService_StreamTransactionsServer) error {
	if req.ChainId == "" {
		return status.Error(codes.InvalidArgument, "chain_id is required")
	}

	if s.eventBus == nil {
		return status.Error(codes.Unimplemented, "event bus not configured")
	}

	// Create a channel to receive events
	eventChan := make(chan *models.Transaction, 100)
	defer close(eventChan)

	// Create a subscriber for transaction events
	handler := func(e *event.Event) {
		if e.ChainID == req.ChainId && e.Type == event.EventTypeTransactionIndexed {
			if payload, ok := e.Payload.(*event.TransactionIndexedPayload); ok {
				select {
				case eventChan <- payload.Transaction:
				default:
					// Drop event if channel is full
				}
			}
		}
	}

	subID, err := s.eventBus.SubscribeChain(req.ChainId, handler)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to subscribe: %v", err)
	}
	defer s.eventBus.Unsubscribe(subID)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case tx := <-eventChan:
			if err := stream.Send(convertTransactionToProto(tx)); err != nil {
				return status.Errorf(codes.Internal, "failed to send transaction: %v", err)
			}
		}
	}
}

// StreamProgress streams indexing progress updates
func (s *Server) StreamProgress(req *indexerv1.StreamProgressRequest, stream indexerv1.IndexerService_StreamProgressServer) error {
	if req.ChainId == "" {
		return status.Error(codes.InvalidArgument, "chain_id is required")
	}

	if s.eventBus == nil {
		return status.Error(codes.Unimplemented, "event bus not configured")
	}

	// For now, periodically send progress updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			// Get current progress
			progressResp, err := s.GetProgress(stream.Context(), &indexerv1.GetProgressRequest{
				ChainId: req.ChainId,
			})
			if err != nil {
				continue
			}
			if err := stream.Send(progressResp.Progress); err != nil {
				return status.Errorf(codes.Internal, "failed to send progress: %v", err)
			}
		}
	}
}

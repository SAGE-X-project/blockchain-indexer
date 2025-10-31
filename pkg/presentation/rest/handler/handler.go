package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sage-x-project/blockchain-indexer/pkg/application/indexer"
	"github.com/sage-x-project/blockchain-indexer/pkg/application/statistics"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"go.uber.org/zap"
)

// Handler handles REST API requests
type Handler struct {
	blockRepo       repository.BlockRepository
	txRepo          repository.TransactionRepository
	chainRepo       repository.ChainRepository
	progressTracker *indexer.ProgressTracker
	gapRecovery     map[string]*indexer.GapRecovery
	statsCollector  *statistics.Collector
	logger          *logger.Logger
	startTime       time.Time
}

// NewHandler creates a new REST API handler
func NewHandler(
	blockRepo repository.BlockRepository,
	txRepo repository.TransactionRepository,
	chainRepo repository.ChainRepository,
	progressTracker *indexer.ProgressTracker,
	gapRecovery map[string]*indexer.GapRecovery,
	statsCollector *statistics.Collector,
	logger *logger.Logger,
) *Handler {
	return &Handler{
		blockRepo:       blockRepo,
		txRepo:          txRepo,
		chainRepo:       chainRepo,
		progressTracker: progressTracker,
		gapRecovery:     gapRecovery,
		statsCollector:  statsCollector,
		logger:          logger,
		startTime:       time.Now(),
	}
}

// Helper functions

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
		Code:    status,
	})
}

// Health check

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(h.startTime)

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    uptime.String(),
		Checks: map[string]string{
			"database": "ok",
			"storage":  "ok",
		},
	}

	h.respondJSON(w, http.StatusOK, response)
}

// Chain handlers

func (h *Handler) GetChain(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")

	chain, err := h.chainRepo.GetChain(r.Context(), chainID)
	if err != nil {
		h.logger.Error("failed to get chain", zap.String("chain_id", chainID), zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve chain")
		return
	}

	if chain == nil {
		h.respondError(w, http.StatusNotFound, "Chain not found")
		return
	}

	response := ChainResponse{
		ChainID:            chain.ChainID,
		ChainType:          string(chain.ChainType),
		Name:               chain.Name,
		Network:            chain.Network,
		Status:             "active",
		StartBlock:         chain.StartBlock,
		LatestIndexedBlock: chain.LatestIndexedBlock,
		LatestChainBlock:   chain.LatestChainBlock,
		LastUpdated:        chain.LastUpdated,
	}

	h.respondJSON(w, http.StatusOK, response)
}

// Block handlers

func (h *Handler) GetBlock(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	blockNumStr := chi.URLParam(r, "number")

	blockNum, err := strconv.ParseUint(blockNumStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid block number")
		return
	}

	block, err := h.blockRepo.GetBlock(r.Context(), chainID, blockNum)
	if err != nil {
		h.logger.Error("failed to get block",
			zap.String("chain_id", chainID),
			zap.Uint64("block_number", blockNum),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve block")
		return
	}

	if block == nil {
		h.respondError(w, http.StatusNotFound, "Block not found")
		return
	}

	response := h.convertBlock(block)
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) GetBlockByHash(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	hash := chi.URLParam(r, "hash")

	block, err := h.blockRepo.GetBlockByHash(r.Context(), chainID, hash)
	if err != nil {
		h.logger.Error("failed to get block by hash",
			zap.String("chain_id", chainID),
			zap.String("hash", hash),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve block")
		return
	}

	if block == nil {
		h.respondError(w, http.StatusNotFound, "Block not found")
		return
	}

	response := h.convertBlock(block)
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) GetLatestBlock(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")

	block, err := h.blockRepo.GetLatestBlock(r.Context(), chainID)
	if err != nil {
		h.logger.Error("failed to get latest block",
			zap.String("chain_id", chainID),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve latest block")
		return
	}

	if block == nil {
		h.respondError(w, http.StatusNotFound, "No blocks found")
		return
	}

	response := h.convertBlock(block)
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) ListBlocks(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")

	// Parse query parameters
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	limitStr := r.URL.Query().Get("limit")

	var start, end uint64 = 0, 100
	var limit int = 10

	if startStr != "" {
		parsed, err := strconv.ParseUint(startStr, 10, 64)
		if err == nil {
			start = parsed
		}
	}

	if endStr != "" {
		parsed, err := strconv.ParseUint(endStr, 10, 64)
		if err == nil {
			end = parsed
		}
	}

	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Adjust end based on limit
	if end > start+uint64(limit) {
		end = start + uint64(limit)
	}

	blocks, err := h.blockRepo.GetBlocks(r.Context(), chainID, start, end)
	if err != nil {
		h.logger.Error("failed to list blocks",
			zap.String("chain_id", chainID),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve blocks")
		return
	}

	responses := make([]BlockResponse, 0, len(blocks))
	for _, block := range blocks {
		responses = append(responses, h.convertBlock(block))
	}

	h.respondJSON(w, http.StatusOK, responses)
}

// Transaction handlers

func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	hash := chi.URLParam(r, "hash")

	tx, err := h.txRepo.GetTransaction(r.Context(), chainID, hash)
	if err != nil {
		h.logger.Error("failed to get transaction",
			zap.String("chain_id", chainID),
			zap.String("hash", hash),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve transaction")
		return
	}

	if tx == nil {
		h.respondError(w, http.StatusNotFound, "Transaction not found")
		return
	}

	response := h.convertTransaction(tx)
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) ListTransactionsByBlock(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	blockNumStr := chi.URLParam(r, "number")

	blockNum, err := strconv.ParseUint(blockNumStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid block number")
		return
	}

	txs, err := h.txRepo.GetTransactionsByBlock(r.Context(), chainID, blockNum)
	if err != nil {
		h.logger.Error("failed to list transactions by block",
			zap.String("chain_id", chainID),
			zap.Uint64("block_number", blockNum),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve transactions")
		return
	}

	responses := make([]TransactionResponse, 0, len(txs))
	for _, tx := range txs {
		responses = append(responses, h.convertTransaction(tx))
	}

	h.respondJSON(w, http.StatusOK, responses)
}

func (h *Handler) ListTransactionsByAddress(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	address := chi.URLParam(r, "address")

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	paginationOpts := &models.PaginationOptions{
		Limit:  limit,
		Offset: 0,
	}

	txs, err := h.txRepo.GetTransactionsByAddress(r.Context(), chainID, address, paginationOpts)
	if err != nil {
		h.logger.Error("failed to list transactions by address",
			zap.String("chain_id", chainID),
			zap.String("address", address),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve transactions")
		return
	}

	responses := make([]TransactionResponse, 0, len(txs))
	for _, tx := range txs {
		responses = append(responses, h.convertTransaction(tx))
	}

	h.respondJSON(w, http.StatusOK, responses)
}

// Progress handlers

func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")

	if h.progressTracker == nil {
		h.respondError(w, http.StatusServiceUnavailable, "Progress tracking not available")
		return
	}

	progress, err := h.progressTracker.GetProgress(r.Context(), chainID)
	if err != nil {
		h.logger.Error("failed to get progress",
			zap.String("chain_id", chainID),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve progress")
		return
	}

	response := ProgressResponse{
		ChainID:            progress.ChainID,
		ChainType:          progress.ChainType,
		LatestIndexedBlock: progress.LatestIndexedBlock,
		LatestChainBlock:   progress.LatestChainBlock,
		TargetBlock:        progress.TargetBlock,
		StartBlock:         progress.StartBlock,
		BlocksBehind:       progress.BlocksBehind,
		ProgressPercentage: progress.ProgressPercentage,
		BlocksPerSecond:    progress.BlocksPerSecond,
		EstimatedTimeLeft:  progress.EstimatedTimeLeft.String(),
		LastUpdated:        progress.LastUpdated,
		Status:             progress.Status,
	}

	h.respondJSON(w, http.StatusOK, response)
}

// Converter functions

func (h *Handler) convertBlock(block *models.Block) BlockResponse {
	response := BlockResponse{
		ChainID:    block.ChainID,
		ChainType:  string(block.ChainType),
		Number:     block.Number,
		Hash:       block.Hash,
		ParentHash: block.ParentHash,
		Timestamp:  block.Timestamp.Time,
		GasUsed:    block.GasUsed,
		GasLimit:   block.GasLimit,
		Miner:      block.Proposer,
		TxCount:    block.TxCount,
		IndexedAt:  block.IndexedAt,
	}

	if len(block.Transactions) > 0 {
		response.Transactions = make([]TransactionResponse, 0, len(block.Transactions))
		for _, tx := range block.Transactions {
			response.Transactions = append(response.Transactions, h.convertTransaction(tx))
		}
	}

	return response
}

func (h *Handler) convertTransaction(tx *models.Transaction) TransactionResponse {
	response := TransactionResponse{
		ChainID:        tx.ChainID,
		Hash:           tx.Hash,
		BlockNumber:    tx.BlockNumber,
		BlockHash:      tx.BlockHash,
		BlockTimestamp: tx.Timestamp.Time,
		TxIndex:        tx.Index,
		From:           tx.From,
		To:             tx.To,
		Value:          tx.Value,
		GasPrice:       tx.GasPrice,
		GasUsed:        tx.GasUsed,
		Nonce:          tx.Nonce,
		Status:         string(tx.Status),
		IndexedAt:      tx.IndexedAt,
	}

	if len(tx.Input) > 0 {
		response.Input = fmt.Sprintf("0x%x", tx.Input)
	}

	if tx.ContractAddress != "" {
		response.ContractAddress = tx.ContractAddress
	}

	if len(tx.Logs) > 0 {
		response.Logs = make([]LogResponse, 0, len(tx.Logs))
		for _, log := range tx.Logs {
			response.Logs = append(response.Logs, LogResponse{
				Address:  log.Address,
				Topics:   log.Topics,
				Data:     fmt.Sprintf("0x%x", log.Data),
				LogIndex: log.Index,
			})
		}
	}

	return response
}

// GetChainGaps handles GET /chains/{chainID}/gaps
func (h *Handler) GetChainGaps(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	if chainID == "" {
		h.respondError(w, http.StatusBadRequest, "chain_id is required")
		return
	}

	// Check if gap recovery is available for this chain
	if h.gapRecovery == nil {
		h.respondJSON(w, http.StatusOK, GapsResponse{Gaps: []GapInfo{}})
		return
	}

	recovery, ok := h.gapRecovery[chainID]
	if !ok {
		h.respondJSON(w, http.StatusOK, GapsResponse{Gaps: []GapInfo{}})
		return
	}

	// Detect gaps
	gaps, err := recovery.DetectGaps(r.Context(), chainID)
	if err != nil {
		h.logger.Error("failed to detect gaps",
			zap.String("chain_id", chainID),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to detect gaps")
		return
	}

	// Convert to response format
	gapInfos := make([]GapInfo, len(gaps))
	for i, gap := range gaps {
		gapInfos[i] = GapInfo{
			ChainID:    gap.ChainID,
			StartBlock: gap.StartBlock,
			EndBlock:   gap.EndBlock,
			Size:       gap.Size,
		}
	}

	h.respondJSON(w, http.StatusOK, GapsResponse{
		Gaps:  gapInfos,
		Count: len(gapInfos),
	})
}

// ListChains handles GET /chains
func (h *Handler) ListChains(w http.ResponseWriter, r *http.Request) {
	chains, err := h.chainRepo.GetAllChains(r.Context())
	if err != nil {
		h.logger.Error("failed to list chains", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to list chains")
		return
	}

	// Convert to response format
	responses := make([]ChainResponse, len(chains))
	for i, chain := range chains {
		responses[i] = ChainResponse{
			ChainID:            chain.ChainID,
			ChainType:          string(chain.ChainType),
			Name:               chain.Name,
			Network:            chain.Network,
			Status:             string(chain.Status),
			StartBlock:         chain.StartBlock,
			LatestIndexedBlock: chain.LatestIndexedBlock,
			LatestChainBlock:   chain.LatestChainBlock,
			LastUpdated:        chain.LastUpdated,
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"chains": responses,
		"count":  len(responses),
	})
}

// GetChainStats handles GET /chains/{chainID}/stats
func (h *Handler) GetChainStats(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	if chainID == "" {
		h.respondError(w, http.StatusBadRequest, "chain_id is required")
		return
	}

	// Check if stats collector is available
	if h.statsCollector == nil {
		h.respondJSON(w, http.StatusOK, StatsResponse{
			TotalBlocks:       0,
			TotalTransactions: 0,
			ChainsIndexed:     0,
			AverageBlockTime:  0,
			AverageTxPerBlock: 0,
		})
		return
	}

	// Get chain statistics
	stats, err := h.statsCollector.GetChainStatistics(r.Context(), chainID)
	if err != nil {
		h.logger.Error("failed to get chain stats",
			zap.String("chain_id", chainID),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to get statistics")
		return
	}

	h.respondJSON(w, http.StatusOK, StatsResponse{
		TotalBlocks:       stats.TotalBlocks,
		TotalTransactions: stats.TotalTransactions,
		ChainsIndexed:     1,
		AverageBlockTime:  stats.AverageBlockTime,
		AverageTxPerBlock: stats.AverageTxPerBlock,
	})
}

// GetGlobalStats handles GET /stats
func (h *Handler) GetGlobalStats(w http.ResponseWriter, r *http.Request) {
	// Check if stats collector is available
	if h.statsCollector == nil {
		h.respondJSON(w, http.StatusOK, StatsResponse{
			TotalBlocks:       0,
			TotalTransactions: 0,
			ChainsIndexed:     0,
			AverageBlockTime:  0,
			AverageTxPerBlock: 0,
		})
		return
	}

	// Get global statistics
	stats, err := h.statsCollector.GetGlobalStatistics(r.Context())
	if err != nil {
		h.logger.Error("failed to get global stats", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to get global statistics")
		return
	}

	h.respondJSON(w, http.StatusOK, StatsResponse{
		TotalBlocks:       stats.TotalBlocks,
		TotalTransactions: stats.TotalTransactions,
		ChainsIndexed:     stats.TotalChains,
		AverageBlockTime:  stats.AverageBlockTime,
		AverageTxPerBlock: stats.AverageTxPerBlock,
	})
}

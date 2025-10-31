package server

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	indexerv1 "github.com/sage-x-project/blockchain-indexer/api/proto/indexer/v1"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// convertChainToProto converts a domain Chain to proto Chain
func convertChainToProto(chain *models.Chain) *indexerv1.Chain {
	return &indexerv1.Chain{
		ChainId:            chain.ChainID,
		ChainType:          convertChainTypeToProto(chain.ChainType),
		Name:               chain.Name,
		Network:            chain.Network,
		Status:             convertChainStatusToProto(chain.Status),
		StartBlock:         chain.StartBlock,
		LatestIndexedBlock: chain.LatestIndexedBlock,
		LatestChainBlock:   chain.LatestChainBlock,
		LastUpdated:        timestamppb.New(chain.LastUpdated),
	}
}

// convertBlockToProto converts a domain Block to proto Block
func convertBlockToProto(block *models.Block) *indexerv1.Block {
	protoBlock := &indexerv1.Block{
		ChainId:      block.ChainID,
		ChainType:    convertChainTypeToProto(block.ChainType),
		Number:       block.Number,
		Hash:         block.Hash,
		ParentHash:   block.ParentHash,
		Timestamp:    timestamppb.New(block.Timestamp.Time),
		GasUsed:      block.GasUsed,
		GasLimit:     block.GasLimit,
		Miner:        block.Proposer, // Proposer is the equivalent of Miner
		TxCount:      int32(block.TxCount),
		IndexedAt:    timestamppb.New(block.IndexedAt),
		Transactions: make([]*indexerv1.Transaction, 0),
	}

	// Convert transactions if present
	if len(block.Transactions) > 0 {
		protoBlock.Transactions = make([]*indexerv1.Transaction, len(block.Transactions))
		for i, tx := range block.Transactions {
			protoBlock.Transactions[i] = convertTransactionToProto(tx)
		}
	}

	return protoBlock
}

// convertTransactionToProto converts a domain Transaction to proto Transaction
func convertTransactionToProto(tx *models.Transaction) *indexerv1.Transaction {
	protoTx := &indexerv1.Transaction{
		ChainId:         tx.ChainID,
		Hash:            tx.Hash,
		BlockNumber:     tx.BlockNumber,
		BlockHash:       tx.BlockHash,
		BlockTimestamp:  timestamppb.New(tx.Timestamp.Time),
		TxIndex:         uint32(tx.Index),
		From:            tx.From,
		To:              tx.To,
		Value:           tx.Value,
		GasPrice:        tx.GasPrice,
		GasUsed:         tx.GasUsed,
		Nonce:           tx.Nonce,
		Status:          convertTransactionStatusToProto(tx.Status),
		ContractAddress: tx.ContractAddress,
		IndexedAt:       timestamppb.New(tx.IndexedAt),
		Logs:            make([]*indexerv1.Log, 0),
	}

	// Convert input data
	if len(tx.Input) > 0 {
		protoTx.Input = tx.Input
	}

	// Convert logs if present
	if len(tx.Logs) > 0 {
		protoTx.Logs = make([]*indexerv1.Log, len(tx.Logs))
		for i, log := range tx.Logs {
			protoTx.Logs[i] = convertLogToProto(log)
		}
	}

	return protoTx
}

// convertLogToProto converts a domain Log to proto Log
func convertLogToProto(log *models.Log) *indexerv1.Log {
	return &indexerv1.Log{
		Address:  log.Address,
		Topics:   log.Topics,
		Data:     log.Data,
		LogIndex: uint32(log.Index),
	}
}

// convertChainTypeToProto converts domain ChainType to proto ChainType
func convertChainTypeToProto(chainType models.ChainType) indexerv1.ChainType {
	switch chainType {
	case models.ChainTypeEVM:
		return indexerv1.ChainType_CHAIN_TYPE_EVM
	case models.ChainTypeSolana:
		return indexerv1.ChainType_CHAIN_TYPE_SOLANA
	case models.ChainTypeCosmos:
		return indexerv1.ChainType_CHAIN_TYPE_COSMOS
	default:
		return indexerv1.ChainType_CHAIN_TYPE_UNSPECIFIED
	}
}

// convertChainStatusToProto converts domain ChainStatus to proto ChainStatus
func convertChainStatusToProto(status models.ChainStatus) indexerv1.ChainStatus {
	switch status {
	case models.ChainStatusLive:
		return indexerv1.ChainStatus_CHAIN_STATUS_ACTIVE
	case models.ChainStatusIdle:
		return indexerv1.ChainStatus_CHAIN_STATUS_INACTIVE
	case models.ChainStatusSyncing:
		return indexerv1.ChainStatus_CHAIN_STATUS_SYNCING
	case models.ChainStatusError:
		return indexerv1.ChainStatus_CHAIN_STATUS_ERROR
	case models.ChainStatusPaused:
		return indexerv1.ChainStatus_CHAIN_STATUS_INACTIVE
	default:
		return indexerv1.ChainStatus_CHAIN_STATUS_UNSPECIFIED
	}
}

// convertTransactionStatusToProto converts domain TxStatus to proto TransactionStatus
func convertTransactionStatusToProto(status models.TxStatus) indexerv1.TransactionStatus {
	switch status {
	case models.TxStatusPending:
		return indexerv1.TransactionStatus_TRANSACTION_STATUS_PENDING
	case models.TxStatusSuccess:
		return indexerv1.TransactionStatus_TRANSACTION_STATUS_SUCCESS
	case models.TxStatusFailed:
		return indexerv1.TransactionStatus_TRANSACTION_STATUS_FAILED
	default:
		return indexerv1.TransactionStatus_TRANSACTION_STATUS_UNSPECIFIED
	}
}

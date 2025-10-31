package evm

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// Normalizer converts EVM-specific data structures to domain models
type Normalizer struct {
	chainID string
	network string
}

// NewNormalizer creates a new normalizer
func NewNormalizer(chainID, network string) *Normalizer {
	return &Normalizer{
		chainID: chainID,
		network: network,
	}
}

// NormalizeBlock converts a go-ethereum Block to a domain Block
func (n *Normalizer) NormalizeBlock(block *types.Block, receipts []*types.Receipt) (*models.Block, error) {
	if block == nil {
		return nil, fmt.Errorf("block is nil")
	}

	// Build transactions map from receipts for quick lookup
	receiptsMap := make(map[common.Hash]*types.Receipt)
	if receipts != nil {
		for _, receipt := range receipts {
			receiptsMap[receipt.TxHash] = receipt
		}
	}

	// Normalize transactions
	transactions := make([]*models.Transaction, 0, len(block.Transactions()))
	for i, tx := range block.Transactions() {
		domainTx, err := n.NormalizeTransaction(tx, block, receiptsMap[tx.Hash()], uint64(i))
		if err != nil {
			return nil, fmt.Errorf("failed to normalize transaction %s: %w", tx.Hash().Hex(), err)
		}
		transactions = append(transactions, domainTx)
	}

	// Build metadata
	metadata := make(map[string]interface{})
	metadata["difficulty"] = block.Difficulty().String()
	metadata["total_difficulty"] = "0" // Not available in block
	metadata["nonce"] = fmt.Sprintf("0x%x", block.Nonce())
	metadata["sha3_uncles"] = block.UncleHash().Hex()
	metadata["transactions_root"] = block.TxHash().Hex()
	metadata["state_root"] = block.Root().Hex()
	metadata["receipts_root"] = block.ReceiptHash().Hex()
	metadata["extra_data"] = "0x" + hex.EncodeToString(block.Extra())
	metadata["uncles"] = formatUncles(block.Uncles())

	if block.BaseFee() != nil {
		metadata["base_fee_per_gas"] = block.BaseFee().String()
	}

	domainBlock := &models.Block{
		ChainType:    models.ChainTypeEVM,
		ChainID:      n.chainID,
		Number:       block.NumberU64(),
		Hash:         block.Hash().Hex(),
		ParentHash:   block.ParentHash().Hex(),
		Timestamp:    models.NewTimestamp(int64(block.Time())),
		TxCount:      len(transactions),
		Transactions: transactions,
		Size:         block.Size(),
		GasLimit:     block.GasLimit(),
		GasUsed:      block.GasUsed(),
		Proposer:     block.Coinbase().Hex(),
		Metadata:     metadata,
	}

	return domainBlock, nil
}

// NormalizeTransaction converts a go-ethereum Transaction to a domain Transaction
func (n *Normalizer) NormalizeTransaction(
	tx *types.Transaction,
	block *types.Block,
	receipt *types.Receipt,
	index uint64,
) (*models.Transaction, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}

	// Get sender address
	signer := types.LatestSignerForChainID(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender: %w", err)
	}

	// Get destination address
	to := ""
	if tx.To() != nil {
		to = tx.To().Hex()
	}

	// Build metadata
	metadata := make(map[string]interface{})
	metadata["nonce"] = tx.Nonce()
	metadata["gas_price"] = tx.GasPrice().String()

	if tx.Type() == types.DynamicFeeTxType {
		if tx.GasFeeCap() != nil {
			metadata["max_fee_per_gas"] = tx.GasFeeCap().String()
		}
		if tx.GasTipCap() != nil {
			metadata["max_priority_fee_per_gas"] = tx.GasTipCap().String()
		}
	}

	metadata["input"] = "0x" + hex.EncodeToString(tx.Data())
	metadata["tx_type"] = tx.Type()

	if tx.ChainId() != nil {
		metadata["chain_id"] = tx.ChainId().String()
	}

	// Add access list for EIP-2930 and EIP-1559 transactions
	if tx.Type() == types.AccessListTxType || tx.Type() == types.DynamicFeeTxType {
		if accessList := tx.AccessList(); accessList != nil {
			metadata["access_list"] = formatAccessList(accessList)
		}
	}

	// Add receipt data if available
	status := models.TxStatusPending
	if receipt != nil {
		if receipt.Status == types.ReceiptStatusSuccessful {
			status = models.TxStatusSuccess
		} else {
			status = models.TxStatusFailed
		}

		metadata["cumulative_gas_used"] = receipt.CumulativeGasUsed
		metadata["effective_gas_price"] = receipt.EffectiveGasPrice.String()
		metadata["logs_bloom"] = "0x" + hex.EncodeToString(receipt.Bloom[:])

		if receipt.ContractAddress != (common.Address{}) {
			metadata["contract_address"] = receipt.ContractAddress.Hex()
		}

		// Add logs
		if len(receipt.Logs) > 0 {
			logs := make([]map[string]interface{}, 0, len(receipt.Logs))
			for _, log := range receipt.Logs {
				logData := map[string]interface{}{
					"address": log.Address.Hex(),
					"topics":  formatTopics(log.Topics),
					"data":    "0x" + hex.EncodeToString(log.Data),
					"index":   log.Index,
				}
				logs = append(logs, logData)
			}
			metadata["logs"] = logs
		}
	}

	// Signature components
	v, r, s := tx.RawSignatureValues()
	metadata["v"] = v.String()
	metadata["r"] = r.String()
	metadata["s"] = s.String()

	// Calculate gas used
	gasUsed := uint64(0)
	if receipt != nil {
		gasUsed = receipt.GasUsed
	}

	// Calculate fee (gasUsed * gasPrice)
	fee := "0"
	if receipt != nil && receipt.EffectiveGasPrice != nil {
		feeVal := new(big.Int).Mul(new(big.Int).SetUint64(gasUsed), receipt.EffectiveGasPrice)
		fee = feeVal.String()
	}

	domainTx := &models.Transaction{
		ChainType:   models.ChainTypeEVM,
		ChainID:     n.chainID,
		Hash:        tx.Hash().Hex(),
		BlockNumber: block.NumberU64(),
		BlockHash:   block.Hash().Hex(),
		Index:       index,
		From:        from.Hex(),
		To:          to,
		Value:       tx.Value().String(),
		Fee:         fee,
		GasUsed:     gasUsed,
		GasPrice:    tx.GasPrice().String(),
		Nonce:       tx.Nonce(),
		Type:        tx.Type(),
		Status:      status,
		Input:       tx.Data(),
		Timestamp:   models.NewTimestamp(int64(block.Time())),
		Metadata:    metadata,
	}

	return domainTx, nil
}

// NormalizeEVMBlock and NormalizeEVMTransaction are not currently used
// as we use go-ethereum's native types directly. They can be implemented
// if needed for custom EVM block/transaction types in the future.

// Helper functions

func formatUncles(uncles []*types.Header) []string {
	if len(uncles) == 0 {
		return []string{}
	}
	result := make([]string, len(uncles))
	for i, uncle := range uncles {
		result[i] = uncle.Hash().Hex()
	}
	return result
}

func formatTopics(topics []common.Hash) []string {
	if len(topics) == 0 {
		return []string{}
	}
	result := make([]string, len(topics))
	for i, topic := range topics {
		result[i] = topic.Hex()
	}
	return result
}

func formatAccessList(accessList types.AccessList) []map[string]interface{} {
	if len(accessList) == 0 {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, len(accessList))
	for i, entry := range accessList {
		keys := make([]string, len(entry.StorageKeys))
		for j, key := range entry.StorageKeys {
			keys[j] = key.Hex()
		}
		result[i] = map[string]interface{}{
			"address":      entry.Address.Hex(),
			"storage_keys": keys,
		}
	}
	return result
}

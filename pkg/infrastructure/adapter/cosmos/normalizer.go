package cosmos

import (
	"encoding/hex"
	"fmt"
	"strings"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// Normalizer normalizes Cosmos/Tendermint data to domain models
type Normalizer struct {
	chainID string
	network string
}

// NewNormalizer creates a new Normalizer
func NewNormalizer(chainID, network string) *Normalizer {
	return &Normalizer{
		chainID: chainID,
		network: network,
	}
}

// NormalizeChainInfo returns normalized chain info
func (n *Normalizer) NormalizeChainInfo() *models.ChainInfo {
	return &models.ChainInfo{
		ChainType: models.ChainTypeCosmos,
		ChainID:   n.chainID,
		Name:      n.chainID,
		Network:   n.network,
	}
}

// NormalizeBlock normalizes a Cosmos block to domain Block
func (n *Normalizer) NormalizeBlock(block *tmtypes.Block, blockResults *coretypes.ResultBlockResults) (*models.Block, error) {
	if block == nil {
		return nil, fmt.Errorf("block is nil")
	}

	// Extract block hash
	blockHash := strings.ToUpper(hex.EncodeToString(block.Hash()))

	// Extract parent hash
	var parentHash string
	if block.Height > 1 {
		parentHash = strings.ToUpper(hex.EncodeToString(block.LastBlockID.Hash))
	}

	// Create block metadata
	metadata := make(map[string]interface{})
	metadata["proposer_address"] = strings.ToUpper(hex.EncodeToString(block.ProposerAddress))
	metadata["chain_id"] = block.ChainID
	metadata["num_txs"] = len(block.Txs)
	metadata["total_gas"] = int64(0)

	// Extract gas information from block results if available
	if blockResults != nil && blockResults.TxsResults != nil {
		var totalGasUsed int64
		var totalGasWanted int64
		for _, txResult := range blockResults.TxsResults {
			totalGasUsed += txResult.GasUsed
			totalGasWanted += txResult.GasWanted
		}
		metadata["total_gas_used"] = totalGasUsed
		metadata["total_gas_wanted"] = totalGasWanted
	}

	// Normalize transactions
	transactions := make([]*models.Transaction, 0, len(block.Txs))
	for i, tx := range block.Txs {
		normalizedTx, err := n.NormalizeTransaction(tx, block.Height, uint32(i), blockResults)
		if err != nil {
			// Log error but continue processing
			continue
		}
		transactions = append(transactions, normalizedTx)
	}

	return &models.Block{
		ChainID:      n.chainID,
		Number:       uint64(block.Height),
		Hash:         blockHash,
		ParentHash:   parentHash,
		Timestamp:    models.NewTimestamp(block.Time.Unix()),
		Transactions: transactions,
		Metadata:     metadata,
	}, nil
}

// NormalizeTransaction normalizes a Cosmos transaction to domain Transaction
func (n *Normalizer) NormalizeTransaction(
	tx tmtypes.Tx,
	blockHeight int64,
	txIndex uint32,
	blockResults *coretypes.ResultBlockResults,
) (*models.Transaction, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}

	// Calculate transaction hash
	txHash := strings.ToUpper(hex.EncodeToString(tx.Hash()))

	// Default values
	var status models.TxStatus = models.TxStatusSuccess
	var gasUsed uint64 = 0
	metadata := make(map[string]interface{})

	// Extract transaction result if available
	if blockResults != nil && blockResults.TxsResults != nil && int(txIndex) < len(blockResults.TxsResults) {
		txResult := blockResults.TxsResults[txIndex]

		// Determine status
		if txResult.Code != 0 {
			status = models.TxStatusFailed
		}

		gasUsed = uint64(txResult.GasUsed)
		metadata["gas_wanted"] = txResult.GasWanted
		metadata["code"] = txResult.Code
		metadata["codespace"] = txResult.Codespace
		metadata["log"] = txResult.Log
		metadata["info"] = txResult.Info

		// Add events if available
		if len(txResult.Events) > 0 {
			events := make([]map[string]interface{}, 0, len(txResult.Events))
			for _, event := range txResult.Events {
				eventData := map[string]interface{}{
					"type": event.Type,
				}

				attributes := make([]map[string]string, 0, len(event.Attributes))
				for _, attr := range event.Attributes {
					attributes = append(attributes, map[string]string{
						"key":   attr.Key,
						"value": attr.Value,
					})
				}
				eventData["attributes"] = attributes
				events = append(events, eventData)
			}
			metadata["events"] = events
		}
	}

	// Store raw transaction
	metadata["raw_tx"] = hex.EncodeToString(tx)
	metadata["tx_index"] = txIndex

	return &models.Transaction{
		ChainID:     n.chainID,
		Hash:        txHash,
		BlockNumber: uint64(blockHeight),
		From:        "", // Cosmos doesn't have a simple "from" address
		To:          "", // Will be extracted from tx decoding if needed
		Value:       "0",
		GasUsed:     gasUsed,
		GasPrice:    "0",
		Status:      status,
		Timestamp:   models.NewTimestamp(0), // Will be set from block timestamp
		Metadata:    metadata,
	}, nil
}

// NormalizeBlockMeta normalizes block metadata
func (n *Normalizer) NormalizeBlockMeta(blockMeta *tmtypes.BlockMeta) (*models.Block, error) {
	if blockMeta == nil {
		return nil, fmt.Errorf("block meta is nil")
	}

	blockHash := strings.ToUpper(hex.EncodeToString(blockMeta.BlockID.Hash))
	var parentHash string
	if blockMeta.Header.Height > 1 {
		parentHash = strings.ToUpper(hex.EncodeToString(blockMeta.Header.LastBlockID.Hash))
	}

	metadata := make(map[string]interface{})
	metadata["proposer_address"] = strings.ToUpper(hex.EncodeToString(blockMeta.Header.ProposerAddress))
	metadata["chain_id"] = blockMeta.Header.ChainID

	return &models.Block{
		ChainID:      n.chainID,
		Number:       uint64(blockMeta.Header.Height),
		Hash:         blockHash,
		ParentHash:   parentHash,
		Timestamp:    models.NewTimestamp(blockMeta.Header.Time.Unix()),
		Transactions: []*models.Transaction{}, // No transactions in metadata
		Metadata:     metadata,
	}, nil
}

// NormalizeValidators normalizes validators
func (n *Normalizer) NormalizeValidators(validators []*tmtypes.Validator) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(validators))

	for _, val := range validators {
		validatorData := map[string]interface{}{
			"address":      strings.ToUpper(hex.EncodeToString(val.Address)),
			"voting_power": val.VotingPower,
			"proposer_priority": val.ProposerPriority,
		}

		if val.PubKey != nil {
			validatorData["pub_key"] = hex.EncodeToString(val.PubKey.Bytes())
		}

		result = append(result, validatorData)
	}

	return result
}

// ExtractAddressesFromTx extracts addresses from a transaction (placeholder)
// This would require transaction decoding which depends on the specific Cosmos chain
func (n *Normalizer) ExtractAddressesFromTx(tx tmtypes.Tx) (from, to string, err error) {
	// This is a placeholder - actual implementation would need to decode the transaction
	// based on the specific Cosmos SDK version and chain configuration
	return "", "", nil
}

// NormalizeEvents normalizes transaction events (helper function for future use)
func (n *Normalizer) NormalizeEvents(eventType string, attributes map[string]string) map[string]interface{} {
	eventData := map[string]interface{}{
		"type":       eventType,
		"attributes": attributes,
	}

	return eventData
}

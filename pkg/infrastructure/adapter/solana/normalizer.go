package solana

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// Normalizer converts Solana-specific types to domain models
type Normalizer struct {
	chainID string
	network string
}

// NewNormalizer creates a new Solana normalizer
func NewNormalizer(chainID, network string) *Normalizer {
	return &Normalizer{
		chainID: chainID,
		network: network,
	}
}

// NormalizeBlock converts a Solana block to a domain Block
func (n *Normalizer) NormalizeBlock(slot uint64, block *GetBlockResponse) (*models.Block, error) {
	if block == nil {
		return nil, fmt.Errorf("block is nil")
	}

	// Calculate block time
	var blockTime time.Time
	if block.BlockTime != nil {
		blockTime = time.Unix(*block.BlockTime, 0)
	} else {
		blockTime = time.Now()
	}

	// Extract transactions
	transactions := make([]*models.Transaction, 0, len(block.Transactions))
	for i, tx := range block.Transactions {
		domainTx, err := n.NormalizeTransaction(slot, block.Blockhash, blockTime, uint64(i), &tx)
		if err != nil {
			// Log error but continue processing other transactions
			continue
		}
		transactions = append(transactions, domainTx)
	}

	// Create domain block
	domainBlock := &models.Block{
		ChainID:      n.chainID,
		ChainType:    models.ChainTypeSolana,
		Number:       slot,
		Hash:         block.Blockhash,
		ParentHash:   block.PreviousBlockhash,
		Timestamp:    models.NewTimestamp(blockTime.Unix()),
		Transactions: transactions,
		TxCount:      len(transactions),
		TxHashes:     make([]string, 0, len(transactions)),
		Size:         0, // Solana doesn't provide block size in the same way
		GasUsed:      0, // Solana uses compute units, not gas
		GasLimit:     0,
		Proposer:     "", // Solana doesn't expose validator in block response
		Metadata:     make(map[string]interface{}),
	}

	// Extract transaction hashes
	for _, tx := range transactions {
		domainBlock.TxHashes = append(domainBlock.TxHashes, tx.Hash)
	}

	// Add Solana-specific data
	if block.BlockHeight != nil {
		domainBlock.Metadata["block_height"] = *block.BlockHeight
	}
	domainBlock.Metadata["parent_slot"] = block.ParentSlot

	// Add rewards if present
	if len(block.Rewards) > 0 {
		rewardsData := make([]map[string]interface{}, len(block.Rewards))
		for i, reward := range block.Rewards {
			rewardsData[i] = map[string]interface{}{
				"pubkey":       reward.Pubkey,
				"lamports":     reward.Lamports,
				"post_balance": reward.PostBalance,
			}
			if reward.RewardType != nil {
				rewardsData[i]["reward_type"] = *reward.RewardType
			}
			if reward.Commission != nil {
				rewardsData[i]["commission"] = *reward.Commission
			}
		}
		domainBlock.Metadata["rewards"] = rewardsData
	}

	return domainBlock, nil
}

// NormalizeTransaction converts a Solana transaction to a domain Transaction
func (n *Normalizer) NormalizeTransaction(
	slot uint64,
	blockhash string,
	blockTime time.Time,
	index uint64,
	txWithMeta *TransactionWithMeta,
) (*models.Transaction, error) {
	if txWithMeta == nil {
		return nil, fmt.Errorf("transaction is nil")
	}

	// Parse transaction - it can be Transaction or []string
	var tx *Transaction
	var signature string

	switch v := txWithMeta.Transaction.(type) {
	case map[string]interface{}:
		// Full transaction
		txBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal transaction: %w", err)
		}
		if err := json.Unmarshal(txBytes, &tx); err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
		}
		if len(tx.Signatures) > 0 {
			signature = tx.Signatures[0]
		}
	case []interface{}:
		// Signatures only
		if len(v) > 0 {
			if sig, ok := v[0].(string); ok {
				signature = sig
			}
		}
	default:
		return nil, fmt.Errorf("unexpected transaction type: %T", v)
	}

	// Determine status
	status := models.TxStatusSuccess
	var errorMsg string
	if txWithMeta.Meta != nil && txWithMeta.Meta.Err != nil {
		status = models.TxStatusFailed
		errBytes, _ := json.Marshal(txWithMeta.Meta.Err)
		errorMsg = string(errBytes)
	}

	// Calculate fee
	var fee uint64
	if txWithMeta.Meta != nil {
		fee = txWithMeta.Meta.Fee
	}

	// Extract accounts involved
	accounts := make([]string, 0)
	if tx != nil && len(tx.Message.AccountKeys) > 0 {
		accounts = tx.Message.AccountKeys
	}

	// Create domain transaction
	domainTx := &models.Transaction{
		ChainID:     n.chainID,
		ChainType:   models.ChainTypeSolana,
		Hash:        signature,
		BlockNumber: slot,
		BlockHash:   blockhash,
		Timestamp:   models.NewTimestamp(blockTime.Unix()),
		Index:       index,
		From:        "",
		To:          "",
		Value:       "0",
		Fee:         fmt.Sprintf("%d", fee),
		GasUsed:     0,
		GasPrice:    "0",
		Status:      status,
		Input:       nil,
		Nonce:       0,
		Type:        0,
		Logs:        make([]*models.Log, 0),
		Metadata:    make(map[string]interface{}),
	}

	// Set from/to based on account keys
	if len(accounts) > 0 {
		domainTx.From = accounts[0] // First account is typically the fee payer/signer
		if len(accounts) > 1 {
			domainTx.To = accounts[1] // Second account is often the recipient
		}
	}

	// Add Solana-specific metadata
	if txWithMeta.Meta != nil {
		meta := txWithMeta.Meta

		// Add compute units consumed
		if meta.ComputeUnitsConsumed != nil {
			domainTx.Metadata["compute_units_consumed"] = *meta.ComputeUnitsConsumed
		}

		// Add log messages as transaction logs
		for i, logMsg := range meta.LogMessages {
			domainTx.Logs = append(domainTx.Logs, &models.Log{
				Index:   uint64(i),
				Address: "",
				Topics:  []string{},
				Data:    []byte(logMsg),
			})
		}

		// Add balance changes
		if len(meta.PreBalances) > 0 && len(meta.PostBalances) > 0 {
			balanceChanges := make([]map[string]interface{}, 0)
			for i := 0; i < len(accounts) && i < len(meta.PreBalances) && i < len(meta.PostBalances); i++ {
				preBalance := meta.PreBalances[i]
				postBalance := meta.PostBalances[i]
				if preBalance != postBalance {
					balanceChanges = append(balanceChanges, map[string]interface{}{
						"account":      accounts[i],
						"pre_balance":  preBalance,
						"post_balance": postBalance,
						"change":       int64(postBalance) - int64(preBalance),
					})
				}
			}
			if len(balanceChanges) > 0 {
				domainTx.Metadata["balance_changes"] = balanceChanges
			}
		}

		// Add token balance changes
		if len(meta.PreTokenBalances) > 0 || len(meta.PostTokenBalances) > 0 {
			domainTx.Metadata["pre_token_balances"] = meta.PreTokenBalances
			domainTx.Metadata["post_token_balances"] = meta.PostTokenBalances
		}

		// Add inner instructions
		if len(meta.InnerInstructions) > 0 {
			domainTx.Metadata["inner_instructions"] = meta.InnerInstructions
		}

		// Add error if present
		if errorMsg != "" {
			domainTx.Metadata["error"] = errorMsg
		}
	}

	// Add transaction details if available
	if tx != nil {
		domainTx.Metadata["signatures"] = tx.Signatures
		domainTx.Metadata["recent_blockhash"] = tx.Message.RecentBlockhash
		domainTx.Metadata["instructions_count"] = len(tx.Message.Instructions)
		domainTx.Metadata["accounts"] = tx.Message.AccountKeys

		// Add instructions summary
		instructions := make([]map[string]interface{}, len(tx.Message.Instructions))
		for i, instr := range tx.Message.Instructions {
			programAccount := ""
			if int(instr.ProgramIdIndex) < len(accounts) {
				programAccount = accounts[instr.ProgramIdIndex]
			}
			instructions[i] = map[string]interface{}{
				"program_id_index": instr.ProgramIdIndex,
				"program_account":  programAccount,
				"accounts":         instr.Accounts,
				"data":             instr.Data,
			}
		}
		domainTx.Metadata["instructions"] = instructions
	}

	// Add version if present
	if txWithMeta.Version != nil {
		domainTx.Metadata["version"] = txWithMeta.Version
	}

	return domainTx, nil
}

// NormalizeChainInfo creates chain info for Solana
func (n *Normalizer) NormalizeChainInfo() *models.ChainInfo {
	return &models.ChainInfo{
		ChainType: models.ChainTypeSolana,
		ChainID:   n.chainID,
		Name:      "Solana",
		Network:   n.network,
	}
}

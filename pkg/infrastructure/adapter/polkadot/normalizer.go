package polkadot

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// Normalizer normalizes Polkadot/Substrate data to domain models
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
		ChainType: models.ChainTypePolkadot,
		ChainID:   n.chainID,
		Name:      n.chainID,
		Network:   n.network,
	}
}

// NormalizeBlock normalizes a Substrate block to domain Block
func (n *Normalizer) NormalizeBlock(signedBlock *types.SignedBlock, blockHash types.Hash) (*models.Block, error) {
	if signedBlock == nil {
		return nil, fmt.Errorf("signed block is nil")
	}

	block := signedBlock.Block
	header := block.Header

	// Extract block hash
	blockHashHex := strings.ToUpper(hex.EncodeToString(blockHash[:]))

	// Extract parent hash
	parentHashHex := strings.ToUpper(hex.EncodeToString(header.ParentHash[:]))

	// Create block metadata
	metadata := make(map[string]interface{})
	metadata["state_root"] = hex.EncodeToString(header.StateRoot[:])
	metadata["extrinsics_root"] = hex.EncodeToString(header.ExtrinsicsRoot[:])
	metadata["number"] = uint64(header.Number)
	metadata["num_extrinsics"] = len(block.Extrinsics)

	// Add digest information
	if len(header.Digest) > 0 {
		digests := make([]map[string]interface{}, 0, len(header.Digest))
		for _, digestItem := range header.Digest {
			digestData := map[string]interface{}{
				"type": fmt.Sprintf("%T", digestItem),
			}
			digests = append(digests, digestData)
		}
		metadata["digests"] = digests
	}

	// Normalize extrinsics (transactions)
	transactions := make([]*models.Transaction, 0, len(block.Extrinsics))
	for i, ext := range block.Extrinsics {
		tx, err := n.NormalizeExtrinsic(&ext, uint64(header.Number), uint32(i))
		if err != nil {
			// Log error but continue
			continue
		}
		transactions = append(transactions, tx)
	}

	return &models.Block{
		ChainID:      n.chainID,
		Number:       uint64(header.Number),
		Hash:         blockHashHex,
		ParentHash:   parentHashHex,
		Timestamp:    models.NewTimestamp(0), // Will need to extract from extrinsics
		Transactions: transactions,
		Metadata:     metadata,
	}, nil
}

// NormalizeExtrinsic normalizes a Substrate extrinsic to domain Transaction
func (n *Normalizer) NormalizeExtrinsic(ext *types.Extrinsic, blockNumber uint64, txIndex uint32) (*models.Transaction, error) {
	if ext == nil {
		return nil, fmt.Errorf("extrinsic is nil")
	}

	// Calculate extrinsic hash (simplified - would need proper hashing)
	txHash := fmt.Sprintf("EXT-%d-%d", blockNumber, txIndex)

	// Default values
	var status models.TxStatus = models.TxStatusSuccess
	metadata := make(map[string]interface{})

	// Extract signature information
	var from string
	metadata["has_signature"] = ext.IsSigned()

	// Add extrinsic details to metadata
	metadata["version"] = ext.Version
	metadata["is_signed"] = ext.IsSigned()
	metadata["tx_index"] = txIndex

	if ext.IsSigned() {
		metadata["era"] = fmt.Sprintf("%v", ext.Signature.Era)
		metadata["signer_type"] = "MultiAddress"
	}

	// Extract method information
	metadata["call_index"] = fmt.Sprintf("%d-%d", ext.Method.CallIndex.SectionIndex, ext.Method.CallIndex.MethodIndex)

	return &models.Transaction{
		ChainID:     n.chainID,
		Hash:        txHash,
		BlockNumber: blockNumber,
		From:        from,
		To:          "", // Polkadot doesn't have a simple "to" address
		Value:       "0",
		GasUsed:     0,
		GasPrice:    "0",
		Status:      status,
		Timestamp:   models.NewTimestamp(0),
		Metadata:    metadata,
	}, nil
}

// NormalizeHeader normalizes a block header
func (n *Normalizer) NormalizeHeader(header *types.Header, blockHash types.Hash) (*models.Block, error) {
	if header == nil {
		return nil, fmt.Errorf("header is nil")
	}

	blockHashHex := strings.ToUpper(hex.EncodeToString(blockHash[:]))
	parentHashHex := strings.ToUpper(hex.EncodeToString(header.ParentHash[:]))

	metadata := make(map[string]interface{})
	metadata["state_root"] = hex.EncodeToString(header.StateRoot[:])
	metadata["extrinsics_root"] = hex.EncodeToString(header.ExtrinsicsRoot[:])
	metadata["number"] = uint64(header.Number)

	return &models.Block{
		ChainID:      n.chainID,
		Number:       uint64(header.Number),
		Hash:         blockHashHex,
		ParentHash:   parentHashHex,
		Timestamp:    models.NewTimestamp(0),
		Transactions: []*models.Transaction{},
		Metadata:     metadata,
	}, nil
}

// ExtractBlockTimestamp extracts timestamp from block extrinsics
func (n *Normalizer) ExtractBlockTimestamp(block *types.SignedBlock) (int64, error) {
	// In Substrate, timestamp is set via the Timestamp.set extrinsic
	// This is typically the first or second extrinsic in the block

	for _, ext := range block.Block.Extrinsics {
		// Check if this is the Timestamp.set call
		// CallIndex varies by runtime, but typically Timestamp is module 3
		if ext.Method.CallIndex.SectionIndex == 3 && ext.Method.CallIndex.MethodIndex == 0 {
			// Try to decode the timestamp from args
			// This is a simplified approach - actual implementation would need
			// to properly decode based on metadata
			if len(ext.Method.Args) > 0 {
				// Timestamp is encoded as Compact<Moment> (u64 in milliseconds)
				// For simplicity, return 0 here - proper implementation would decode
				return 0, nil
			}
		}
	}

	return 0, fmt.Errorf("timestamp not found in block")
}

// NormalizeAccountInfo normalizes account information
func (n *Normalizer) NormalizeAccountInfo(accountInfo *types.AccountInfo) map[string]interface{} {
	if accountInfo == nil {
		return nil
	}

	return map[string]interface{}{
		"nonce":       uint64(accountInfo.Nonce),
		"consumers":   uint32(accountInfo.Consumers),
		"providers":   uint32(accountInfo.Providers),
		"sufficients": uint32(accountInfo.Sufficients),
		"data": map[string]interface{}{
			"free":        accountInfo.Data.Free.String(),
			"reserved":    accountInfo.Data.Reserved.String(),
			"misc_frozen": accountInfo.Data.MiscFrozen.String(),
		},
	}
}

// NormalizeChainProperties normalizes chain properties
func (n *Normalizer) NormalizeChainProperties(props types.ChainProperties) map[string]interface{} {
	// ChainProperties is a map in the actual implementation
	// Just return it as-is for now
	return map[string]interface{}{
		"properties": props,
	}
}

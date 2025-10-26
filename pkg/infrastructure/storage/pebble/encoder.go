package pebble

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// Encoder handles encoding and decoding of data for storage
type Encoder struct{}

// NewEncoder creates a new encoder instance
func NewEncoder() *Encoder {
	return &Encoder{}
}

// EncodeBlock encodes a Block model to bytes
func (e *Encoder) EncodeBlock(block *models.Block) ([]byte, error) {
	if block == nil {
		return nil, fmt.Errorf("block cannot be nil")
	}

	data, err := json.Marshal(block)
	if err != nil {
		return nil, fmt.Errorf("failed to encode block: %w", err)
	}

	return data, nil
}

// DecodeBlock decodes bytes to a Block model
func (e *Encoder) DecodeBlock(data []byte) (*models.Block, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var block models.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("failed to decode block: %w", err)
	}

	return &block, nil
}

// EncodeTransaction encodes a Transaction model to bytes
func (e *Encoder) EncodeTransaction(tx *models.Transaction) ([]byte, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction cannot be nil")
	}

	data, err := json.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to encode transaction: %w", err)
	}

	return data, nil
}

// DecodeTransaction decodes bytes to a Transaction model
func (e *Encoder) DecodeTransaction(data []byte) (*models.Transaction, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var tx models.Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %w", err)
	}

	return &tx, nil
}

// EncodeChain encodes a Chain model to bytes
func (e *Encoder) EncodeChain(chain *models.Chain) ([]byte, error) {
	if chain == nil {
		return nil, fmt.Errorf("chain cannot be nil")
	}

	data, err := json.Marshal(chain)
	if err != nil {
		return nil, fmt.Errorf("failed to encode chain: %w", err)
	}

	return data, nil
}

// DecodeChain decodes bytes to a Chain model
func (e *Encoder) DecodeChain(data []byte) (*models.Chain, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var chain models.Chain
	if err := json.Unmarshal(data, &chain); err != nil {
		return nil, fmt.Errorf("failed to decode chain: %w", err)
	}

	return &chain, nil
}

// EncodeUint64 encodes a uint64 to bytes
func (e *Encoder) EncodeUint64(value uint64) []byte {
	return []byte(strconv.FormatUint(value, 10))
}

// DecodeUint64 decodes bytes to uint64
func (e *Encoder) DecodeUint64(data []byte) (uint64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("data cannot be empty")
	}

	value, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to decode uint64: %w", err)
	}

	return value, nil
}

// EncodeString encodes a string to bytes
func (e *Encoder) EncodeString(value string) []byte {
	return []byte(value)
}

// DecodeString decodes bytes to string
func (e *Encoder) DecodeString(data []byte) string {
	return string(data)
}

// EncodeBlockSummary encodes a BlockSummary to bytes
func (e *Encoder) EncodeBlockSummary(summary *models.BlockSummary) ([]byte, error) {
	if summary == nil {
		return nil, fmt.Errorf("block summary cannot be nil")
	}

	data, err := json.Marshal(summary)
	if err != nil {
		return nil, fmt.Errorf("failed to encode block summary: %w", err)
	}

	return data, nil
}

// DecodeBlockSummary decodes bytes to a BlockSummary
func (e *Encoder) DecodeBlockSummary(data []byte) (*models.BlockSummary, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var summary models.BlockSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("failed to decode block summary: %w", err)
	}

	return &summary, nil
}

// EncodeTransactionSummary encodes a TransactionSummary to bytes
func (e *Encoder) EncodeTransactionSummary(summary *models.TransactionSummary) ([]byte, error) {
	if summary == nil {
		return nil, fmt.Errorf("transaction summary cannot be nil")
	}

	data, err := json.Marshal(summary)
	if err != nil {
		return nil, fmt.Errorf("failed to encode transaction summary: %w", err)
	}

	return data, nil
}

// DecodeTransactionSummary decodes bytes to a TransactionSummary
func (e *Encoder) DecodeTransactionSummary(data []byte) (*models.TransactionSummary, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var summary models.TransactionSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("failed to decode transaction summary: %w", err)
	}

	return &summary, nil
}

// EncodeChainStats encodes a ChainStats to bytes
func (e *Encoder) EncodeChainStats(stats *models.ChainStats) ([]byte, error) {
	if stats == nil {
		return nil, fmt.Errorf("chain stats cannot be nil")
	}

	data, err := json.Marshal(stats)
	if err != nil {
		return nil, fmt.Errorf("failed to encode chain stats: %w", err)
	}

	return data, nil
}

// DecodeChainStats decodes bytes to ChainStats
func (e *Encoder) DecodeChainStats(data []byte) (*models.ChainStats, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var stats models.ChainStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("failed to decode chain stats: %w", err)
	}

	return &stats, nil
}

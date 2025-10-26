package models

import "errors"

var (
	// Block errors
	ErrInvalidBlockHash  = errors.New("invalid block hash")
	ErrInvalidParentHash = errors.New("invalid parent hash")
	ErrBlockNotFound     = errors.New("block not found")

	// Transaction errors
	ErrInvalidTxHash       = errors.New("invalid transaction hash")
	ErrInvalidFromAddress  = errors.New("invalid from address")
	ErrInvalidToAddress    = errors.New("invalid to address")
	ErrTransactionNotFound = errors.New("transaction not found")

	// Chain errors
	ErrInvalidChainType = errors.New("invalid chain type")
	ErrInvalidChainID   = errors.New("invalid chain ID")
	ErrChainNotFound    = errors.New("chain not found")

	// Common validation errors
	ErrInvalidTimestamp = errors.New("invalid timestamp")
	ErrInvalidLimit     = errors.New("invalid limit")
	ErrLimitTooLarge    = errors.New("limit too large")
	ErrInvalidOffset    = errors.New("invalid offset")

	// Data errors
	ErrInvalidData     = errors.New("invalid data")
	ErrDataCorrupted   = errors.New("data corrupted")
	ErrEncodingFailed  = errors.New("encoding failed")
	ErrDecodingFailed  = errors.New("decoding failed")
)

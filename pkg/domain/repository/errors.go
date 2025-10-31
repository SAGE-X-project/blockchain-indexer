package repository

import "errors"

var (
	// Storage errors
	ErrStorageClosed       = errors.New("storage is closed")
	ErrInvalidStorageType  = errors.New("invalid storage type")
	ErrStorageNotFound     = errors.New("storage not found")
	ErrStorageInitFailed   = errors.New("storage initialization failed")

	// Data errors
	ErrNotFound            = errors.New("not found")
	ErrAlreadyExists       = errors.New("already exists")
	ErrInvalidData         = errors.New("invalid data")
	ErrDataCorrupted       = errors.New("data corrupted")

	// Entity-specific errors
	ErrBlockNotFound       = errors.New("block not found")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrChainNotFound       = errors.New("chain not found")

	// Batch errors
	ErrBatchTooLarge       = errors.New("batch too large")
	ErrBatchCommitFailed   = errors.New("batch commit failed")
	ErrBatchClosed         = errors.New("batch is closed")

	// Query errors
	ErrInvalidFilter       = errors.New("invalid filter")
	ErrInvalidPagination   = errors.New("invalid pagination")
	ErrQueryFailed         = errors.New("query failed")
)

package service

import "errors"

var (
	// Adapter errors
	ErrInvalidRPCEndpoint  = errors.New("invalid RPC endpoint")
	ErrAdapterNotFound     = errors.New("adapter not found")
	ErrAdapterNotHealthy   = errors.New("adapter not healthy")
	ErrConnectionFailed    = errors.New("connection failed")
	ErrSubscriptionFailed  = errors.New("subscription failed")

	// Indexer errors
	ErrIndexerNotRunning   = errors.New("indexer not running")
	ErrIndexerAlreadyRunning = errors.New("indexer already running")
	ErrIndexingFailed      = errors.New("indexing failed")

	// Provider errors
	ErrProviderNotRunning  = errors.New("provider not running")
	ErrProviderStartFailed = errors.New("provider start failed")
	ErrProviderStopFailed  = errors.New("provider stop failed")
	ErrInvalidProviderType = errors.New("invalid provider type")
	ErrInvalidPort         = errors.New("invalid port")
	ErrInvalidTLSConfig    = errors.New("invalid TLS configuration")

	// Operation errors
	ErrOperationTimeout    = errors.New("operation timeout")
	ErrOperationCancelled  = errors.New("operation cancelled")
	ErrMaxRetriesExceeded  = errors.New("max retries exceeded")
)

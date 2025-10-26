package evm

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// Client wraps Ethereum RPC client with connection management and retry logic
type Client struct {
	config     *Config
	clients    []*ethclient.Client
	rpcClients []*rpc.Client
	current    uint32 // atomic counter for round-robin
	mu         sync.RWMutex
	connected  bool
	lastError  error
	errorCount uint64

	// Metrics
	requestCount  uint64
	successCount  uint64
	failureCount  uint64
	totalLatency  int64
	lastChecked   int64
}

// NewClient creates a new EVM RPC client
func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client := &Client{
		config:    config,
		connected: false,
	}

	// Initialize RPC clients
	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}

// connect establishes connections to all RPC endpoints
func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	endpoints := c.config.RPCEndpoints
	if len(endpoints) == 0 {
		return fmt.Errorf("no RPC endpoints configured")
	}

	clients := make([]*ethclient.Client, 0, len(endpoints))
	rpcClients := make([]*rpc.Client, 0, len(endpoints))

	for _, endpoint := range endpoints {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.ConnectionTimeout)
		defer cancel()

		// Create RPC client
		rpcClient, err := rpc.DialContext(ctx, endpoint)
		if err != nil {
			continue
		}

		// Create eth client
		ethClient := ethclient.NewClient(rpcClient)

		// Verify connection
		if _, err := ethClient.ChainID(ctx); err != nil {
			rpcClient.Close()
			continue
		}

		clients = append(clients, ethClient)
		rpcClients = append(rpcClients, rpcClient)
	}

	if len(clients) == 0 {
		return fmt.Errorf("failed to connect to any RPC endpoint")
	}

	c.clients = clients
	c.rpcClients = rpcClients
	c.connected = true
	c.errorCount = 0
	c.lastError = nil

	return nil
}

// getClient returns a client using round-robin selection
func (c *Client) getClient() *ethclient.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.clients) == 0 {
		return nil
	}

	idx := atomic.AddUint32(&c.current, 1) % uint32(len(c.clients))
	return c.clients[idx]
}

// executeWithRetry executes a function with retry logic
func (c *Client) executeWithRetry(ctx context.Context, fn func(context.Context, *ethclient.Client) error) error {
	atomic.AddUint64(&c.requestCount, 1)
	start := time.Now()

	var lastErr error
	delay := c.config.RetryDelay

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				delay = time.Duration(float64(delay) * c.config.RetryBackoff)
				if delay > c.config.MaxRetryDelay {
					delay = c.config.MaxRetryDelay
				}
			}
		}

		client := c.getClient()
		if client == nil {
			return fmt.Errorf("no available clients")
		}

		reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
		err := fn(reqCtx, client)
		cancel()

		if err == nil {
			atomic.AddUint64(&c.successCount, 1)
			atomic.AddInt64(&c.totalLatency, time.Since(start).Milliseconds())
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !c.isRetryableError(err) {
			break
		}
	}

	atomic.AddUint64(&c.failureCount, 1)
	atomic.AddUint64(&c.errorCount, 1)
	c.lastError = lastErr

	return lastErr
}

// isRetryableError checks if an error is retryable
func (c *Client) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Network errors are retryable
	errMsg := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"too many requests",
		"rate limit",
		"connection reset",
		"EOF",
	}

	for _, pattern := range retryablePatterns {
		if contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// ChainID returns the chain ID
func (c *Client) ChainID(ctx context.Context) (*big.Int, error) {
	var chainID *big.Int
	err := c.executeWithRetry(ctx, func(ctx context.Context, client *ethclient.Client) error {
		var err error
		chainID, err = client.ChainID(ctx)
		return err
	})
	return chainID, err
}

// BlockNumber returns the latest block number
func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	var blockNumber uint64
	err := c.executeWithRetry(ctx, func(ctx context.Context, client *ethclient.Client) error {
		var err error
		blockNumber, err = client.BlockNumber(ctx)
		return err
	})
	return blockNumber, err
}

// BlockByNumber returns a block by number
func (c *Client) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	var block *types.Block
	err := c.executeWithRetry(ctx, func(ctx context.Context, client *ethclient.Client) error {
		var err error
		block, err = client.BlockByNumber(ctx, number)
		return err
	})
	return block, err
}

// BlockByHash returns a block by hash
func (c *Client) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	var block *types.Block
	err := c.executeWithRetry(ctx, func(ctx context.Context, client *ethclient.Client) error {
		var err error
		block, err = client.BlockByHash(ctx, hash)
		return err
	})
	return block, err
}

// TransactionByHash returns a transaction by hash
func (c *Client) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	var tx *types.Transaction
	var isPending bool
	err := c.executeWithRetry(ctx, func(ctx context.Context, client *ethclient.Client) error {
		var err error
		tx, isPending, err = client.TransactionByHash(ctx, hash)
		return err
	})
	return tx, isPending, err
}

// TransactionReceipt returns a transaction receipt
func (c *Client) TransactionReceipt(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	var receipt *types.Receipt
	err := c.executeWithRetry(ctx, func(ctx context.Context, client *ethclient.Client) error {
		var err error
		receipt, err = client.TransactionReceipt(ctx, hash)
		return err
	})
	return receipt, err
}

// HeaderByNumber returns a block header by number
func (c *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	var header *types.Header
	err := c.executeWithRetry(ctx, func(ctx context.Context, client *ethclient.Client) error {
		var err error
		header, err = client.HeaderByNumber(ctx, number)
		return err
	})
	return header, err
}

// FilterLogs executes a filter query
func (c *Client) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	var logs []types.Log
	err := c.executeWithRetry(ctx, func(ctx context.Context, client *ethclient.Client) error {
		var err error
		logs, err = client.FilterLogs(ctx, query)
		return err
	})
	return logs, err
}

// SubscribeNewHead subscribes to new block headers
func (c *Client) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	client := c.getClient()
	if client == nil {
		return nil, fmt.Errorf("no available clients")
	}
	return client.SubscribeNewHead(ctx, ch)
}

// SyncProgress retrieves the current synchronization progress
func (c *Client) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	var progress *ethereum.SyncProgress
	err := c.executeWithRetry(ctx, func(ctx context.Context, client *ethclient.Client) error {
		var err error
		progress, err = client.SyncProgress(ctx)
		return err
	})
	return progress, err
}

// PeerCount returns the number of connected peers
func (c *Client) PeerCount(ctx context.Context) (uint64, error) {
	var count uint64
	err := c.executeWithRetry(ctx, func(ctx context.Context, client *ethclient.Client) error {
		// Call RPC method directly
		rpcClient := c.rpcClients[atomic.LoadUint32(&c.current)%uint32(len(c.rpcClients))]
		var hexCount string
		err := rpcClient.CallContext(ctx, &hexCount, "net_peerCount")
		if err != nil {
			return err
		}
		count = hexToUint64(hexCount)
		return nil
	})
	return count, err
}

// HealthStatus returns the health status of the client
func (c *Client) HealthStatus(ctx context.Context) (*HealthStatus, error) {
	start := time.Now()
	status := &HealthStatus{
		Connected:    c.connected,
		ErrorCount:   atomic.LoadUint64(&c.errorCount),
		LastError:    "",
		LastChecked:  time.Now().Unix(),
		ResponseTime: 0,
	}

	if c.lastError != nil {
		status.LastError = c.lastError.Error()
	}

	// Try to fetch latest block
	blockNumber, err := c.BlockNumber(ctx)
	if err != nil {
		status.Connected = false
		return status, err
	}

	status.BlockHeight = blockNumber
	status.ResponseTime = time.Since(start).Milliseconds()

	// Check sync progress
	syncProgress, err := c.SyncProgress(ctx)
	if err == nil && syncProgress != nil {
		status.Syncing = true
		status.SyncProgress = &SyncProgress{
			StartingBlock: syncProgress.StartingBlock,
			CurrentBlock:  syncProgress.CurrentBlock,
			HighestBlock:  syncProgress.HighestBlock,
			PulledStates:  syncProgress.PulledStates,
			KnownStates:   syncProgress.KnownStates,
		}
	}

	// Get peer count
	peerCount, err := c.PeerCount(ctx)
	if err == nil {
		status.PeerCount = peerCount
	}

	return status, nil
}

// Close closes all client connections
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, client := range c.rpcClients {
		client.Close()
	}

	c.clients = nil
	c.rpcClients = nil
	c.connected = false

	return nil
}

// GetMetrics returns client metrics
func (c *Client) GetMetrics() map[string]interface{} {
	requestCount := atomic.LoadUint64(&c.requestCount)
	successCount := atomic.LoadUint64(&c.successCount)
	failureCount := atomic.LoadUint64(&c.failureCount)
	totalLatency := atomic.LoadInt64(&c.totalLatency)

	avgLatency := int64(0)
	if successCount > 0 {
		avgLatency = totalLatency / int64(successCount)
	}

	return map[string]interface{}{
		"request_count":  requestCount,
		"success_count":  successCount,
		"failure_count":  failureCount,
		"avg_latency_ms": avgLatency,
		"error_count":    atomic.LoadUint64(&c.errorCount),
		"connected":      c.connected,
		"client_count":   len(c.clients),
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func hexToUint64(hex string) uint64 {
	if len(hex) >= 2 && hex[0] == '0' && (hex[1] == 'x' || hex[1] == 'X') {
		hex = hex[2:]
	}
	var result uint64
	for i := 0; i < len(hex); i++ {
		result *= 16
		c := hex[i]
		if c >= '0' && c <= '9' {
			result += uint64(c - '0')
		} else if c >= 'a' && c <= 'f' {
			result += uint64(c - 'a' + 10)
		} else if c >= 'A' && c <= 'F' {
			result += uint64(c - 'A' + 10)
		}
	}
	return result
}

package solana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// Client wraps the Solana JSON-RPC API
type Client struct {
	config      *Config
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	requestID   atomic.Int64
	errorCount  atomic.Int32
	mu          sync.RWMutex
	connected   bool
}

// NewClient creates a new Solana RPC client
func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
		rateLimiter: rate.NewLimiter(rate.Limit(config.RateLimitPerSecond), config.RateLimitPerSecond),
		connected:   true,
	}

	return client, nil
}

// call makes a JSON-RPC call to the Solana node
func (c *Client) call(ctx context.Context, method string, params []interface{}, result interface{}) error {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter: %w", err)
	}

	// Create request
	reqID := int(c.requestID.Add(1))
	rpcReq := RPCRequest{
		JSONRPC: "2.0",
		ID:      reqID,
		Method:  method,
		Params:  params,
	}

	reqBody, err := json.Marshal(rpcReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.RPCEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request with retries
	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.config.RetryDelay):
			}
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			c.errorCount.Add(1)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = err
			c.errorCount.Add(1)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
			c.errorCount.Add(1)
			continue
		}

		// Parse response
		var rpcResp RPCResponse
		if err := json.Unmarshal(body, &rpcResp); err != nil {
			lastErr = fmt.Errorf("failed to unmarshal response: %w", err)
			c.errorCount.Add(1)
			continue
		}

		// Check for RPC error
		if rpcResp.Error != nil {
			return fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
		}

		// Unmarshal result
		if result != nil && len(rpcResp.Result) > 0 {
			if err := json.Unmarshal(rpcResp.Result, result); err != nil {
				return fmt.Errorf("failed to unmarshal result: %w", err)
			}
		}

		// Reset error count on success
		c.errorCount.Store(0)
		return nil
	}

	return fmt.Errorf("request failed after %d retries: %w", c.config.MaxRetries, lastErr)
}

// GetSlot returns the current slot
func (c *Client) GetSlot(ctx context.Context) (uint64, error) {
	var result GetSlotResponse
	params := []interface{}{
		map[string]string{"commitment": CommitmentFinalized},
	}

	if err := c.call(ctx, "getSlot", params, &result); err != nil {
		return 0, fmt.Errorf("getSlot: %w", err)
	}

	return uint64(result), nil
}

// GetBlockHeight returns the current block height
func (c *Client) GetBlockHeight(ctx context.Context) (uint64, error) {
	var result GetBlockHeightResponse
	params := []interface{}{
		map[string]string{"commitment": CommitmentFinalized},
	}

	if err := c.call(ctx, "getBlockHeight", params, &result); err != nil {
		return 0, fmt.Errorf("getBlockHeight: %w", err)
	}

	return uint64(result), nil
}

// GetBlock returns a block at the specified slot
func (c *Client) GetBlock(ctx context.Context, slot uint64) (*GetBlockResponse, error) {
	var result GetBlockResponse

	// Build params based on configuration
	config := map[string]interface{}{
		"commitment":                     CommitmentFinalized,
		"maxSupportedTransactionVersion": 0,
		"rewards":                        c.config.EnableRewards,
	}

	// Set transaction detail level
	switch c.config.TransactionDetails {
	case "full":
		config["encoding"] = EncodingJSON
		config["transactionDetails"] = "full"
	case "accounts":
		config["encoding"] = EncodingJSON
		config["transactionDetails"] = "accounts"
	case "signatures":
		config["transactionDetails"] = "signatures"
	case "none":
		config["transactionDetails"] = "none"
	}

	params := []interface{}{slot, config}

	if err := c.call(ctx, "getBlock", params, &result); err != nil {
		return nil, fmt.Errorf("getBlock: %w", err)
	}

	return &result, nil
}

// GetTransaction returns a transaction by signature
func (c *Client) GetTransaction(ctx context.Context, signature string) (*GetTransactionResponse, error) {
	var result GetTransactionResponse

	config := map[string]interface{}{
		"commitment":                     CommitmentFinalized,
		"encoding":                       EncodingJSON,
		"maxSupportedTransactionVersion": 0,
	}

	params := []interface{}{signature, config}

	if err := c.call(ctx, "getTransaction", params, &result); err != nil {
		return nil, fmt.Errorf("getTransaction: %w", err)
	}

	return &result, nil
}

// GetHealth returns the health status
func (c *Client) GetHealth(ctx context.Context) (string, error) {
	var result GetHealthResponse
	if err := c.call(ctx, "getHealth", nil, &result); err != nil {
		return "", fmt.Errorf("getHealth: %w", err)
	}

	return string(result), nil
}

// GetVersion returns the version
func (c *Client) GetVersion(ctx context.Context) (*GetVersionResponse, error) {
	var result GetVersionResponse
	if err := c.call(ctx, "getVersion", nil, &result); err != nil {
		return nil, fmt.Errorf("getVersion: %w", err)
	}

	return &result, nil
}

// GetBlocksInRange returns blocks in the specified range
func (c *Client) GetBlocksInRange(ctx context.Context, startSlot, endSlot uint64) ([]uint64, error) {
	var result []uint64

	config := map[string]string{
		"commitment": CommitmentFinalized,
	}

	params := []interface{}{startSlot, endSlot, config}

	if err := c.call(ctx, "getBlocks", params, &result); err != nil {
		return nil, fmt.Errorf("getBlocks: %w", err)
	}

	return result, nil
}

// HealthStatus returns the current health status
type HealthStatus struct {
	Connected  bool
	ErrorCount int32
}

// GetHealthStatus returns the current health status
func (c *Client) GetHealthStatus(ctx context.Context) (*HealthStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &HealthStatus{
		Connected:  c.connected,
		ErrorCount: c.errorCount.Load(),
	}, nil
}

// Close closes the client
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = false
	c.httpClient.CloseIdleConnections()

	return nil
}

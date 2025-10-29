package solana

import (
	"errors"
	"time"
)

// Config holds the configuration for the Solana adapter
type Config struct {
	// Chain identification
	ChainID   string `json:"chain_id"`
	ChainName string `json:"chain_name"`
	Network   string `json:"network"` // mainnet-beta, testnet, devnet

	// RPC configuration
	RPCEndpoint string `json:"rpc_endpoint"`
	WSEndpoint  string `json:"ws_endpoint,omitempty"`

	// Request configuration
	RequestTimeout      time.Duration `json:"request_timeout"`       // Timeout for RPC requests
	MaxRetries          int           `json:"max_retries"`           // Max number of retries
	RetryDelay          time.Duration `json:"retry_delay"`           // Delay between retries
	MaxConcurrentReqs   int           `json:"max_concurrent_reqs"`   // Max concurrent requests
	RateLimitPerSecond  int           `json:"rate_limit_per_second"` // Rate limit (requests/sec)

	// Fetching configuration
	ConcurrentFetches  int    `json:"concurrent_fetches"`  // Number of concurrent block fetches
	MaxBlockRange      uint64 `json:"max_block_range"`     // Max block range per request
	EnableWebSocket    bool   `json:"enable_websocket"`    // Enable WebSocket support
	EnableVotes        bool   `json:"enable_votes"`        // Include vote transactions
	EnableRewards      bool   `json:"enable_rewards"`      // Include block rewards
	TransactionDetails string `json:"transaction_details"` // Transaction detail level: full, accounts, signatures, none

	// Subscription configuration
	SubscriptionBufferSize int `json:"subscription_buffer_size"` // Buffer size for subscriptions

	// Error handling
	MaxErrorCount int `json:"max_error_count"` // Max errors before marking unhealthy
}

// DefaultConfig returns a default configuration for Solana
func DefaultConfig(chainID, network, rpcEndpoint string) *Config {
	return &Config{
		ChainID:                chainID,
		ChainName:              "Solana",
		Network:                network,
		RPCEndpoint:            rpcEndpoint,
		RequestTimeout:         30 * time.Second,
		MaxRetries:             3,
		RetryDelay:             2 * time.Second,
		MaxConcurrentReqs:      10,
		RateLimitPerSecond:     100,
		ConcurrentFetches:      5,
		MaxBlockRange:          100,
		EnableWebSocket:        false,
		EnableVotes:            false,
		EnableRewards:          true,
		TransactionDetails:     "full",
		SubscriptionBufferSize: 100,
		MaxErrorCount:          10,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ChainID == "" {
		return errors.New("chain_id is required")
	}
	if c.ChainName == "" {
		return errors.New("chain_name is required")
	}
	if c.Network == "" {
		return errors.New("network is required")
	}
	if c.RPCEndpoint == "" {
		return errors.New("rpc_endpoint is required")
	}
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 30 * time.Second
	}
	if c.MaxRetries < 0 {
		c.MaxRetries = 3
	}
	if c.ConcurrentFetches <= 0 {
		c.ConcurrentFetches = 5
	}
	if c.MaxBlockRange <= 0 {
		c.MaxBlockRange = 100
	}
	if c.MaxConcurrentReqs <= 0 {
		c.MaxConcurrentReqs = 10
	}
	if c.RateLimitPerSecond <= 0 {
		c.RateLimitPerSecond = 100
	}
	if c.SubscriptionBufferSize <= 0 {
		c.SubscriptionBufferSize = 100
	}
	if c.MaxErrorCount <= 0 {
		c.MaxErrorCount = 10
	}

	// Validate transaction detail level
	switch c.TransactionDetails {
	case "full", "accounts", "signatures", "none":
		// Valid
	default:
		return errors.New("invalid transaction_details: must be full, accounts, signatures, or none")
	}

	return nil
}

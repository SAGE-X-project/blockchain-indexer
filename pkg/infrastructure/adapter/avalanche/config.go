package avalanche

import (
	"fmt"
	"net/url"
	"time"
)

// ChainType represents the type of Avalanche chain
type ChainType string

const (
	// CChain is the Contract Chain (EVM compatible)
	CChain ChainType = "C-Chain"
	// XChain is the Exchange Chain
	XChain ChainType = "X-Chain"
	// PChain is the Platform Chain
	PChain ChainType = "P-Chain"
)

// Config represents the configuration for Avalanche adapter
type Config struct {
	// ChainID is the unique identifier for the chain
	ChainID string `yaml:"chain_id" json:"chain_id"`

	// ChainName is the human-readable name of the chain
	ChainName string `yaml:"chain_name" json:"chain_name"`

	// Network is the network type (mainnet, testnet, devnet)
	Network string `yaml:"network" json:"network"`

	// ChainType specifies which Avalanche chain (C-Chain, X-Chain, P-Chain)
	ChainType ChainType `yaml:"chain_type" json:"chain_type"`

	// RPCURL is the RPC endpoint URL
	RPCURL string `yaml:"rpc_url" json:"rpc_url"`

	// WSURL is the WebSocket endpoint URL (optional)
	WSURL string `yaml:"ws_url" json:"ws_url"`

	// Timeout is the request timeout duration
	Timeout time.Duration `yaml:"timeout" json:"timeout"`

	// RetryAttempts is the number of retry attempts for failed requests
	RetryAttempts int `yaml:"retry_attempts" json:"retry_attempts"`

	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration `yaml:"retry_delay" json:"retry_delay"`

	// MaxConnections is the maximum number of connections
	MaxConnections int `yaml:"max_connections" json:"max_connections"`

	// BatchSize is the number of blocks to fetch in a single batch
	BatchSize int `yaml:"batch_size" json:"batch_size"`

	// StartBlock is the block number to start indexing from
	StartBlock uint64 `yaml:"start_block" json:"start_block"`

	// EnableWebSocket enables WebSocket subscription support
	EnableWebSocket bool `yaml:"enable_websocket" json:"enable_websocket"`
}

// DefaultCChainConfig returns a default configuration for Avalanche C-Chain
func DefaultCChainConfig() *Config {
	return &Config{
		ChainID:         "avalanche-c",
		ChainName:       "Avalanche C-Chain",
		Network:         "mainnet",
		ChainType:       CChain,
		RPCURL:          "https://api.avax.network/ext/bc/C/rpc",
		WSURL:           "wss://api.avax.network/ext/bc/C/ws",
		Timeout:         30 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      1 * time.Second,
		MaxConnections:  10,
		BatchSize:       100,
		StartBlock:      0,
		EnableWebSocket: false,
	}
}

// FujiCChainConfig returns a configuration for Avalanche Fuji testnet C-Chain
func FujiCChainConfig() *Config {
	return &Config{
		ChainID:         "avalanche-fuji-c",
		ChainName:       "Avalanche Fuji C-Chain",
		Network:         "testnet",
		ChainType:       CChain,
		RPCURL:          "https://api.avax-test.network/ext/bc/C/rpc",
		WSURL:           "wss://api.avax-test.network/ext/bc/C/ws",
		Timeout:         30 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      1 * time.Second,
		MaxConnections:  10,
		BatchSize:       100,
		StartBlock:      0,
		EnableWebSocket: false,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ChainID == "" {
		return fmt.Errorf("chain_id is required")
	}

	if c.ChainName == "" {
		return fmt.Errorf("chain_name is required")
	}

	if c.Network == "" {
		return fmt.Errorf("network is required")
	}

	if c.ChainType == "" {
		return fmt.Errorf("chain_type is required")
	}

	if c.ChainType != CChain && c.ChainType != XChain && c.ChainType != PChain {
		return fmt.Errorf("invalid chain_type: must be one of C-Chain, X-Chain, P-Chain")
	}

	if c.RPCURL == "" {
		return fmt.Errorf("rpc_url is required")
	}

	// Validate RPC URL
	if _, err := url.Parse(c.RPCURL); err != nil {
		return fmt.Errorf("invalid rpc_url: %w", err)
	}

	// Validate WebSocket URL if WebSocket is enabled
	if c.EnableWebSocket {
		if c.WSURL == "" {
			return fmt.Errorf("ws_url is required when enable_websocket is true")
		}
		if _, err := url.Parse(c.WSURL); err != nil {
			return fmt.Errorf("invalid ws_url: %w", err)
		}
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if c.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts must be non-negative")
	}

	if c.RetryDelay < 0 {
		return fmt.Errorf("retry_delay must be non-negative")
	}

	if c.MaxConnections <= 0 {
		return fmt.Errorf("max_connections must be positive")
	}

	if c.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be positive")
	}

	return nil
}

// GetRPCURL returns the RPC URL
func (c *Config) GetRPCURL() string {
	return c.RPCURL
}

// GetWSURL returns the WebSocket URL
func (c *Config) GetWSURL() string {
	return c.WSURL
}

// IsWebSocketEnabled returns whether WebSocket is enabled
func (c *Config) IsWebSocketEnabled() bool {
	return c.EnableWebSocket
}

// IsCChain returns whether this is a C-Chain configuration
func (c *Config) IsCChain() bool {
	return c.ChainType == CChain
}

// IsXChain returns whether this is an X-Chain configuration
func (c *Config) IsXChain() bool {
	return c.ChainType == XChain
}

// IsPChain returns whether this is a P-Chain configuration
func (c *Config) IsPChain() bool {
	return c.ChainType == PChain
}

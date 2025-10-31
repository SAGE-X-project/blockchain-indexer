package ripple

import (
	"fmt"
	"net/url"
	"time"
)

// Config represents the configuration for Ripple (XRP Ledger) adapter
type Config struct {
	// ChainID is the unique identifier for the chain
	ChainID string `yaml:"chain_id" json:"chain_id"`

	// ChainName is the human-readable name of the chain
	ChainName string `yaml:"chain_name" json:"chain_name"`

	// Network is the network type (mainnet, testnet, devnet)
	Network string `yaml:"network" json:"network"`

	// RPCURL is the rippled RPC endpoint URL
	RPCURL string `yaml:"rpc_url" json:"rpc_url"`

	// WebSocketURL is the WebSocket endpoint URL
	WebSocketURL string `yaml:"websocket_url" json:"websocket_url"`

	// Timeout is the request timeout duration
	Timeout time.Duration `yaml:"timeout" json:"timeout"`

	// RetryAttempts is the number of retry attempts for failed requests
	RetryAttempts int `yaml:"retry_attempts" json:"retry_attempts"`

	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration `yaml:"retry_delay" json:"retry_delay"`

	// MaxConnections is the maximum number of connections
	MaxConnections int `yaml:"max_connections" json:"max_connections"`

	// BatchSize is the number of ledgers to fetch in a single batch
	BatchSize int `yaml:"batch_size" json:"batch_size"`

	// StartLedger is the ledger index to start indexing from
	StartLedger uint64 `yaml:"start_ledger" json:"start_ledger"`

	// EnableWebSocket enables WebSocket subscription support
	EnableWebSocket bool `yaml:"enable_websocket" json:"enable_websocket"`

	// IncludeTransactionMetadata determines whether to include transaction metadata
	IncludeTransactionMetadata bool `yaml:"include_transaction_metadata" json:"include_transaction_metadata"`
}

// DefaultConfig returns a default configuration for XRP Ledger mainnet
func DefaultConfig() *Config {
	return &Config{
		ChainID:                    "xrpl-mainnet",
		ChainName:                  "XRP Ledger",
		Network:                    "mainnet",
		RPCURL:                     "https://xrplcluster.com",
		WebSocketURL:               "wss://xrplcluster.com",
		Timeout:                    30 * time.Second,
		RetryAttempts:              3,
		RetryDelay:                 1 * time.Second,
		MaxConnections:             10,
		BatchSize:                  100,
		StartLedger:                0,
		EnableWebSocket:            false,
		IncludeTransactionMetadata: true,
	}
}

// TestnetConfig returns a configuration for XRP Ledger testnet
func TestnetConfig() *Config {
	return &Config{
		ChainID:                    "xrpl-testnet",
		ChainName:                  "XRP Ledger Testnet",
		Network:                    "testnet",
		RPCURL:                     "https://s.altnet.rippletest.net:51234",
		WebSocketURL:               "wss://s.altnet.rippletest.net:51233",
		Timeout:                    30 * time.Second,
		RetryAttempts:              3,
		RetryDelay:                 1 * time.Second,
		MaxConnections:             10,
		BatchSize:                  100,
		StartLedger:                0,
		EnableWebSocket:            false,
		IncludeTransactionMetadata: true,
	}
}

// DevnetConfig returns a configuration for XRP Ledger devnet
func DevnetConfig() *Config {
	return &Config{
		ChainID:                    "xrpl-devnet",
		ChainName:                  "XRP Ledger Devnet",
		Network:                    "devnet",
		RPCURL:                     "https://s.devnet.rippletest.net:51234",
		WebSocketURL:               "wss://s.devnet.rippletest.net:51233",
		Timeout:                    30 * time.Second,
		RetryAttempts:              3,
		RetryDelay:                 1 * time.Second,
		MaxConnections:             10,
		BatchSize:                  100,
		StartLedger:                0,
		EnableWebSocket:            false,
		IncludeTransactionMetadata: true,
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

	if c.RPCURL == "" {
		return fmt.Errorf("rpc_url is required")
	}

	// Validate RPC URL
	if _, err := url.Parse(c.RPCURL); err != nil {
		return fmt.Errorf("invalid rpc_url: %w", err)
	}

	// Validate WebSocket URL if WebSocket is enabled
	if c.EnableWebSocket {
		if c.WebSocketURL == "" {
			return fmt.Errorf("websocket_url is required when enable_websocket is true")
		}
		if _, err := url.Parse(c.WebSocketURL); err != nil {
			return fmt.Errorf("invalid websocket_url: %w", err)
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

// GetWebSocketURL returns the WebSocket URL
func (c *Config) GetWebSocketURL() string {
	return c.WebSocketURL
}

// IsWebSocketEnabled returns whether WebSocket is enabled
func (c *Config) IsWebSocketEnabled() bool {
	return c.EnableWebSocket
}

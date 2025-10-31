package polkadot

import (
	"fmt"
	"net/url"
	"time"
)

// Config represents the configuration for Polkadot adapter
type Config struct {
	// ChainID is the unique identifier for the chain
	ChainID string `yaml:"chain_id" json:"chain_id"`

	// ChainName is the human-readable name of the chain
	ChainName string `yaml:"chain_name" json:"chain_name"`

	// Network is the network type (mainnet, testnet, devnet)
	Network string `yaml:"network" json:"network"`

	// RPCURL is the Substrate RPC endpoint URL
	RPCURL string `yaml:"rpc_url" json:"rpc_url"`

	// WSURL is the WebSocket endpoint URL
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

	// SS58Format is the address format (e.g., 0 for Polkadot, 2 for Kusama)
	SS58Format uint16 `yaml:"ss58_format" json:"ss58_format"`

	// TokenSymbol is the token symbol (e.g., "DOT", "KSM")
	TokenSymbol string `yaml:"token_symbol" json:"token_symbol"`

	// TokenDecimals is the number of decimal places for the token
	TokenDecimals uint8 `yaml:"token_decimals" json:"token_decimals"`

	// IncludeExtrinsics determines whether to include extrinsics
	IncludeExtrinsics bool `yaml:"include_extrinsics" json:"include_extrinsics"`

	// IncludeEvents determines whether to include events
	IncludeEvents bool `yaml:"include_events" json:"include_events"`

	// IncludeLogs determines whether to include logs
	IncludeLogs bool `yaml:"include_logs" json:"include_logs"`
}

// DefaultConfig returns a default configuration for Polkadot
func DefaultConfig() *Config {
	return &Config{
		ChainID:            "polkadot",
		ChainName:          "Polkadot",
		Network:            "mainnet",
		RPCURL:             "https://rpc.polkadot.io",
		WSURL:              "wss://rpc.polkadot.io",
		Timeout:            30 * time.Second,
		RetryAttempts:      3,
		RetryDelay:         1 * time.Second,
		MaxConnections:     10,
		BatchSize:          100,
		StartBlock:         0,
		EnableWebSocket:    false,
		SS58Format:         0,
		TokenSymbol:        "DOT",
		TokenDecimals:      10,
		IncludeExtrinsics:  true,
		IncludeEvents:      true,
		IncludeLogs:        false,
	}
}

// KusamaConfig returns a default configuration for Kusama
func KusamaConfig() *Config {
	return &Config{
		ChainID:            "kusama",
		ChainName:          "Kusama",
		Network:            "mainnet",
		RPCURL:             "https://kusama-rpc.polkadot.io",
		WSURL:              "wss://kusama-rpc.polkadot.io",
		Timeout:            30 * time.Second,
		RetryAttempts:      3,
		RetryDelay:         1 * time.Second,
		MaxConnections:     10,
		BatchSize:          100,
		StartBlock:         0,
		EnableWebSocket:    false,
		SS58Format:         2,
		TokenSymbol:        "KSM",
		TokenDecimals:      12,
		IncludeExtrinsics:  true,
		IncludeEvents:      true,
		IncludeLogs:        false,
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

	if c.TokenSymbol == "" {
		return fmt.Errorf("token_symbol is required")
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

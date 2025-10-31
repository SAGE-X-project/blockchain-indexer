package cosmos

import (
	"fmt"
	"net/url"
	"time"
)

// Config represents the configuration for Cosmos adapter
type Config struct {
	// ChainID is the unique identifier for the chain
	ChainID string `yaml:"chain_id" json:"chain_id"`

	// ChainName is the human-readable name of the chain
	ChainName string `yaml:"chain_name" json:"chain_name"`

	// Network is the network type (mainnet, testnet, devnet)
	Network string `yaml:"network" json:"network"`

	// RPCURL is the Tendermint RPC endpoint URL
	RPCURL string `yaml:"rpc_url" json:"rpc_url"`

	// RESTURL is the Cosmos REST API endpoint URL (optional)
	RESTURL string `yaml:"rest_url" json:"rest_url"`

	// GRPCEndpoint is the gRPC endpoint URL (optional)
	GRPCEndpoint string `yaml:"grpc_endpoint" json:"grpc_endpoint"`

	// Timeout is the request timeout duration
	Timeout time.Duration `yaml:"timeout" json:"timeout"`

	// RetryAttempts is the number of retry attempts for failed requests
	RetryAttempts int `yaml:"retry_attempts" json:"retry_attempts"`

	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration `yaml:"retry_delay" json:"retry_delay"`

	// MaxConnections is the maximum number of connections to the RPC server
	MaxConnections int `yaml:"max_connections" json:"max_connections"`

	// BatchSize is the number of blocks to fetch in a single batch
	BatchSize int `yaml:"batch_size" json:"batch_size"`

	// StartBlock is the block number to start indexing from
	StartBlock uint64 `yaml:"start_block" json:"start_block"`

	// EnableWebSocket enables WebSocket subscription support
	EnableWebSocket bool `yaml:"enable_websocket" json:"enable_websocket"`

	// WebSocketURL is the WebSocket endpoint URL
	WebSocketURL string `yaml:"websocket_url" json:"websocket_url"`

	// Bech32Prefix is the address prefix (e.g., "cosmos", "osmo")
	Bech32Prefix string `yaml:"bech32_prefix" json:"bech32_prefix"`

	// CoinDenom is the coin denomination (e.g., "uatom", "uosmo")
	CoinDenom string `yaml:"coin_denom" json:"coin_denom"`

	// IncludeTxEvents determines whether to include transaction events
	IncludeTxEvents bool `yaml:"include_tx_events" json:"include_tx_events"`

	// IncludeBeginBlockEvents determines whether to include begin block events
	IncludeBeginBlockEvents bool `yaml:"include_begin_block_events" json:"include_begin_block_events"`

	// IncludeEndBlockEvents determines whether to include end block events
	IncludeEndBlockEvents bool `yaml:"include_end_block_events" json:"include_end_block_events"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		ChainID:                 "cosmoshub-4",
		ChainName:               "Cosmos Hub",
		Network:                 "mainnet",
		RPCURL:                  "https://rpc.cosmos.network:443",
		RESTURL:                 "https://api.cosmos.network",
		Timeout:                 30 * time.Second,
		RetryAttempts:           3,
		RetryDelay:              1 * time.Second,
		MaxConnections:          10,
		BatchSize:               100,
		StartBlock:              0,
		EnableWebSocket:         false,
		Bech32Prefix:            "cosmos",
		CoinDenom:               "uatom",
		IncludeTxEvents:         true,
		IncludeBeginBlockEvents: false,
		IncludeEndBlockEvents:   false,
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

	// Validate REST URL if provided
	if c.RESTURL != "" {
		if _, err := url.Parse(c.RESTURL); err != nil {
			return fmt.Errorf("invalid rest_url: %w", err)
		}
	}

	// Validate gRPC endpoint if provided
	if c.GRPCEndpoint != "" {
		if _, err := url.Parse(c.GRPCEndpoint); err != nil {
			return fmt.Errorf("invalid grpc_endpoint: %w", err)
		}
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

	if c.Bech32Prefix == "" {
		return fmt.Errorf("bech32_prefix is required")
	}

	if c.CoinDenom == "" {
		return fmt.Errorf("coin_denom is required")
	}

	return nil
}

// GetRPCURL returns the RPC URL
func (c *Config) GetRPCURL() string {
	return c.RPCURL
}

// GetRESTURL returns the REST API URL
func (c *Config) GetRESTURL() string {
	return c.RESTURL
}

// GetGRPCEndpoint returns the gRPC endpoint
func (c *Config) GetGRPCEndpoint() string {
	return c.GRPCEndpoint
}

// GetWebSocketURL returns the WebSocket URL
func (c *Config) GetWebSocketURL() string {
	if !c.EnableWebSocket {
		return ""
	}
	return c.WebSocketURL
}

// IsWebSocketEnabled returns whether WebSocket is enabled
func (c *Config) IsWebSocketEnabled() bool {
	return c.EnableWebSocket
}

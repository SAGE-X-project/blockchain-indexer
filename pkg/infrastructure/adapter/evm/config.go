package evm

import (
	"time"
)

// Config represents EVM adapter configuration
type Config struct {
	// Chain configuration
	ChainID   string
	ChainName string
	Network   string // mainnet, testnet, devnet

	// RPC endpoints
	RPCEndpoints []string
	WSEndpoints  []string

	// Connection settings
	MaxConnections    int
	ConnectionTimeout time.Duration
	RequestTimeout    time.Duration
	IdleTimeout       time.Duration

	// Retry settings
	MaxRetries    int
	RetryDelay    time.Duration
	RetryBackoff  float64 // exponential backoff multiplier
	MaxRetryDelay time.Duration

	// Rate limiting
	RequestsPerSecond int
	BurstSize         int

	// Block fetching
	BatchSize            int
	ConcurrentFetches    int
	BlockConfirmations   uint64
	BlockFetchDelay      time.Duration
	MaxBlockRange        uint64 // max blocks per batch request
	EnableReceiptFetch   bool
	EnableTraceAPI       bool

	// Subscription settings
	EnableWebSocket        bool
	WSReconnectDelay       time.Duration
	WSPingInterval         time.Duration
	WSMaxMessageSize       int64
	SubscriptionBufferSize int

	// Health check
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
	MaxErrorCount       uint64

	// Cache settings
	EnableCache      bool
	CacheTTL         time.Duration
	CacheMaxSize     int
	CacheBlockCount  int // number of recent blocks to cache
	CacheTxCount     int // number of recent transactions to cache

	// Performance tuning
	EnableBatchRequest bool
	BatchRequestSize   int
	EnableHTTP2        bool
	KeepAlive          bool

	// Security
	EnableTLS       bool
	TLSCertFile     string
	TLSKeyFile      string
	InsecureSkipTLS bool

	// Authentication
	AuthType     string // none, bearer, basic
	AuthToken    string
	AuthUsername string
	AuthPassword string

	// Metrics
	EnableMetrics    bool
	MetricsNamespace string
}

// DefaultConfig returns default EVM adapter configuration
func DefaultConfig() *Config {
	return &Config{
		// Connection settings
		MaxConnections:    10,
		ConnectionTimeout: 30 * time.Second,
		RequestTimeout:    10 * time.Second,
		IdleTimeout:       90 * time.Second,

		// Retry settings
		MaxRetries:    3,
		RetryDelay:    1 * time.Second,
		RetryBackoff:  2.0,
		MaxRetryDelay: 30 * time.Second,

		// Rate limiting
		RequestsPerSecond: 100,
		BurstSize:         10,

		// Block fetching
		BatchSize:            100,
		ConcurrentFetches:    10,
		BlockConfirmations:   12,
		BlockFetchDelay:      100 * time.Millisecond,
		MaxBlockRange:        1000,
		EnableReceiptFetch:   true,
		EnableTraceAPI:       false,

		// Subscription settings
		EnableWebSocket:        true,
		WSReconnectDelay:       5 * time.Second,
		WSPingInterval:         30 * time.Second,
		WSMaxMessageSize:       1024 * 1024 * 10, // 10MB
		SubscriptionBufferSize: 1000,

		// Health check
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		MaxErrorCount:       10,

		// Cache settings
		EnableCache:     true,
		CacheTTL:        5 * time.Minute,
		CacheMaxSize:    1000,
		CacheBlockCount: 100,
		CacheTxCount:    1000,

		// Performance tuning
		EnableBatchRequest: true,
		BatchRequestSize:   100,
		EnableHTTP2:        true,
		KeepAlive:          true,

		// Security
		EnableTLS:       false,
		InsecureSkipTLS: false,

		// Authentication
		AuthType: "none",

		// Metrics
		EnableMetrics:    true,
		MetricsNamespace: "evm_adapter",
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ChainID == "" {
		return &ConfigError{Field: "ChainID", Message: "chain ID is required"}
	}

	if len(c.RPCEndpoints) == 0 && len(c.WSEndpoints) == 0 {
		return &ConfigError{Field: "RPCEndpoints", Message: "at least one RPC or WebSocket endpoint is required"}
	}

	if c.MaxConnections <= 0 {
		return &ConfigError{Field: "MaxConnections", Message: "max connections must be positive"}
	}

	if c.BatchSize <= 0 {
		return &ConfigError{Field: "BatchSize", Message: "batch size must be positive"}
	}

	if c.ConcurrentFetches <= 0 {
		return &ConfigError{Field: "ConcurrentFetches", Message: "concurrent fetches must be positive"}
	}

	if c.MaxRetries < 0 {
		return &ConfigError{Field: "MaxRetries", Message: "max retries cannot be negative"}
	}

	if c.RetryBackoff <= 1.0 {
		return &ConfigError{Field: "RetryBackoff", Message: "retry backoff must be greater than 1.0"}
	}

	return nil
}

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return "config error [" + e.Field + "]: " + e.Message
}

// Clone creates a deep copy of the configuration
func (c *Config) Clone() *Config {
	clone := *c

	// Deep copy slices
	if len(c.RPCEndpoints) > 0 {
		clone.RPCEndpoints = make([]string, len(c.RPCEndpoints))
		copy(clone.RPCEndpoints, c.RPCEndpoints)
	}

	if len(c.WSEndpoints) > 0 {
		clone.WSEndpoints = make([]string, len(c.WSEndpoints))
		copy(clone.WSEndpoints, c.WSEndpoints)
	}

	return &clone
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// Config represents the complete application configuration
type Config struct {
	// Application settings
	App AppConfig `yaml:"app"`

	// Storage configuration
	Storage StorageConfig `yaml:"storage"`

	// Chain configurations
	Chains []ChainConfig `yaml:"chains"`

	// Server configurations
	Server ServerConfig `yaml:"server"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging"`

	// Metrics configuration
	Metrics MetricsConfig `yaml:"metrics"`
}

// AppConfig contains application-level settings
type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"` // development, staging, production
}

// StorageConfig contains storage settings
type StorageConfig struct {
	Type string `yaml:"type"` // pebble, postgres, etc.

	// PebbleDB specific settings
	Pebble PebbleConfig `yaml:"pebble,omitempty"`

	// PostgreSQL specific settings
	Postgres PostgresConfig `yaml:"postgres,omitempty"`
}

// PebbleConfig contains PebbleDB specific settings
type PebbleConfig struct {
	Path             string `yaml:"path"`
	CacheSize        int64  `yaml:"cache_size"`         // bytes
	MaxOpenFiles     int    `yaml:"max_open_files"`
	WriteBufferSize  int    `yaml:"write_buffer_size"`  // bytes
	MaxConcurrentMem int    `yaml:"max_concurrent_mem"`
	DisableWAL       bool   `yaml:"disable_wal"`
	BytesPerSync     int    `yaml:"bytes_per_sync"` // bytes
}

// PostgresConfig contains PostgreSQL specific settings
type PostgresConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Database        string `yaml:"database"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	SSLMode         string `yaml:"ssl_mode"`
	MaxConnections  int    `yaml:"max_connections"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime"`
}

// ChainConfig contains blockchain-specific configuration
type ChainConfig struct {
	ChainType          string   `yaml:"chain_type"`
	ChainID            string   `yaml:"chain_id"`
	Name               string   `yaml:"name"`
	Network            string   `yaml:"network"`
	Enabled            bool     `yaml:"enabled"`
	RPCEndpoints       []string `yaml:"rpc_endpoints"`
	WSEndpoints        []string `yaml:"ws_endpoints,omitempty"`
	StartBlock         uint64   `yaml:"start_block"`
	BatchSize          int      `yaml:"batch_size"`
	Workers            int      `yaml:"workers"`
	ConfirmationBlocks uint64   `yaml:"confirmation_blocks"`
	RetryAttempts      int      `yaml:"retry_attempts"`
	RetryDelay         string   `yaml:"retry_delay"`
}

// ServerConfig contains server configuration
type ServerConfig struct {
	// HTTP server
	HTTP HTTPConfig `yaml:"http"`

	// gRPC server
	GRPC GRPCConfig `yaml:"grpc"`

	// GraphQL server
	GraphQL GraphQLConfig `yaml:"graphql"`
}

// HTTPConfig contains HTTP server settings
type HTTPConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	TLSEnabled  bool   `yaml:"tls_enabled"`
	TLSCertFile string `yaml:"tls_cert_file,omitempty"`
	TLSKeyFile  string `yaml:"tls_key_file,omitempty"`
	ReadTimeout string `yaml:"read_timeout"`
	WriteTimeout string `yaml:"write_timeout"`
}

// GRPCConfig contains gRPC server settings
type GRPCConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	TLSEnabled  bool   `yaml:"tls_enabled"`
	TLSCertFile string `yaml:"tls_cert_file,omitempty"`
	TLSKeyFile  string `yaml:"tls_key_file,omitempty"`
}

// GraphQLConfig contains GraphQL server settings
type GraphQLConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Playground bool   `yaml:"playground"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `yaml:"level"`       // debug, info, warn, error
	Format     string `yaml:"format"`      // json, console
	Output     string `yaml:"output"`      // stdout, stderr, file
	FilePath   string `yaml:"file_path,omitempty"`
	MaxSize    int    `yaml:"max_size"`    // megabytes
	MaxBackups int    `yaml:"max_backups"` // number of backups
	MaxAge     int    `yaml:"max_age"`     // days
	Compress   bool   `yaml:"compress"`
}

// MetricsConfig contains metrics settings
type MetricsConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Path     string `yaml:"path"`
	Interval string `yaml:"interval"` // collection interval
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// LoadFromEnv loads configuration from environment variable
// CONFIG_PATH environment variable specifies the config file path
func LoadFromEnv() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config.yaml" // default
	}

	return Load(path)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate app config
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	// Validate storage config
	if c.Storage.Type == "" {
		return fmt.Errorf("storage.type is required")
	}

	switch c.Storage.Type {
	case "pebble":
		if c.Storage.Pebble.Path == "" {
			return fmt.Errorf("storage.pebble.path is required")
		}
	case "postgres":
		if c.Storage.Postgres.Host == "" {
			return fmt.Errorf("storage.postgres.host is required")
		}
		if c.Storage.Postgres.Database == "" {
			return fmt.Errorf("storage.postgres.database is required")
		}
	default:
		return fmt.Errorf("unsupported storage type: %s", c.Storage.Type)
	}

	// Validate chains
	if len(c.Chains) == 0 {
		return fmt.Errorf("at least one chain must be configured")
	}

	for i, chain := range c.Chains {
		if err := chain.Validate(); err != nil {
			return fmt.Errorf("chain[%d]: %w", i, err)
		}
	}

	// Validate logging
	if c.Logging.Level == "" {
		c.Logging.Level = "info" // default
	}

	return nil
}

// Validate validates chain configuration
func (c *ChainConfig) Validate() error {
	if c.ChainType == "" {
		return fmt.Errorf("chain_type is required")
	}

	chainType := models.ChainType(c.ChainType)
	if !chainType.IsValid() {
		return fmt.Errorf("invalid chain_type: %s", c.ChainType)
	}

	if c.ChainID == "" {
		return fmt.Errorf("chain_id is required")
	}

	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(c.RPCEndpoints) == 0 {
		return fmt.Errorf("at least one rpc_endpoint is required")
	}

	if c.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be positive")
	}

	if c.Workers <= 0 {
		return fmt.Errorf("workers must be positive")
	}

	return nil
}

// GetEnabledChains returns only enabled chains
func (c *Config) GetEnabledChains() []ChainConfig {
	enabled := make([]ChainConfig, 0)
	for _, chain := range c.Chains {
		if chain.Enabled {
			enabled = append(enabled, chain)
		}
	}
	return enabled
}

// GetChainByID returns chain configuration by ID
func (c *Config) GetChainByID(chainID string) (*ChainConfig, bool) {
	for _, chain := range c.Chains {
		if chain.ChainID == chainID {
			return &chain, true
		}
	}
	return nil, false
}

// GetRetryDelay parses retry delay duration
func (c *ChainConfig) GetRetryDelay() time.Duration {
	if c.RetryDelay == "" {
		return 5 * time.Second // default
	}

	duration, err := time.ParseDuration(c.RetryDelay)
	if err != nil {
		return 5 * time.Second // fallback to default
	}

	return duration
}

// GetHTTPReadTimeout parses HTTP read timeout
func (h *HTTPConfig) GetReadTimeout() time.Duration {
	if h.ReadTimeout == "" {
		return 30 * time.Second
	}

	duration, err := time.ParseDuration(h.ReadTimeout)
	if err != nil {
		return 30 * time.Second
	}

	return duration
}

// GetHTTPWriteTimeout parses HTTP write timeout
func (h *HTTPConfig) GetWriteTimeout() time.Duration {
	if h.WriteTimeout == "" {
		return 30 * time.Second
	}

	duration, err := time.ParseDuration(h.WriteTimeout)
	if err != nil {
		return 30 * time.Second
	}

	return duration
}

// GetMetricsInterval parses metrics collection interval
func (m *MetricsConfig) GetInterval() time.Duration {
	if m.Interval == "" {
		return 10 * time.Second
	}

	duration, err := time.ParseDuration(m.Interval)
	if err != nil {
		return 10 * time.Second
	}

	return duration
}

// Default returns a default configuration
func Default() *Config {
	return &Config{
		App: AppConfig{
			Name:        "blockchain-indexer",
			Version:     "0.1.0",
			Environment: "development",
		},
		Storage: StorageConfig{
			Type: "pebble",
			Pebble: PebbleConfig{
				Path:             "./data/pebble",
				CacheSize:        64 << 20, // 64MB
				MaxOpenFiles:     1000,
				WriteBufferSize:  64 << 20, // 64MB
				MaxConcurrentMem: 2,
				DisableWAL:       false,
				BytesPerSync:     512 << 10, // 512KB
			},
		},
		Server: ServerConfig{
			HTTP: HTTPConfig{
				Enabled:      true,
				Host:         "0.0.0.0",
				Port:         8080,
				TLSEnabled:   false,
				ReadTimeout:  "30s",
				WriteTimeout: "30s",
			},
			GRPC: GRPCConfig{
				Enabled:    true,
				Host:       "0.0.0.0",
				Port:       9090,
				TLSEnabled: false,
			},
			GraphQL: GraphQLConfig{
				Enabled:    true,
				Host:       "0.0.0.0",
				Port:       8081,
				Playground: true,
			},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
		Metrics: MetricsConfig{
			Enabled:  true,
			Host:     "0.0.0.0",
			Port:     9091,
			Path:     "/metrics",
			Interval: "10s",
		},
	}
}

// Save saves configuration to a YAML file
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

package service

import (
	"context"
	"fmt"
)

// DataProvider defines the interface for external data access APIs
// Following the Interface Segregation Principle
// Different providers (GraphQL, gRPC, REST) implement this interface
type DataProvider interface {
	// Lifecycle methods
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// Configuration
	GetType() ProviderType
	GetPort() int
	GetAddress() string
}

// ProviderType represents the type of data provider
type ProviderType string

const (
	// ProviderTypeGraphQL represents GraphQL API provider
	ProviderTypeGraphQL ProviderType = "graphql"

	// ProviderTypeGRPC represents gRPC API provider
	ProviderTypeGRPC ProviderType = "grpc"

	// ProviderTypeREST represents REST API provider
	ProviderTypeREST ProviderType = "rest"
)

// String returns the string representation of ProviderType
func (p ProviderType) String() string {
	return string(p)
}

// IsValid checks if the provider type is valid
func (p ProviderType) IsValid() bool {
	switch p {
	case ProviderTypeGraphQL, ProviderTypeGRPC, ProviderTypeREST:
		return true
	default:
		return false
	}
}

// ProviderConfig holds configuration for a data provider
type ProviderConfig struct {
	Type    ProviderType
	Host    string
	Port    int
	Enabled bool

	// TLS configuration
	EnableTLS  bool
	CertFile   string
	KeyFile    string
	CAFile     string

	// CORS configuration (for HTTP-based providers)
	EnableCORS     bool
	AllowedOrigins []string

	// Rate limiting
	EnableRateLimit bool
	RequestsPerMin  int

	// Authentication
	EnableAuth bool
	JWTSecret  string

	// Provider-specific options
	Options map[string]interface{}
}

// Validate validates the provider configuration
func (c *ProviderConfig) Validate() error {
	if !c.Type.IsValid() {
		return ErrInvalidProviderType
	}
	if c.Host == "" {
		c.Host = "localhost" // Default
	}
	if c.Port <= 0 || c.Port > 65535 {
		return ErrInvalidPort
	}
	if c.EnableTLS {
		if c.CertFile == "" || c.KeyFile == "" {
			return ErrInvalidTLSConfig
		}
	}
	return nil
}

// GetAddress returns the full address (host:port)
func (c *ProviderConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// DefaultProviderConfig returns a default provider configuration
func DefaultProviderConfig(providerType ProviderType) *ProviderConfig {
	var defaultPort int
	switch providerType {
	case ProviderTypeGraphQL:
		defaultPort = 8080
	case ProviderTypeGRPC:
		defaultPort = 9090
	case ProviderTypeREST:
		defaultPort = 8081
	default:
		defaultPort = 8080
	}

	return &ProviderConfig{
		Type:            providerType,
		Host:            "localhost",
		Port:            defaultPort,
		Enabled:         true,
		EnableTLS:       true, // Always recommend TLS
		EnableCORS:      true,
		AllowedOrigins:  []string{"*"},
		EnableRateLimit: true,
		RequestsPerMin:  1000,
		EnableAuth:      false,
		Options:         make(map[string]interface{}),
	}
}

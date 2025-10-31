package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/config"
	"gopkg.in/yaml.v3"
)

// NewConfigCmd creates a config command
func NewConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
		Long:  "Commands for managing blockchain indexer configuration",
	}

	configCmd.AddCommand(&cobra.Command{
		Use:   "generate",
		Short: "Generate example configuration file",
		Long:  "Generate an example configuration file (config.example.yaml)",
		Run: func(cmd *cobra.Command, args []string) {
			generateExampleConfig()
		},
	})

	configCmd.AddCommand(&cobra.Command{
		Use:   "validate [config-file]",
		Short: "Validate configuration file",
		Long:  "Validate a configuration file for correctness",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return validateConfig(args[0])
		},
	})

	return configCmd
}

func generateExampleConfig() {
	exampleConfig := `# Blockchain Indexer Configuration

app:
  name: blockchain-indexer
  env: development
  log_level: info

storage:
  type: pebbledb
  path: ./data
  cache_size_mb: 512
  write_buffer_mb: 64
  max_open_files: 1000

# Chain configurations
chains:
  - chain_id: ethereum
    chain_type: evm
    name: Ethereum Mainnet
    network: mainnet
    rpc_urls:
      - https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY
      - https://mainnet.infura.io/v3/YOUR_API_KEY
    start_block: 0
    enabled: true

  - chain_id: polygon
    chain_type: evm
    name: Polygon Mainnet
    network: mainnet
    rpc_urls:
      - https://polygon-rpc.com
    start_block: 0
    enabled: false

# Server configuration
server:
  http_port: 8080
  grpc_port: 9090
  graphql_enabled: true
  grpc_enabled: true
  rest_enabled: true
  enable_playground: true

# Logging configuration
logging:
  level: info
  format: json
  output: stdout
  file_path: ./logs/indexer.log
  max_size: 100
  max_backups: 3
  max_age: 7
  compress: true

# Metrics configuration
metrics:
  enabled: true
  port: 9091
  path: /metrics
`

	filename := "config.example.yaml"
	if err := os.WriteFile(filename, []byte(exampleConfig), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing example config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated example configuration file: %s\n", filename)
}

func validateConfig(configFile string) error {
	fmt.Printf("Validating configuration file: %s\n", configFile)

	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Print summary
	fmt.Println("\nConfiguration is valid!")
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  App Name:       %s\n", cfg.App.Name)
	fmt.Printf("  Environment:    %s\n", cfg.App.Environment)
	fmt.Printf("  Storage Type:   %s\n", cfg.Storage.Type)
	fmt.Printf("  Storage Path:   %s\n", cfg.Storage.Pebble.Path)
	fmt.Printf("  Chains:         %d configured\n", len(cfg.Chains))
	fmt.Printf("  HTTP Port:      %d\n", cfg.Server.HTTP.Port)
	fmt.Printf("  gRPC Port:      %d\n", cfg.Server.GRPC.Port)
	fmt.Printf("  GraphQL:        %v\n", cfg.Server.GraphQL.Enabled)
	fmt.Printf("  HTTP Enabled:   %v\n", cfg.Server.HTTP.Enabled)
	fmt.Printf("  Metrics Port:   %d\n", cfg.Metrics.Port)

	fmt.Printf("\nChains:\n")
	for _, chain := range cfg.Chains {
		status := "disabled"
		if chain.Enabled {
			status = "enabled"
		}
		fmt.Printf("  - %s (%s) - %s [%s]\n", chain.ChainID, chain.ChainType, chain.Name, status)
	}

	return nil
}

func printYAML(v interface{}) {
	data, err := yaml.Marshal(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling YAML: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

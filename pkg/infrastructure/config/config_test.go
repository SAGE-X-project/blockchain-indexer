package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("load valid config", func(t *testing.T) {
		// Create temp config file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		configYAML := `
app:
  name: test-indexer
  version: 1.0.0
  environment: development

storage:
  type: pebble
  pebble:
    path: ./data/test

chains:
  - chain_type: evm
    chain_id: ethereum
    name: Ethereum
    network: mainnet
    enabled: true
    rpc_endpoints:
      - http://localhost:8545
    start_block: 0
    batch_size: 100
    workers: 10
    confirmation_blocks: 12
    retry_attempts: 3
    retry_delay: 5s

server:
  http:
    enabled: true
    host: 0.0.0.0
    port: 8080
    tls_enabled: false
    read_timeout: 30s
    write_timeout: 30s

logging:
  level: info
  format: json
  output: stdout

metrics:
  enabled: true
  host: 0.0.0.0
  port: 9091
  path: /metrics
  interval: 10s
`

		if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		// Load config
		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Verify
		if cfg.App.Name != "test-indexer" {
			t.Errorf("App.Name = %v, want test-indexer", cfg.App.Name)
		}

		if cfg.Storage.Type != "pebble" {
			t.Errorf("Storage.Type = %v, want pebble", cfg.Storage.Type)
		}

		if len(cfg.Chains) != 1 {
			t.Errorf("len(Chains) = %v, want 1", len(cfg.Chains))
		}
	})

	t.Run("load non-existent file", func(t *testing.T) {
		_, err := Load("non-existent.yaml")
		if err == nil {
			t.Error("Load() should return error for non-existent file")
		}
	})

	t.Run("load invalid yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		invalidYAML := `
app:
  name: test
  invalid yaml here
`

		if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		_, err := Load(configPath)
		if err == nil {
			t.Error("Load() should return error for invalid YAML")
		}
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := Default()
		cfg.Chains = []ChainConfig{
			{
				ChainType:    "evm",
				ChainID:      "ethereum",
				Name:         "Ethereum",
				Enabled:      true,
				RPCEndpoints: []string{"http://localhost:8545"},
				BatchSize:    100,
				Workers:      10,
			},
		}

		err := cfg.Validate()
		if err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("missing app name", func(t *testing.T) {
		cfg := Default()
		cfg.App.Name = ""

		err := cfg.Validate()
		if err == nil {
			t.Error("Validate() should return error for missing app name")
		}
	})

	t.Run("missing storage type", func(t *testing.T) {
		cfg := Default()
		cfg.Storage.Type = ""

		err := cfg.Validate()
		if err == nil {
			t.Error("Validate() should return error for missing storage type")
		}
	})

	t.Run("invalid storage type", func(t *testing.T) {
		cfg := Default()
		cfg.Storage.Type = "invalid"

		err := cfg.Validate()
		if err == nil {
			t.Error("Validate() should return error for invalid storage type")
		}
	})

	t.Run("no chains configured", func(t *testing.T) {
		cfg := Default()
		cfg.Chains = []ChainConfig{}

		err := cfg.Validate()
		if err == nil {
			t.Error("Validate() should return error for no chains")
		}
	})
}

func TestChainConfig_Validate(t *testing.T) {
	t.Run("valid chain config", func(t *testing.T) {
		chain := ChainConfig{
			ChainType:    "evm",
			ChainID:      "ethereum",
			Name:         "Ethereum",
			RPCEndpoints: []string{"http://localhost:8545"},
			BatchSize:    100,
			Workers:      10,
		}

		err := chain.Validate()
		if err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("missing chain type", func(t *testing.T) {
		chain := ChainConfig{
			ChainID:      "ethereum",
			Name:         "Ethereum",
			RPCEndpoints: []string{"http://localhost:8545"},
			BatchSize:    100,
			Workers:      10,
		}

		err := chain.Validate()
		if err == nil {
			t.Error("Validate() should return error for missing chain type")
		}
	})

	t.Run("invalid chain type", func(t *testing.T) {
		chain := ChainConfig{
			ChainType:    "invalid",
			ChainID:      "ethereum",
			Name:         "Ethereum",
			RPCEndpoints: []string{"http://localhost:8545"},
			BatchSize:    100,
			Workers:      10,
		}

		err := chain.Validate()
		if err == nil {
			t.Error("Validate() should return error for invalid chain type")
		}
	})

	t.Run("no rpc endpoints", func(t *testing.T) {
		chain := ChainConfig{
			ChainType:    "evm",
			ChainID:      "ethereum",
			Name:         "Ethereum",
			RPCEndpoints: []string{},
			BatchSize:    100,
			Workers:      10,
		}

		err := chain.Validate()
		if err == nil {
			t.Error("Validate() should return error for no RPC endpoints")
		}
	})

	t.Run("invalid batch size", func(t *testing.T) {
		chain := ChainConfig{
			ChainType:    "evm",
			ChainID:      "ethereum",
			Name:         "Ethereum",
			RPCEndpoints: []string{"http://localhost:8545"},
			BatchSize:    0,
			Workers:      10,
		}

		err := chain.Validate()
		if err == nil {
			t.Error("Validate() should return error for invalid batch size")
		}
	})
}

func TestConfig_GetEnabledChains(t *testing.T) {
	cfg := Default()
	cfg.Chains = []ChainConfig{
		{
			ChainID: "ethereum",
			Enabled: true,
		},
		{
			ChainID: "bsc",
			Enabled: false,
		},
		{
			ChainID: "polygon",
			Enabled: true,
		},
	}

	enabled := cfg.GetEnabledChains()

	if len(enabled) != 2 {
		t.Errorf("len(enabled) = %v, want 2", len(enabled))
	}

	if enabled[0].ChainID != "ethereum" {
		t.Errorf("enabled[0].ChainID = %v, want ethereum", enabled[0].ChainID)
	}

	if enabled[1].ChainID != "polygon" {
		t.Errorf("enabled[1].ChainID = %v, want polygon", enabled[1].ChainID)
	}
}

func TestConfig_GetChainByID(t *testing.T) {
	cfg := Default()
	cfg.Chains = []ChainConfig{
		{ChainID: "ethereum"},
		{ChainID: "bsc"},
	}

	t.Run("existing chain", func(t *testing.T) {
		chain, found := cfg.GetChainByID("ethereum")
		if !found {
			t.Error("GetChainByID() should find ethereum")
		}

		if chain.ChainID != "ethereum" {
			t.Errorf("ChainID = %v, want ethereum", chain.ChainID)
		}
	})

	t.Run("non-existent chain", func(t *testing.T) {
		_, found := cfg.GetChainByID("nonexistent")
		if found {
			t.Error("GetChainByID() should not find non-existent chain")
		}
	})
}

func TestChainConfig_GetRetryDelay(t *testing.T) {
	t.Run("valid delay", func(t *testing.T) {
		chain := ChainConfig{RetryDelay: "10s"}
		delay := chain.GetRetryDelay()

		if delay.Seconds() != 10 {
			t.Errorf("GetRetryDelay() = %v, want 10s", delay)
		}
	})

	t.Run("empty delay uses default", func(t *testing.T) {
		chain := ChainConfig{RetryDelay: ""}
		delay := chain.GetRetryDelay()

		if delay.Seconds() != 5 {
			t.Errorf("GetRetryDelay() = %v, want 5s (default)", delay)
		}
	})

	t.Run("invalid delay uses default", func(t *testing.T) {
		chain := ChainConfig{RetryDelay: "invalid"}
		delay := chain.GetRetryDelay()

		if delay.Seconds() != 5 {
			t.Errorf("GetRetryDelay() = %v, want 5s (default)", delay)
		}
	})
}

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg == nil {
		t.Fatal("Default() returned nil")
	}

	if cfg.App.Name != "blockchain-indexer" {
		t.Errorf("App.Name = %v, want blockchain-indexer", cfg.App.Name)
	}

	if cfg.Storage.Type != "pebble" {
		t.Errorf("Storage.Type = %v, want pebble", cfg.Storage.Type)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %v, want info", cfg.Logging.Level)
	}
}

func TestConfig_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := Default()
	cfg.Chains = []ChainConfig{
		{
			ChainType:    "evm",
			ChainID:      "ethereum",
			Name:         "Ethereum",
			RPCEndpoints: []string{"http://localhost:8545"},
			BatchSize:    100,
			Workers:      10,
		},
	}

	// Save config
	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Save() should create config file")
	}

	// Load and verify
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.App.Name != cfg.App.Name {
		t.Errorf("Loaded config name mismatch")
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Run("load from CONFIG_PATH env", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		// Create config
		cfg := Default()
		cfg.Chains = []ChainConfig{
			{
				ChainType:    "evm",
				ChainID:      "ethereum",
				Name:         "Ethereum",
				RPCEndpoints: []string{"http://localhost:8545"},
				BatchSize:    100,
				Workers:      10,
			},
		}
		if err := cfg.Save(configPath); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Set env var
		os.Setenv("CONFIG_PATH", configPath)
		defer os.Unsetenv("CONFIG_PATH")

		// Load
		loaded, err := LoadFromEnv()
		if err != nil {
			t.Fatalf("LoadFromEnv() error = %v", err)
		}

		if loaded.App.Name != "blockchain-indexer" {
			t.Errorf("Loaded config name mismatch")
		}
	})
}

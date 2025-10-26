package models

import (
	"testing"
)

func TestNewChain(t *testing.T) {
	chainType := ChainTypeEVM
	chainID := "ethereum"
	name := "Ethereum Mainnet"

	chain := NewChain(chainType, chainID, name)

	if chain.ChainType != chainType {
		t.Errorf("NewChain().ChainType = %v, want %v", chain.ChainType, chainType)
	}

	if chain.ChainID != chainID {
		t.Errorf("NewChain().ChainID = %v, want %v", chain.ChainID, chainID)
	}

	if chain.Name != name {
		t.Errorf("NewChain().Name = %v, want %v", chain.Name, name)
	}

	if !chain.Enabled {
		t.Error("NewChain().Enabled should be true")
	}

	if chain.BatchSize != 100 {
		t.Errorf("NewChain().BatchSize = %v, want 100", chain.BatchSize)
	}

	if chain.Workers != 10 {
		t.Errorf("NewChain().Workers = %v, want 10", chain.Workers)
	}

	if chain.Status != ChainStatusIdle {
		t.Errorf("NewChain().Status = %v, want %v", chain.Status, ChainStatusIdle)
	}

	if chain.RPCEndpoints == nil || len(chain.RPCEndpoints) != 0 {
		t.Error("NewChain().RPCEndpoints should be empty slice")
	}

	if chain.Config == nil {
		t.Error("NewChain().Config should be initialized")
	}
}

func TestChain_Validate(t *testing.T) {
	tests := []struct {
		name    string
		chain   *Chain
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid chain",
			chain: &Chain{
				ChainType:    ChainTypeEVM,
				ChainID:      "ethereum",
				Name:         "Ethereum",
				RPCEndpoints: []string{"http://localhost:8545"},
				BatchSize:    100,
				Workers:      10,
			},
			wantErr: false,
		},
		{
			name: "invalid chain type",
			chain: &Chain{
				ChainType:    ChainType("invalid"),
				ChainID:      "ethereum",
				Name:         "Ethereum",
				RPCEndpoints: []string{"http://localhost:8545"},
				BatchSize:    100,
				Workers:      10,
			},
			wantErr: true,
			errMsg:  "invalid chain type",
		},
		{
			name: "empty chain ID",
			chain: &Chain{
				ChainType:    ChainTypeEVM,
				ChainID:      "",
				Name:         "Ethereum",
				RPCEndpoints: []string{"http://localhost:8545"},
				BatchSize:    100,
				Workers:      10,
			},
			wantErr: true,
			errMsg:  "invalid chain ID",
		},
		{
			name: "empty name",
			chain: &Chain{
				ChainType:    ChainTypeEVM,
				ChainID:      "ethereum",
				Name:         "",
				RPCEndpoints: []string{"http://localhost:8545"},
				BatchSize:    100,
				Workers:      10,
			},
			wantErr: true,
			errMsg:  "chain name is required",
		},
		{
			name: "no RPC endpoints",
			chain: &Chain{
				ChainType:    ChainTypeEVM,
				ChainID:      "ethereum",
				Name:         "Ethereum",
				RPCEndpoints: []string{},
				BatchSize:    100,
				Workers:      10,
			},
			wantErr: true,
			errMsg:  "at least one RPC endpoint is required",
		},
		{
			name: "invalid batch size",
			chain: &Chain{
				ChainType:    ChainTypeEVM,
				ChainID:      "ethereum",
				Name:         "Ethereum",
				RPCEndpoints: []string{"http://localhost:8545"},
				BatchSize:    0,
				Workers:      10,
			},
			wantErr: true,
			errMsg:  "batch size must be positive",
		},
		{
			name: "invalid workers",
			chain: &Chain{
				ChainType:    ChainTypeEVM,
				ChainID:      "ethereum",
				Name:         "Ethereum",
				RPCEndpoints: []string{"http://localhost:8545"},
				BatchSize:    100,
				Workers:      0,
			},
			wantErr: true,
			errMsg:  "workers must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.chain.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Chain.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Chain.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestChain_IsSynced(t *testing.T) {
	tests := []struct {
		name   string
		status ChainStatus
		want   bool
	}{
		{
			name:   "live status",
			status: ChainStatusLive,
			want:   true,
		},
		{
			name:   "syncing status",
			status: ChainStatusSyncing,
			want:   false,
		},
		{
			name:   "idle status",
			status: ChainStatusIdle,
			want:   false,
		},
		{
			name:   "error status",
			status: ChainStatusError,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := &Chain{Status: tt.status}
			if got := chain.IsSynced(); got != tt.want {
				t.Errorf("Chain.IsSynced() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChain_GetSyncProgress(t *testing.T) {
	tests := []struct {
		name              string
		latestIndexed     uint64
		latestChain       uint64
		expectedProgress  float64
	}{
		{
			name:             "fully synced",
			latestIndexed:    1000,
			latestChain:      1000,
			expectedProgress: 100.0,
		},
		{
			name:             "half synced",
			latestIndexed:    500,
			latestChain:      1000,
			expectedProgress: 50.0,
		},
		{
			name:             "no blocks",
			latestIndexed:    0,
			latestChain:      0,
			expectedProgress: 0.0,
		},
		{
			name:             "just started",
			latestIndexed:    1,
			latestChain:      1000,
			expectedProgress: 0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := &Chain{
				LatestIndexedBlock: tt.latestIndexed,
				LatestChainBlock:   tt.latestChain,
			}
			got := chain.GetSyncProgress()
			if got != tt.expectedProgress {
				t.Errorf("Chain.GetSyncProgress() = %v, want %v", got, tt.expectedProgress)
			}
		})
	}
}

func TestChain_GetBlocksBehind(t *testing.T) {
	tests := []struct {
		name          string
		latestIndexed uint64
		latestChain   uint64
		want          uint64
	}{
		{
			name:          "synced",
			latestIndexed: 1000,
			latestChain:   1000,
			want:          0,
		},
		{
			name:          "100 blocks behind",
			latestIndexed: 900,
			latestChain:   1000,
			want:          100,
		},
		{
			name:          "ahead (shouldn't happen)",
			latestIndexed: 1100,
			latestChain:   1000,
			want:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := &Chain{
				LatestIndexedBlock: tt.latestIndexed,
				LatestChainBlock:   tt.latestChain,
			}
			if got := chain.GetBlocksBehind(); got != tt.want {
				t.Errorf("Chain.GetBlocksBehind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChain_UpdateStatus(t *testing.T) {
	chain := NewChain(ChainTypeEVM, "ethereum", "Ethereum")
	initialTime := chain.LastUpdated

	// Small delay to ensure time difference
	newStatus := ChainStatusSyncing
	chain.UpdateStatus(newStatus)

	if chain.Status != newStatus {
		t.Errorf("Chain.UpdateStatus() Status = %v, want %v", chain.Status, newStatus)
	}

	if !chain.LastUpdated.After(initialTime) {
		t.Error("Chain.UpdateStatus() should update LastUpdated time")
	}
}

func TestChain_UpdateLatestBlock(t *testing.T) {
	chain := NewChain(ChainTypeEVM, "ethereum", "Ethereum")
	initialTime := chain.LastUpdated

	indexedBlock := uint64(1000)
	chainBlock := uint64(1100)

	chain.UpdateLatestBlock(indexedBlock, chainBlock)

	if chain.LatestIndexedBlock != indexedBlock {
		t.Errorf("Chain.UpdateLatestBlock() LatestIndexedBlock = %v, want %v", chain.LatestIndexedBlock, indexedBlock)
	}

	if chain.LatestChainBlock != chainBlock {
		t.Errorf("Chain.UpdateLatestBlock() LatestChainBlock = %v, want %v", chain.LatestChainBlock, chainBlock)
	}

	if !chain.LastUpdated.After(initialTime) {
		t.Error("Chain.UpdateLatestBlock() should update LastUpdated time")
	}
}

func TestChain_Config(t *testing.T) {
	chain := NewChain(ChainTypeEVM, "ethereum", "Ethereum")

	t.Run("SetConfig and GetConfig", func(t *testing.T) {
		key := "max_retries"
		value := 3

		chain.SetConfig(key, value)

		got, exists := chain.GetConfig(key)
		if !exists {
			t.Error("GetConfig() should return true for existing key")
		}

		if got != value {
			t.Errorf("GetConfig() = %v, want %v", got, value)
		}
	})

	t.Run("GetConfig for non-existent key", func(t *testing.T) {
		_, exists := chain.GetConfig("non_existent")
		if exists {
			t.Error("GetConfig() should return false for non-existent key")
		}
	})

	t.Run("SetConfig with nil config", func(t *testing.T) {
		chain.Config = nil
		chain.SetConfig("key", "value")

		if chain.Config == nil {
			t.Error("SetConfig() should initialize Config if nil")
		}
	})
}

func TestChain_GetStats(t *testing.T) {
	chain := NewChain(ChainTypeEVM, "ethereum", "Ethereum")
	chain.LatestIndexedBlock = 900
	chain.LatestChainBlock = 1000
	chain.Status = ChainStatusSyncing

	stats := chain.GetStats()

	if stats.ChainID != chain.ChainID {
		t.Errorf("GetStats().ChainID = %v, want %v", stats.ChainID, chain.ChainID)
	}

	if stats.ChainType != chain.ChainType {
		t.Errorf("GetStats().ChainType = %v, want %v", stats.ChainType, chain.ChainType)
	}

	if stats.LatestIndexedBlock != chain.LatestIndexedBlock {
		t.Errorf("GetStats().LatestIndexedBlock = %v, want %v", stats.LatestIndexedBlock, chain.LatestIndexedBlock)
	}

	if stats.LatestChainBlock != chain.LatestChainBlock {
		t.Errorf("GetStats().LatestChainBlock = %v, want %v", stats.LatestChainBlock, chain.LatestChainBlock)
	}

	if stats.BlocksBehind != 100 {
		t.Errorf("GetStats().BlocksBehind = %v, want 100", stats.BlocksBehind)
	}

	if stats.SyncProgress != 90.0 {
		t.Errorf("GetStats().SyncProgress = %v, want 90.0", stats.SyncProgress)
	}

	if stats.Status != chain.Status {
		t.Errorf("GetStats().Status = %v, want %v", stats.Status, chain.Status)
	}
}

func TestChainStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status ChainStatus
		want   string
	}{
		{
			name:   "idle",
			status: ChainStatusIdle,
			want:   "idle",
		},
		{
			name:   "syncing",
			status: ChainStatusSyncing,
			want:   "syncing",
		},
		{
			name:   "live",
			status: ChainStatusLive,
			want:   "live",
		},
		{
			name:   "error",
			status: ChainStatusError,
			want:   "error",
		},
		{
			name:   "paused",
			status: ChainStatusPaused,
			want:   "paused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("ChainStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmark tests
func BenchmarkNewChain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewChain(ChainTypeEVM, "ethereum", "Ethereum")
	}
}

func BenchmarkChain_Validate(b *testing.B) {
	chain := NewChain(ChainTypeEVM, "ethereum", "Ethereum")
	chain.RPCEndpoints = []string{"http://localhost:8545"}
	for i := 0; i < b.N; i++ {
		_ = chain.Validate()
	}
}

func BenchmarkChain_GetSyncProgress(b *testing.B) {
	chain := &Chain{
		LatestIndexedBlock: 900,
		LatestChainBlock:   1000,
	}
	for i := 0; i < b.N; i++ {
		_ = chain.GetSyncProgress()
	}
}

func BenchmarkChain_GetStats(b *testing.B) {
	chain := NewChain(ChainTypeEVM, "ethereum", "Ethereum")
	chain.LatestIndexedBlock = 900
	chain.LatestChainBlock = 1000
	for i := 0; i < b.N; i++ {
		_ = chain.GetStats()
	}
}

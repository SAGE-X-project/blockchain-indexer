package indexer

import (
	"testing"
	"time"
)

func TestDefaultWorkerPoolConfig(t *testing.T) {
	cfg := DefaultWorkerPoolConfig()

	if cfg.WorkerCount <= 0 {
		t.Error("WorkerCount should be positive")
	}

	if cfg.QueueSize <= 0 {
		t.Error("QueueSize should be positive")
	}

	if cfg.ResultSize <= 0 {
		t.Error("ResultSize should be positive")
	}
}

func TestDefaultBlockIndexerConfig(t *testing.T) {
	chainID := "ethereum"
	cfg := DefaultBlockIndexerConfig(chainID)

	if cfg.ChainID != chainID {
		t.Errorf("ChainID = %s, want %s", cfg.ChainID, chainID)
	}

	if cfg.BatchSize <= 0 {
		t.Error("BatchSize should be positive")
	}

	if cfg.WorkerCount <= 0 {
		t.Error("WorkerCount should be positive")
	}

	if cfg.PollInterval <= 0 {
		t.Error("PollInterval should be positive")
	}
}

func TestJobType_String(t *testing.T) {
	tests := []struct {
		jobType JobType
		want    string
	}{
		{JobTypeBlock, "block"},
		{JobTypeBlockRange, "block_range"},
		{JobTypeTransaction, "transaction"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.jobType.String(); got != tt.want {
				t.Errorf("JobType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkerPoolStats_String(t *testing.T) {
	stats := &WorkerPoolStats{
		WorkerCount:     10,
		ActiveWorkers:   8,
		TotalJobs:       100,
		CompletedJobs:   90,
		FailedJobs:      5,
		QueueLength:     10,
		QueueCapacity:   100,
		ResultsLength:   5,
		ResultsCapacity: 100,
		AvgDuration:     100 * time.Millisecond,
	}

	str := stats.String()
	if str == "" {
		t.Error("String() should not be empty")
	}
}

func TestGap_String(t *testing.T) {
	gap := &Gap{
		ChainID:    "ethereum",
		StartBlock: 100,
		EndBlock:   200,
		Size:       101,
	}

	str := gap.String()
	if str == "" {
		t.Error("String() should not be empty")
	}
}

func TestProgress_String(t *testing.T) {
	progress := &Progress{
		ChainID:            "ethereum",
		ChainType:          "evm",
		LatestIndexedBlock: 1000,
		LatestChainBlock:   2000,
		BlocksBehind:       1000,
		ProgressPercentage: 50.0,
		BlocksPerSecond:    10.0,
		EstimatedTimeLeft:  100 * time.Second,
		Status:             "syncing",
	}

	str := progress.String()
	if str == "" {
		t.Error("String() should not be empty")
	}
}

// Note: Full integration tests would require mocking the dependencies
// (adapter, repositories, processor, etc.). For now, we focus on unit tests
// of the configuration and data structures.

package metrics

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Run("create metrics with default config", func(t *testing.T) {
		m := New(nil)

		if m == nil {
			t.Fatal("New() returned nil")
		}

		if m.registry == nil {
			t.Error("registry should not be nil")
		}

		if m.BlocksIndexed == nil {
			t.Error("BlocksIndexed should not be nil")
		}

		if m.TransactionsIndexed == nil {
			t.Error("TransactionsIndexed should not be nil")
		}
	})

	t.Run("create metrics with custom config", func(t *testing.T) {
		cfg := &Config{
			Enabled:  true,
			Host:     "localhost",
			Port:     9999,
			Path:     "/custom",
			Interval: 5 * time.Second,
		}

		m := New(cfg)

		if m == nil {
			t.Fatal("New() returned nil")
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Enabled {
		t.Error("Enabled should be true")
	}

	if cfg.Port != 9091 {
		t.Errorf("Port = %v, want 9091", cfg.Port)
	}

	if cfg.Path != "/metrics" {
		t.Errorf("Path = %v, want /metrics", cfg.Path)
	}
}

func TestMetrics_RecordBlockIndexed(t *testing.T) {
	m := New(nil)

	chainID := "ethereum"
	m.RecordBlockIndexed(chainID)

	// We can't easily verify the counter value, but we can check it doesn't panic
}

func TestMetrics_RecordBlockProcessed(t *testing.T) {
	m := New(nil)

	chainID := "ethereum"

	t.Run("record success", func(t *testing.T) {
		m.RecordBlockProcessed(chainID, true)
	})

	t.Run("record error", func(t *testing.T) {
		m.RecordBlockProcessed(chainID, false)
	})
}

func TestMetrics_RecordBlockProcessTime(t *testing.T) {
	m := New(nil)

	chainID := "ethereum"
	duration := 100 * time.Millisecond

	m.RecordBlockProcessTime(chainID, duration)
}

func TestMetrics_UpdateLatestBlockHeight(t *testing.T) {
	m := New(nil)

	chainID := "ethereum"
	height := uint64(12345)

	m.UpdateLatestBlockHeight(chainID, height)
}

func TestMetrics_RecordTransactionIndexed(t *testing.T) {
	m := New(nil)

	chainID := "ethereum"
	m.RecordTransactionIndexed(chainID)
}

func TestMetrics_UpdateChainSyncStatus(t *testing.T) {
	m := New(nil)

	chainID := "ethereum"

	t.Run("stopped", func(t *testing.T) {
		m.UpdateChainSyncStatus(chainID, 0)
	})

	t.Run("syncing", func(t *testing.T) {
		m.UpdateChainSyncStatus(chainID, 1)
	})

	t.Run("synced", func(t *testing.T) {
		m.UpdateChainSyncStatus(chainID, 2)
	})
}

func TestMetrics_UpdateChainSyncProgress(t *testing.T) {
	m := New(nil)

	chainID := "ethereum"
	progress := 75.5

	m.UpdateChainSyncProgress(chainID, progress)
}

func TestMetrics_UpdateChainBlocksBehind(t *testing.T) {
	m := New(nil)

	chainID := "ethereum"
	blocks := uint64(100)

	m.UpdateChainBlocksBehind(chainID, blocks)
}

func TestMetrics_Handler(t *testing.T) {
	m := New(nil)

	handler := m.Handler()

	if handler == nil {
		t.Error("Handler() returned nil")
	}
}

func TestGlobalMetrics(t *testing.T) {
	t.Run("init global", func(t *testing.T) {
		cfg := DefaultConfig()
		InitGlobal(cfg)

		m := Global()
		if m == nil {
			t.Error("Global() returned nil")
		}
	})

	t.Run("use global functions", func(t *testing.T) {
		chainID := "ethereum"

		RecordBlockIndexed(chainID)
		RecordBlockProcessed(chainID, true)
		UpdateLatestBlockHeight(chainID, 12345)
		RecordTransactionIndexed(chainID)
		UpdateChainSyncStatus(chainID, 1)
	})
}

func BenchmarkMetrics_RecordBlockIndexed(b *testing.B) {
	m := New(nil)
	chainID := "ethereum"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.RecordBlockIndexed(chainID)
	}
}

func BenchmarkMetrics_RecordBlockProcessTime(b *testing.B) {
	m := New(nil)
	chainID := "ethereum"
	duration := 100 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.RecordBlockProcessTime(chainID, duration)
	}
}

func BenchmarkMetrics_UpdateLatestBlockHeight(b *testing.B) {
	m := New(nil)
	chainID := "ethereum"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.UpdateLatestBlockHeight(chainID, uint64(i))
	}
}

package models

import (
	"testing"
	"time"
)

func TestChainType_String(t *testing.T) {
	tests := []struct {
		name     string
		chainType ChainType
		want     string
	}{
		{
			name:     "EVM chain type",
			chainType: ChainTypeEVM,
			want:     "evm",
		},
		{
			name:     "Solana chain type",
			chainType: ChainTypeSolana,
			want:     "solana",
		},
		{
			name:     "Cosmos chain type",
			chainType: ChainTypeCosmos,
			want:     "cosmos",
		},
		{
			name:     "Polkadot chain type",
			chainType: ChainTypePolkadot,
			want:     "polkadot",
		},
		{
			name:     "Avalanche chain type",
			chainType: ChainTypeAvalanche,
			want:     "avalanche",
		},
		{
			name:     "Ripple chain type",
			chainType: ChainTypeRipple,
			want:     "ripple",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.chainType.String(); got != tt.want {
				t.Errorf("ChainType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChainType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		chainType ChainType
		want     bool
	}{
		{
			name:     "valid EVM",
			chainType: ChainTypeEVM,
			want:     true,
		},
		{
			name:     "valid Solana",
			chainType: ChainTypeSolana,
			want:     true,
		},
		{
			name:     "valid Cosmos",
			chainType: ChainTypeCosmos,
			want:     true,
		},
		{
			name:     "invalid chain type",
			chainType: ChainType("invalid"),
			want:     false,
		},
		{
			name:     "empty chain type",
			chainType: ChainType(""),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.chainType.IsValid(); got != tt.want {
				t.Errorf("ChainType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTxStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status TxStatus
		want   string
	}{
		{
			name:   "pending status",
			status: TxStatusPending,
			want:   "pending",
		},
		{
			name:   "success status",
			status: TxStatusSuccess,
			want:   "success",
		},
		{
			name:   "failed status",
			status: TxStatusFailed,
			want:   "failed",
		},
		{
			name:   "unknown status",
			status: TxStatus(99),
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("TxStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTimestamp(t *testing.T) {
	tests := []struct {
		name string
		unix int64
	}{
		{
			name: "current time",
			unix: time.Now().Unix(),
		},
		{
			name: "epoch time",
			unix: 0,
		},
		{
			name: "specific time",
			unix: 1609459200, // 2021-01-01 00:00:00 UTC
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTimestamp(tt.unix)

			if ts.Unix != tt.unix {
				t.Errorf("NewTimestamp().Unix = %v, want %v", ts.Unix, tt.unix)
			}

			expectedTime := time.Unix(tt.unix, 0)
			if !ts.Time.Equal(expectedTime) {
				t.Errorf("NewTimestamp().Time = %v, want %v", ts.Time, expectedTime)
			}

			if ts.Slot != nil {
				t.Errorf("NewTimestamp().Slot should be nil, got %v", *ts.Slot)
			}

			if ts.Epoch != nil {
				t.Errorf("NewTimestamp().Epoch should be nil, got %v", *ts.Epoch)
			}
		})
	}
}

func TestNewTimestampWithSlot(t *testing.T) {
	tests := []struct {
		name string
		unix int64
		slot uint64
	}{
		{
			name: "with slot number",
			unix: time.Now().Unix(),
			slot: 12345,
		},
		{
			name: "zero slot",
			unix: 1609459200,
			slot: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTimestampWithSlot(tt.unix, tt.slot)

			if ts.Unix != tt.unix {
				t.Errorf("NewTimestampWithSlot().Unix = %v, want %v", ts.Unix, tt.unix)
			}

			if ts.Slot == nil {
				t.Fatal("NewTimestampWithSlot().Slot should not be nil")
			}

			if *ts.Slot != tt.slot {
				t.Errorf("NewTimestampWithSlot().Slot = %v, want %v", *ts.Slot, tt.slot)
			}
		})
	}
}

// Benchmark tests
func BenchmarkChainType_IsValid(b *testing.B) {
	ct := ChainTypeEVM
	for i := 0; i < b.N; i++ {
		_ = ct.IsValid()
	}
}

func BenchmarkTxStatus_String(b *testing.B) {
	status := TxStatusSuccess
	for i := 0; i < b.N; i++ {
		_ = status.String()
	}
}

func BenchmarkNewTimestamp(b *testing.B) {
	unix := time.Now().Unix()
	for i := 0; i < b.N; i++ {
		_ = NewTimestamp(unix)
	}
}

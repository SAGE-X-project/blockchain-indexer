package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewTransaction(t *testing.T) {
	chainType := ChainTypeEVM
	chainID := "ethereum"
	hash := "0xtx123"

	tx := NewTransaction(chainType, chainID, hash)

	if tx.ChainType != chainType {
		t.Errorf("NewTransaction().ChainType = %v, want %v", tx.ChainType, chainType)
	}

	if tx.ChainID != chainID {
		t.Errorf("NewTransaction().ChainID = %v, want %v", tx.ChainID, chainID)
	}

	if tx.Hash != hash {
		t.Errorf("NewTransaction().Hash = %v, want %v", tx.Hash, hash)
	}

	if tx.Metadata == nil {
		t.Error("NewTransaction().Metadata should be initialized")
	}

	if tx.Logs == nil {
		t.Error("NewTransaction().Logs should be initialized")
	}

	if tx.Timestamp == nil {
		t.Error("NewTransaction().Timestamp should be initialized")
	}

	if tx.IndexedAt.IsZero() {
		t.Error("NewTransaction().IndexedAt should be set")
	}
}

func TestTransaction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      *Transaction
		wantErr error
	}{
		{
			name: "valid transaction",
			tx: &Transaction{
				ChainType: ChainTypeEVM,
				ChainID:   "ethereum",
				Hash:      "0xtx123",
				From:      "0xfrom",
				Timestamp: NewTimestamp(time.Now().Unix()),
			},
			wantErr: nil,
		},
		{
			name: "invalid chain type",
			tx: &Transaction{
				ChainType: ChainType("invalid"),
				ChainID:   "ethereum",
				Hash:      "0xtx123",
				From:      "0xfrom",
				Timestamp: NewTimestamp(time.Now().Unix()),
			},
			wantErr: ErrInvalidChainType,
		},
		{
			name: "empty chain ID",
			tx: &Transaction{
				ChainType: ChainTypeEVM,
				ChainID:   "",
				Hash:      "0xtx123",
				From:      "0xfrom",
				Timestamp: NewTimestamp(time.Now().Unix()),
			},
			wantErr: ErrInvalidChainID,
		},
		{
			name: "empty hash",
			tx: &Transaction{
				ChainType: ChainTypeEVM,
				ChainID:   "ethereum",
				Hash:      "",
				From:      "0xfrom",
				Timestamp: NewTimestamp(time.Now().Unix()),
			},
			wantErr: ErrInvalidTxHash,
		},
		{
			name: "empty from address",
			tx: &Transaction{
				ChainType: ChainTypeEVM,
				ChainID:   "ethereum",
				Hash:      "0xtx123",
				From:      "",
				Timestamp: NewTimestamp(time.Now().Unix()),
			},
			wantErr: ErrInvalidFromAddress,
		},
		{
			name: "nil timestamp",
			tx: &Transaction{
				ChainType: ChainTypeEVM,
				ChainID:   "ethereum",
				Hash:      "0xtx123",
				From:      "0xfrom",
				Timestamp: nil,
			},
			wantErr: ErrInvalidTimestamp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.Validate()
			if err != tt.wantErr {
				t.Errorf("Transaction.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransaction_IsContractCreation(t *testing.T) {
	tests := []struct {
		name string
		tx   *Transaction
		want bool
	}{
		{
			name: "contract creation",
			tx: &Transaction{
				To:              "",
				ContractAddress: "0xcontract",
			},
			want: true,
		},
		{
			name: "normal transaction",
			tx: &Transaction{
				To:              "0xto",
				ContractAddress: "",
			},
			want: false,
		},
		{
			name: "empty to without contract",
			tx: &Transaction{
				To:              "",
				ContractAddress: "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tx.IsContractCreation(); got != tt.want {
				t.Errorf("Transaction.IsContractCreation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransaction_IsSuccess(t *testing.T) {
	tests := []struct {
		name   string
		status TxStatus
		want   bool
	}{
		{
			name:   "success status",
			status: TxStatusSuccess,
			want:   true,
		},
		{
			name:   "failed status",
			status: TxStatusFailed,
			want:   false,
		},
		{
			name:   "pending status",
			status: TxStatusPending,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &Transaction{Status: tt.status}
			if got := tx.IsSuccess(); got != tt.want {
				t.Errorf("Transaction.IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransaction_Metadata(t *testing.T) {
	tx := NewTransaction(ChainTypeEVM, "ethereum", "0xtx123")

	t.Run("SetMetadata and GetMetadata", func(t *testing.T) {
		key := "gas_price_gwei"
		value := "100"

		tx.SetMetadata(key, value)

		got, exists := tx.GetMetadata(key)
		if !exists {
			t.Error("GetMetadata() should return true for existing key")
		}

		if got != value {
			t.Errorf("GetMetadata() = %v, want %v", got, value)
		}
	})

	t.Run("GetMetadata for non-existent key", func(t *testing.T) {
		_, exists := tx.GetMetadata("non_existent")
		if exists {
			t.Error("GetMetadata() should return false for non-existent key")
		}
	})

	t.Run("SetMetadata with nil metadata", func(t *testing.T) {
		tx.Metadata = nil
		tx.SetMetadata("key", "value")

		if tx.Metadata == nil {
			t.Error("SetMetadata() should initialize Metadata if nil")
		}
	})
}

func TestTransaction_ToSummary(t *testing.T) {
	tx := NewTransaction(ChainTypeEVM, "ethereum", "0xtx123")
	tx.BlockNumber = 12345
	tx.From = "0xfrom"
	tx.To = "0xto"
	tx.Value = "1000000000000000000" // 1 ETH
	tx.Status = TxStatusSuccess

	summary := tx.ToSummary()

	if summary.ChainType != tx.ChainType {
		t.Errorf("ToSummary().ChainType = %v, want %v", summary.ChainType, tx.ChainType)
	}

	if summary.ChainID != tx.ChainID {
		t.Errorf("ToSummary().ChainID = %v, want %v", summary.ChainID, tx.ChainID)
	}

	if summary.Hash != tx.Hash {
		t.Errorf("ToSummary().Hash = %v, want %v", summary.Hash, tx.Hash)
	}

	if summary.BlockNumber != tx.BlockNumber {
		t.Errorf("ToSummary().BlockNumber = %v, want %v", summary.BlockNumber, tx.BlockNumber)
	}

	if summary.From != tx.From {
		t.Errorf("ToSummary().From = %v, want %v", summary.From, tx.From)
	}

	if summary.To != tx.To {
		t.Errorf("ToSummary().To = %v, want %v", summary.To, tx.To)
	}

	if summary.Value != tx.Value {
		t.Errorf("ToSummary().Value = %v, want %v", summary.Value, tx.Value)
	}

	if summary.Status != tx.Status {
		t.Errorf("ToSummary().Status = %v, want %v", summary.Status, tx.Status)
	}
}

func TestTransaction_MarshalJSON(t *testing.T) {
	tx := NewTransaction(ChainTypeEVM, "ethereum", "0xtx123")
	tx.Status = TxStatusSuccess

	data, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Check that chain_type is a string
	if chainType, ok := result["chain_type"].(string); !ok {
		t.Error("chain_type should be a string in JSON")
	} else if chainType != "evm" {
		t.Errorf("chain_type = %v, want evm", chainType)
	}

	// Check that status is a string
	if status, ok := result["status"].(string); !ok {
		t.Error("status should be a string in JSON")
	} else if status != "success" {
		t.Errorf("status = %v, want success", status)
	}
}

func TestNewLog(t *testing.T) {
	index := uint64(0)
	address := "0xcontract"

	log := NewLog(index, address)

	if log.Index != index {
		t.Errorf("NewLog().Index = %v, want %v", log.Index, index)
	}

	if log.Address != address {
		t.Errorf("NewLog().Address = %v, want %v", log.Address, address)
	}

	if log.Topics == nil {
		t.Error("NewLog().Topics should be initialized")
	}

	if len(log.Topics) != 0 {
		t.Errorf("NewLog().Topics length = %v, want 0", len(log.Topics))
	}
}

// Benchmark tests
func BenchmarkNewTransaction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewTransaction(ChainTypeEVM, "ethereum", "0xtx123")
	}
}

func BenchmarkTransaction_Validate(b *testing.B) {
	tx := NewTransaction(ChainTypeEVM, "ethereum", "0xtx123")
	tx.From = "0xfrom"
	for i := 0; i < b.N; i++ {
		_ = tx.Validate()
	}
}

func BenchmarkTransaction_ToSummary(b *testing.B) {
	tx := NewTransaction(ChainTypeEVM, "ethereum", "0xtx123")
	for i := 0; i < b.N; i++ {
		_ = tx.ToSummary()
	}
}

func BenchmarkTransaction_MarshalJSON(b *testing.B) {
	tx := NewTransaction(ChainTypeEVM, "ethereum", "0xtx123")
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(tx)
	}
}

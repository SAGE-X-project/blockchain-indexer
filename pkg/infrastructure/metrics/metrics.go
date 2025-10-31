package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all application metrics
type Metrics struct {
	// Block metrics
	BlocksIndexed    *prometheus.CounterVec
	BlocksProcessed  *prometheus.CounterVec
	BlockProcessTime *prometheus.HistogramVec
	LatestBlockHeight *prometheus.GaugeVec

	// Transaction metrics
	TransactionsIndexed  *prometheus.CounterVec
	TransactionsProcessed *prometheus.CounterVec
	TransactionProcessTime *prometheus.HistogramVec

	// Storage metrics
	StorageReads  *prometheus.CounterVec
	StorageWrites *prometheus.CounterVec
	StorageErrors *prometheus.CounterVec
	StorageSize   *prometheus.GaugeVec

	// RPC metrics
	RPCRequests       *prometheus.CounterVec
	RPCErrors         *prometheus.CounterVec
	RPCLatency        *prometheus.HistogramVec
	RPCConnectionPool *prometheus.GaugeVec

	// Chain sync metrics
	ChainSyncStatus    *prometheus.GaugeVec   // 0=stopped, 1=syncing, 2=synced
	ChainSyncProgress  *prometheus.GaugeVec   // percentage
	ChainBlocksBehind  *prometheus.GaugeVec
	ChainSyncErrors    *prometheus.CounterVec

	// Application metrics
	AppUptime *prometheus.CounterVec
	AppInfo   *prometheus.GaugeVec

	registry *prometheus.Registry
	mu       sync.RWMutex
}

// Config holds metrics configuration
type Config struct {
	Enabled  bool
	Host     string
	Port     int
	Path     string
	Interval time.Duration
}

// DefaultConfig returns default metrics configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:  true,
		Host:     "0.0.0.0",
		Port:     9091,
		Path:     "/metrics",
		Interval: 10 * time.Second,
	}
}

// New creates a new metrics instance
func New(cfg *Config) *Metrics {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	registry := prometheus.NewRegistry()

	m := &Metrics{
		registry: registry,
	}

	// Initialize block metrics
	m.BlocksIndexed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "indexer_blocks_indexed_total",
			Help: "Total number of blocks indexed",
		},
		[]string{"chain_id"},
	)

	m.BlocksProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "indexer_blocks_processed_total",
			Help: "Total number of blocks processed",
		},
		[]string{"chain_id", "status"}, // status: success, error
	)

	m.BlockProcessTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "indexer_block_process_duration_seconds",
			Help:    "Time spent processing a block",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~16s
		},
		[]string{"chain_id"},
	)

	m.LatestBlockHeight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "indexer_latest_block_height",
			Help: "Latest indexed block height",
		},
		[]string{"chain_id"},
	)

	// Initialize transaction metrics
	m.TransactionsIndexed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "indexer_transactions_indexed_total",
			Help: "Total number of transactions indexed",
		},
		[]string{"chain_id"},
	)

	m.TransactionsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "indexer_transactions_processed_total",
			Help: "Total number of transactions processed",
		},
		[]string{"chain_id", "status"},
	)

	m.TransactionProcessTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "indexer_transaction_process_duration_seconds",
			Help:    "Time spent processing a transaction",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 15), // 0.1ms to ~1.6s
		},
		[]string{"chain_id"},
	)

	// Initialize storage metrics
	m.StorageReads = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "indexer_storage_reads_total",
			Help: "Total number of storage read operations",
		},
		[]string{"operation"},
	)

	m.StorageWrites = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "indexer_storage_writes_total",
			Help: "Total number of storage write operations",
		},
		[]string{"operation"},
	)

	m.StorageErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "indexer_storage_errors_total",
			Help: "Total number of storage errors",
		},
		[]string{"operation", "error_type"},
	)

	m.StorageSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "indexer_storage_size_bytes",
			Help: "Storage size in bytes",
		},
		[]string{"chain_id"},
	)

	// Initialize RPC metrics
	m.RPCRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "indexer_rpc_requests_total",
			Help: "Total number of RPC requests",
		},
		[]string{"chain_id", "method"},
	)

	m.RPCErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "indexer_rpc_errors_total",
			Help: "Total number of RPC errors",
		},
		[]string{"chain_id", "method", "error_type"},
	)

	m.RPCLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "indexer_rpc_latency_seconds",
			Help:    "RPC request latency",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 12), // 10ms to ~40s
		},
		[]string{"chain_id", "method"},
	)

	m.RPCConnectionPool = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "indexer_rpc_connection_pool",
			Help: "Number of active RPC connections",
		},
		[]string{"chain_id"},
	)

	// Initialize chain sync metrics
	m.ChainSyncStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "indexer_chain_sync_status",
			Help: "Chain synchronization status (0=stopped, 1=syncing, 2=synced)",
		},
		[]string{"chain_id"},
	)

	m.ChainSyncProgress = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "indexer_chain_sync_progress",
			Help: "Chain synchronization progress percentage (0-100)",
		},
		[]string{"chain_id"},
	)

	m.ChainBlocksBehind = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "indexer_chain_blocks_behind",
			Help: "Number of blocks behind the chain head",
		},
		[]string{"chain_id"},
	)

	m.ChainSyncErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "indexer_chain_sync_errors_total",
			Help: "Total number of chain sync errors",
		},
		[]string{"chain_id", "error_type"},
	)

	// Initialize application metrics
	m.AppUptime = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "indexer_uptime_seconds",
			Help: "Application uptime in seconds",
		},
		[]string{"version"},
	)

	m.AppInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "indexer_info",
			Help: "Application information",
		},
		[]string{"version", "environment"},
	)

	// Register all metrics
	m.register()

	return m
}

// register registers all metrics with the registry
func (m *Metrics) register() {
	m.registry.MustRegister(
		// Block metrics
		m.BlocksIndexed,
		m.BlocksProcessed,
		m.BlockProcessTime,
		m.LatestBlockHeight,

		// Transaction metrics
		m.TransactionsIndexed,
		m.TransactionsProcessed,
		m.TransactionProcessTime,

		// Storage metrics
		m.StorageReads,
		m.StorageWrites,
		m.StorageErrors,
		m.StorageSize,

		// RPC metrics
		m.RPCRequests,
		m.RPCErrors,
		m.RPCLatency,
		m.RPCConnectionPool,

		// Chain sync metrics
		m.ChainSyncStatus,
		m.ChainSyncProgress,
		m.ChainBlocksBehind,
		m.ChainSyncErrors,

		// Application metrics
		m.AppUptime,
		m.AppInfo,
	)
}

// RecordBlockIndexed records a block being indexed
func (m *Metrics) RecordBlockIndexed(chainID string) {
	m.BlocksIndexed.WithLabelValues(chainID).Inc()
}

// RecordBlockProcessed records a block being processed
func (m *Metrics) RecordBlockProcessed(chainID string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	m.BlocksProcessed.WithLabelValues(chainID, status).Inc()
}

// RecordBlockProcessTime records block processing time
func (m *Metrics) RecordBlockProcessTime(chainID string, duration time.Duration) {
	m.BlockProcessTime.WithLabelValues(chainID).Observe(duration.Seconds())
}

// UpdateLatestBlockHeight updates the latest block height
func (m *Metrics) UpdateLatestBlockHeight(chainID string, height uint64) {
	m.LatestBlockHeight.WithLabelValues(chainID).Set(float64(height))
}

// RecordTransactionIndexed records a transaction being indexed
func (m *Metrics) RecordTransactionIndexed(chainID string) {
	m.TransactionsIndexed.WithLabelValues(chainID).Inc()
}

// UpdateChainSyncStatus updates chain sync status
func (m *Metrics) UpdateChainSyncStatus(chainID string, status int) {
	// 0=stopped, 1=syncing, 2=synced
	m.ChainSyncStatus.WithLabelValues(chainID).Set(float64(status))
}

// UpdateChainSyncProgress updates chain sync progress
func (m *Metrics) UpdateChainSyncProgress(chainID string, progress float64) {
	m.ChainSyncProgress.WithLabelValues(chainID).Set(progress)
}

// UpdateChainBlocksBehind updates the number of blocks behind
func (m *Metrics) UpdateChainBlocksBehind(chainID string, blocks uint64) {
	m.ChainBlocksBehind.WithLabelValues(chainID).Set(float64(blocks))
}

// Handler returns the HTTP handler for Prometheus metrics
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// StartServer starts the metrics HTTP server
func (m *Metrics) StartServer(cfg *Config) error {
	if !cfg.Enabled {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	mux := http.NewServeMux()
	mux.Handle(cfg.Path, m.Handler())

	return http.ListenAndServe(addr, mux)
}

// Global metrics instance
var (
	globalMetrics *Metrics
	once          sync.Once
)

// InitGlobal initializes the global metrics instance
func InitGlobal(cfg *Config) {
	once.Do(func() {
		globalMetrics = New(cfg)
	})
}

// Global returns the global metrics instance
func Global() *Metrics {
	if globalMetrics == nil {
		InitGlobal(DefaultConfig())
	}
	return globalMetrics
}

// RecordBlockIndexed records a block being indexed using the global metrics
func RecordBlockIndexed(chainID string) {
	Global().RecordBlockIndexed(chainID)
}

// RecordBlockProcessed records a block being processed using the global metrics
func RecordBlockProcessed(chainID string, success bool) {
	Global().RecordBlockProcessed(chainID, success)
}

// UpdateLatestBlockHeight updates the latest block height using the global metrics
func UpdateLatestBlockHeight(chainID string, height uint64) {
	Global().UpdateLatestBlockHeight(chainID, height)
}

// RecordTransactionIndexed records a transaction being indexed using the global metrics
func RecordTransactionIndexed(chainID string) {
	Global().RecordTransactionIndexed(chainID)
}

// UpdateChainSyncStatus updates chain sync status using the global metrics
func UpdateChainSyncStatus(chainID string, status int) {
	Global().UpdateChainSyncStatus(chainID, status)
}

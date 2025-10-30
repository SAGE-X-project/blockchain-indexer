package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/sage-x-project/blockchain-indexer/pkg/application/indexer"
	"github.com/sage-x-project/blockchain-indexer/pkg/application/processor"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/config"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/event"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/metrics"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/storage/pebble"
)

var (
	indexConfigFile string
	chainID         string
)

// NewIndexCmd creates an index command
func NewIndexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Start blockchain indexing",
		Long: `Start the blockchain indexer to sync and index blockchain data.

The indexer will:
  - Connect to configured blockchain RPC endpoints
  - Sync blocks and transactions
  - Detect and recover gaps
  - Emit events for real-time updates
  - Track progress and metrics`,
		RunE: runIndexer,
	}

	cmd.Flags().StringVarP(&indexConfigFile, "config", "c", "config.yaml", "Path to configuration file")
	cmd.Flags().StringVar(&chainID, "chain", "", "Index specific chain only (optional)")

	return cmd
}

func runIndexer(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(indexConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Filter chains if specific chain ID is provided
	var chainsToIndex []*config.ChainConfig
	if chainID != "" {
		found := false
		for i, chain := range cfg.Chains {
			if chain.ChainID == chainID && chain.Enabled {
				chainsToIndex = append(chainsToIndex, &cfg.Chains[i])
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("chain %s not found or not enabled", chainID)
		}
	} else {
		// Index all enabled chains
		for i, chain := range cfg.Chains {
			if chain.Enabled {
				chainsToIndex = append(chainsToIndex, &cfg.Chains[i])
			}
		}
	}

	if len(chainsToIndex) == 0 {
		return fmt.Errorf("no enabled chains found in configuration")
	}

	fmt.Println("🚀 Blockchain Indexer")
	fmt.Printf("📋 Chains to index: %d\n", len(chainsToIndex))
	for _, chain := range chainsToIndex {
		fmt.Printf("  • %s (%s): %s\n", chain.ChainID, chain.ChainType, chain.Name)
	}
	fmt.Println()

	// Initialize logger
	logCfg := &logger.Config{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		Output:     cfg.Logging.Output,
		FilePath:   cfg.Logging.FilePath,
		MaxSize:    cfg.Logging.MaxSize,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge,
		Compress:   cfg.Logging.Compress,
	}
	log, err := logger.New(logCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer log.Sync()

	log.Info("indexer starting",
		zap.Int("chain_count", len(chainsToIndex)),
		zap.String("config_file", indexConfigFile),
	)

	// Initialize metrics server
	metricsCfg := &metrics.Config{
		Enabled: cfg.Metrics.Enabled,
		Port:    cfg.Metrics.Port,
		Path:    cfg.Metrics.Path,
	}
	appMetrics := metrics.New(metricsCfg)

	// Start metrics server in background
	if cfg.Metrics.Enabled {
		go func() {
			metricsAddr := fmt.Sprintf(":%d", cfg.Metrics.Port)
			log.Info("metrics server starting", zap.String("address", metricsAddr))
			if err := appMetrics.StartServer(metricsCfg); err != nil {
				log.Error("metrics server error", zap.Error(err))
			}
		}()
	}

	// Initialize storage (PebbleDB)
	storagePath := cfg.Storage.Pebble.Path
	if storagePath == "" {
		storagePath = "./data/indexer"
	}
	storagePath = filepath.Clean(storagePath)

	log.Info("initializing storage", zap.String("path", storagePath))
	storage, err := pebble.NewStorage(&pebble.Config{
		Path:      storagePath,
		CacheSize: 128 << 20, // 128 MB cache
	})
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer storage.Close()

	// Initialize event bus
	eventBusConfig := &event.EventBusConfig{
		WorkerCount: 10,
		QueueSize:   10000,
	}
	eventBus := event.NewEventBus(eventBusConfig, log)
	if err := eventBus.Start(); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}
	defer eventBus.Stop()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create indexers for each chain
	indexers := make([]*indexer.BlockIndexer, 0, len(chainsToIndex))
	var indexersMu sync.Mutex

	for _, chainCfg := range chainsToIndex {
		log.Info("initializing chain indexer",
			zap.String("chain_id", chainCfg.ChainID),
			zap.String("chain_type", chainCfg.ChainType),
		)

		// Create chain adapter
		adapter, err := CreateChainAdapter(chainCfg, log)
		if err != nil {
			log.Error("failed to create chain adapter",
				zap.String("chain_id", chainCfg.ChainID),
				zap.Error(err),
			)
			continue
		}

		// Create block processor
		blockProcessor := processor.NewBlockProcessor(
			storage,
			storage,
			storage,
			eventBus,
			log,
			appMetrics,
		)

		// Create gap recovery
		gapRecovery := indexer.NewGapRecovery(
			adapter,
			storage,
			blockProcessor,
			eventBus,
			log,
		)

		// Create progress tracker
		progressTracker := indexer.NewProgressTracker(
			adapter,
			storage,
			storage,
			blockProcessor,
			log,
			appMetrics,
		)

		// Create block indexer configuration
		indexerConfig := &indexer.BlockIndexerConfig{
			ChainID:            chainCfg.ChainID,
			StartBlock:         chainCfg.StartBlock,
			EndBlock:           0, // Continuous indexing
			BatchSize:          chainCfg.BatchSize,
			WorkerCount:        chainCfg.Workers,
			ConfirmationBlocks: chainCfg.ConfirmationBlocks,
			PollInterval:       5 * time.Second,
			EnableGapRecovery:  true,
		}

		// Create block indexer
		blockIndexer := indexer.NewBlockIndexer(
			adapter,
			blockProcessor,
			gapRecovery,
			progressTracker,
			indexerConfig,
			log,
		)

		// Start indexer
		if err := blockIndexer.Start(ctx); err != nil {
			log.Error("failed to start indexer",
				zap.String("chain_id", chainCfg.ChainID),
				zap.Error(err),
			)
			continue
		}

		indexersMu.Lock()
		indexers = append(indexers, blockIndexer)
		indexersMu.Unlock()

		log.Info("chain indexer started",
			zap.String("chain_id", chainCfg.ChainID),
			zap.Uint64("start_block", chainCfg.StartBlock),
			zap.Int("workers", chainCfg.Workers),
		)
	}

	if len(indexers) == 0 {
		return fmt.Errorf("failed to start any indexers")
	}

	log.Info("all indexers started successfully",
		zap.Int("active_indexers", len(indexers)),
	)
	fmt.Printf("\n✅ Indexing started for %d chain(s)\n", len(indexers))
	fmt.Println("📊 Metrics available at: http://localhost:9091/metrics")
	fmt.Println("Press Ctrl+C to stop gracefully...")
	fmt.Println()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	log.Info("shutdown signal received",
		zap.String("signal", sig.String()),
	)
	fmt.Printf("\n🛑 Shutdown signal received (%s), stopping gracefully...\n", sig.String())

	// Cancel context to stop all indexers
	cancel()

	// Stop all indexers
	var wg sync.WaitGroup
	for _, idx := range indexers {
		wg.Add(1)
		go func(i *indexer.BlockIndexer) {
			defer wg.Done()
			if err := i.Stop(); err != nil {
				log.Error("error stopping indexer", zap.Error(err))
			}
		}(idx)
	}

	// Wait for all indexers to stop (with timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("all indexers stopped successfully")
		fmt.Println("✅ All indexers stopped successfully")
	case <-time.After(30 * time.Second):
		log.Warn("timeout waiting for indexers to stop")
		fmt.Println("⚠️  Timeout waiting for indexers to stop")
	}

	log.Info("indexer shutdown complete")
	fmt.Println("👋 Goodbye!")

	return nil
}

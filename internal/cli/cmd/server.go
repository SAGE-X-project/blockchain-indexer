package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/sage-x-project/blockchain-indexer/pkg/application/health"
	"github.com/sage-x-project/blockchain-indexer/pkg/application/indexer"
	"github.com/sage-x-project/blockchain-indexer/pkg/application/processor"
	"github.com/sage-x-project/blockchain-indexer/pkg/application/statistics"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/config"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/event"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/metrics"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/storage/pebble"
	"github.com/sage-x-project/blockchain-indexer/pkg/presentation/graphql/resolver"
	grpcserver "github.com/sage-x-project/blockchain-indexer/pkg/presentation/grpc/server"
	"github.com/sage-x-project/blockchain-indexer/pkg/presentation/rest"
	"github.com/sage-x-project/blockchain-indexer/pkg/presentation/rest/handler"
)

var configFile string

// NewServerCmd creates a server command
func NewServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the API server",
		Long: `Start the blockchain indexer API server with REST, GraphQL, and gRPC endpoints.

The server provides:
  - REST API on HTTP port (default: 8080)
  - GraphQL API on HTTP port (default: 8080)
  - gRPC API on gRPC port (default: 9090)
  - Prometheus metrics on metrics port (default: 9091)`,
		RunE: runServer,
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "Path to configuration file")

	return cmd
}

func runServer(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize logger
	log, err := logger.New(&logger.Config{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		Output:     cfg.Logging.Output,
		FilePath:   cfg.Logging.FilePath,
		MaxSize:    cfg.Logging.MaxSize,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge,
		Compress:   cfg.Logging.Compress,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer log.Sync()

	log.Info("starting blockchain indexer server",
		zap.String("app", cfg.App.Name),
		zap.String("environment", cfg.App.Environment),
	)

	// Initialize metrics
	var appMetrics *metrics.Metrics
	if cfg.Metrics.Enabled {
		appMetrics = metrics.New(&metrics.Config{
			Enabled: true,
			Host:    cfg.Metrics.Host,
			Port:    cfg.Metrics.Port,
			Path:    cfg.Metrics.Path,
		})

		go func() {
			addr := fmt.Sprintf("%s:%d", cfg.Metrics.Host, cfg.Metrics.Port)
			log.Info("starting metrics server", zap.String("addr", addr), zap.String("path", cfg.Metrics.Path))
			if err := http.ListenAndServe(addr, appMetrics.Handler()); err != nil {
				log.Error("metrics server error", zap.Error(err))
			}
		}()
	}

	// Initialize storage (PebbleDB)
	storagePath := cfg.Storage.Pebble.Path
	if storagePath == "" {
		storagePath = "./data" // default path
	}
	log.Info("initializing storage", zap.String("type", "pebble"), zap.String("path", storagePath))
	storage, err := pebble.NewStorage(&pebble.Config{
		Path: storagePath,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer storage.Close()

	// Create repositories
	blockRepo := storage
	transactionRepo := storage
	chainRepo := storage

	// Initialize event bus
	log.Info("initializing event bus")
	eventBusConfig := event.DefaultEventBusConfig()
	eventBusConfig.WorkerCount = 10
	eventBusConfig.QueueSize = 10000

	eventBus := event.NewEventBus(eventBusConfig, log)
	if err := eventBus.Start(); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}
	defer eventBus.Stop()

	// Initialize statistics collector
	log.Info("initializing statistics collector")
	statsCollector := statistics.NewCollector(
		storage,          // statsRepo
		storage,          // blockRepo
		storage,          // txRepo
		storage,          // chainRepo
		eventBus,
		appMetrics,
		log,
		statistics.DefaultConfig(),
	)
	if err := statsCollector.Start(ctx); err != nil {
		return fmt.Errorf("failed to start statistics collector: %w", err)
	}
	defer statsCollector.Stop()

	// Initialize gap recovery for each chain
	log.Info("initializing gap recovery")
	gapRecoveryMap := make(map[string]*indexer.GapRecovery)

	// Get all configured chains
	chains, err := chainRepo.GetAllChains(ctx)
	if err != nil {
		log.Warn("failed to get chains for gap recovery", zap.Error(err))
	} else {
		// Create block processor for gap recovery
		blockProcessor := processor.NewBlockProcessor(
			storage,
			storage,
			storage,
			eventBus,
			log,
			appMetrics,
		)

		// Create gap recovery for each chain
		for _, chain := range chains {
			// Note: In a full implementation, we'd need the actual chain adapter here
			// For now, we create gap recovery without the adapter (it will be nil)
			// This means gap recovery will only work for detection, not actual recovery
			gapRecovery := indexer.NewGapRecovery(
				nil, // adapter - would need to be created per chain
				storage,
				blockProcessor,
				eventBus,
				log,
			)
			gapRecoveryMap[chain.ChainID] = gapRecovery
		}
		log.Info("gap recovery initialized", zap.Int("chains", len(gapRecoveryMap)))
	}

	// Initialize health checker
	log.Info("initializing health checker")
	healthChecker := health.NewChecker(log, 30*time.Second)

	// Register health checks
	healthChecker.RegisterCheck("storage", health.StorageHealthCheck(chainRepo))
	healthChecker.RegisterCheck("memory", health.MemoryHealthCheck(1024)) // 1GB threshold
	healthChecker.RegisterCheck("goroutines", health.GoroutineHealthCheck(10000))

	// Start background health checking
	go healthChecker.StartPeriodicChecks(ctx)

	// Create HTTP mux
	httpMux := http.NewServeMux()

	// Basic health check endpoint
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","version":"%s"}`, cfg.App.Version)
	})

	// Detailed health check endpoint
	httpMux.HandleFunc("/health/detailed", func(w http.ResponseWriter, r *http.Request) {
		report := healthChecker.RunChecks(r.Context())
		w.Header().Set("Content-Type", "application/json")

		statusCode := http.StatusOK
		if report.Status == health.StatusDegraded {
			statusCode = http.StatusOK // Still OK, but degraded
		} else if report.Status == health.StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(report)
	})

	// Debug endpoints (pprof)
	httpMux.HandleFunc("/debug/pprof/", pprof.Index)
	httpMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	httpMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	httpMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	httpMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	httpMux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	httpMux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	httpMux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	httpMux.Handle("/debug/pprof/block", pprof.Handler("block"))
	httpMux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	httpMux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))

	// Runtime debug endpoint
	httpMux.HandleFunc("/debug/stats", func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		stats := map[string]interface{}{
			"memory": map[string]interface{}{
				"alloc_mb":      m.Alloc / 1024 / 1024,
				"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
				"sys_mb":        m.Sys / 1024 / 1024,
				"num_gc":        m.NumGC,
				"gc_pause_ms":   float64(m.PauseNs[(m.NumGC+255)%256]) / 1e6,
			},
			"runtime": map[string]interface{}{
				"goroutines": runtime.NumGoroutine(),
				"num_cpu":    runtime.NumCPU(),
				"version":    runtime.Version(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	// Initialize REST API
	if cfg.Server.HTTP.Enabled {
		log.Info("initializing REST API")
		restHandler := handler.NewHandler(blockRepo, transactionRepo, chainRepo, nil, gapRecoveryMap, statsCollector, log)
		restRouter := rest.NewRouter(restHandler, log)
		httpMux.Handle("/api/", http.StripPrefix("/api", restRouter))
		log.Info("REST API registered at /api/*")
	}

	// Initialize GraphQL API
	if cfg.Server.HTTP.Enabled {
		log.Info("initializing GraphQL API")
		graphqlResolver := resolver.NewResolver(blockRepo, transactionRepo, chainRepo, nil, statsCollector, gapRecoveryMap, eventBus, log)
		graphqlHandler := resolver.NewGraphQLHandler(graphqlResolver, true) // enable playground
		httpMux.Handle("/graphql", graphqlHandler)
		log.Info("GraphQL API registered at /graphql")
	}

	// Create HTTP server
	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port),
		Handler:           httpMux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start HTTP server
	if cfg.Server.HTTP.Enabled {
		go func() {
			log.Info("starting HTTP server",
				zap.String("host", cfg.Server.HTTP.Host),
				zap.Int("port", cfg.Server.HTTP.Port),
			)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error("HTTP server error", zap.Error(err))
			}
		}()
	}

	// Initialize and start gRPC server
	var grpcSrv *grpcserver.Server
	if cfg.Server.GRPC.Enabled {
		log.Info("initializing gRPC server")
		grpcSrv, err = grpcserver.NewServer(grpcserver.Config{
			Port:             cfg.Server.GRPC.Port,
			BlockRepo:        blockRepo,
			TransactionRepo:  transactionRepo,
			ChainRepo:        chainRepo,
			GapRecovery:      gapRecoveryMap,
			StatsCollector:   statsCollector,
			EventBus:         eventBus,
			EnableReflection: true,
		})
		if err != nil {
			return fmt.Errorf("failed to create gRPC server: %w", err)
		}

		go func() {
			log.Info("starting gRPC server", zap.Int("port", cfg.Server.GRPC.Port))
			if err := grpcSrv.Start(ctx); err != nil {
				log.Error("gRPC server error", zap.Error(err))
			}
		}()
	}

	log.Info("server started successfully",
		zap.Int("http_port", cfg.Server.HTTP.Port),
		zap.Int("grpc_port", cfg.Server.GRPC.Port),
		zap.Int("metrics_port", cfg.Metrics.Port),
	)
	log.Info("endpoints available:",
		zap.String("health", fmt.Sprintf("http://%s:%d/health", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port)),
		zap.String("rest", fmt.Sprintf("http://%s:%d/api", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port)),
		zap.String("graphql", fmt.Sprintf("http://%s:%d/graphql", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port)),
		zap.String("metrics", fmt.Sprintf("http://%s:%d%s", cfg.Metrics.Host, cfg.Metrics.Port, cfg.Metrics.Path)),
	)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Shutdown components in reverse order
	// 1. Stop gRPC server
	if grpcSrv != nil {
		log.Info("stopping gRPC server")
		if err := grpcSrv.Stop(shutdownCtx); err != nil {
			log.Error("gRPC server shutdown error", zap.Error(err))
		}
	}

	// 2. Shutdown HTTP server
	if cfg.Server.HTTP.Enabled {
		log.Info("stopping HTTP server")
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error("HTTP server shutdown error", zap.Error(err))
		}
	}

	// 3. Stop statistics collector
	log.Info("stopping statistics collector")
	if err := statsCollector.Stop(); err != nil {
		log.Error("statistics collector shutdown error", zap.Error(err))
	}

	// 4. Stop event bus
	log.Info("stopping event bus")
	if err := eventBus.Stop(); err != nil {
		log.Error("event bus shutdown error", zap.Error(err))
	}

	// 5. Close storage
	log.Info("closing storage")
	if err := storage.Close(); err != nil {
		log.Error("storage close error", zap.Error(err))
	}

	log.Info("server stopped gracefully")
	return nil
}

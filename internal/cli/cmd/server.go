package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

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

	// Create HTTP mux
	httpMux := http.NewServeMux()

	// Health check endpoint
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","version":"%s"}`, cfg.App.Version)
	})

	// Initialize REST API
	if cfg.Server.HTTP.Enabled {
		log.Info("initializing REST API")
		restHandler := handler.NewHandler(blockRepo, transactionRepo, chainRepo, nil, log)
		restRouter := rest.NewRouter(restHandler, log)
		httpMux.Handle("/api/", http.StripPrefix("/api", restRouter))
		log.Info("REST API registered at /api/*")
	}

	// Initialize GraphQL API
	if cfg.Server.HTTP.Enabled {
		log.Info("initializing GraphQL API")
		graphqlResolver := resolver.NewResolver(blockRepo, transactionRepo, chainRepo, nil, statsCollector, eventBus, log)
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

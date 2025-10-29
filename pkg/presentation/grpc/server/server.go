package server

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	indexerv1 "github.com/sage-x-project/blockchain-indexer/api/proto/indexer/v1"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/repository"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/event"
)

// Server implements the gRPC server for the indexer service
type Server struct {
	indexerv1.UnimplementedIndexerServiceServer

	grpcServer       *grpc.Server
	listener         net.Listener
	blockRepo        repository.BlockRepository
	transactionRepo  repository.TransactionRepository
	chainRepo        repository.ChainRepository
	eventBus         event.EventBus
	port             int
}

// Config holds the configuration for the gRPC server
type Config struct {
	Port             int
	BlockRepo        repository.BlockRepository
	TransactionRepo  repository.TransactionRepository
	ChainRepo        repository.ChainRepository
	EventBus         event.EventBus
	EnableReflection bool
}

// NewServer creates a new gRPC server
func NewServer(cfg Config) (*Server, error) {
	if cfg.Port == 0 {
		cfg.Port = 50051
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
		grpc.MaxSendMsgSize(10 * 1024 * 1024), // 10MB
	)

	s := &Server{
		grpcServer:      grpcServer,
		listener:        listener,
		blockRepo:       cfg.BlockRepo,
		transactionRepo: cfg.TransactionRepo,
		chainRepo:       cfg.ChainRepo,
		eventBus:        cfg.EventBus,
		port:            cfg.Port,
	}

	// Register the service
	indexerv1.RegisterIndexerServiceServer(grpcServer, s)

	// Enable reflection for tools like grpcurl
	if cfg.EnableReflection {
		reflection.Register(grpcServer)
	}

	return s, nil
}

// Start starts the gRPC server
func (s *Server) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		s.Stop(context.Background())
	}()

	if err := s.grpcServer.Serve(s.listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop(ctx context.Context) error {
	s.grpcServer.GracefulStop()
	return nil
}

// GetPort returns the port the server is listening on
func (s *Server) GetPort() int {
	return s.port
}

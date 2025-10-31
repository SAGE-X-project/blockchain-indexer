package rest

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"github.com/sage-x-project/blockchain-indexer/pkg/presentation/rest/handler"
	restmw "github.com/sage-x-project/blockchain-indexer/pkg/presentation/rest/middleware"
)

// NewRouter creates a new HTTP router with all routes configured
func NewRouter(h *handler.Handler, logger *logger.Logger) chi.Router {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(restmw.LoggerMiddleware(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", h.Health)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Blockchain Indexer API","version":"1.0.0","docs":"/api/v1/docs"}`))
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Chain routes
		r.Route("/chains", func(r chi.Router) {
			r.Get("/{chainID}", h.GetChain)
		})

		// Block routes
		r.Route("/chains/{chainID}/blocks", func(r chi.Router) {
			r.Get("/", h.ListBlocks)
			r.Get("/latest", h.GetLatestBlock)
			r.Get("/{number}", h.GetBlock)
			r.Get("/hash/{hash}", h.GetBlockByHash)
		})

		// Transaction routes
		r.Route("/chains/{chainID}/transactions", func(r chi.Router) {
			r.Get("/{hash}", h.GetTransaction)
			r.Get("/block/{number}", h.ListTransactionsByBlock)
			r.Get("/address/{address}", h.ListTransactionsByAddress)
		})

		// Progress routes
		r.Route("/chains/{chainID}/progress", func(r chi.Router) {
			r.Get("/", h.GetProgress)
		})

		// Gap routes
		r.Route("/chains/{chainID}/gaps", func(r chi.Router) {
			r.Get("/", h.GetChainGaps)
		})
	})

	return r
}

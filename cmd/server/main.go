package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/hanjo-erp/hanjo/internal/common/database"
	"github.com/hanjo-erp/hanjo/internal/common/eventbus"
	"github.com/hanjo-erp/hanjo/internal/common/middleware"
	"github.com/hanjo-erp/hanjo/pkg/config"
	"github.com/hanjo-erp/hanjo/pkg/logger"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg.App.Environment, cfg.App.Debug)

	log.Info().
		Str("name", cfg.App.Name).
		Str("version", cfg.App.Version).
		Str("environment", cfg.App.Environment).
		Msg("starting Hanjo ERP & POS server")

	// Connect to database
	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, cfg.Database, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	// Initialize event bus
	_ = eventbus.NewInMemoryEventBus()

	// ==================== Setup HTTP Router ====================
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.Recovery(log))
	r.Use(middleware.Logging(log))
	r.Use(middleware.CORS([]string{"*"}))
	r.Use(chiMiddleware.Compress(5))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"status":  "healthy",
			"service": cfg.App.Name,
			"version": cfg.App.Version,
		})
	})

	// API info
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"name":    "Hanjo ERP & POS API",
			"version": cfg.App.Version,
			"docs":    "/docs",
		})
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("address", cfg.Server.Address()).Msg("HTTP server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Wait for interrupt
	<-done
	log.Info().Msg("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced shutdown")
	}

	fmt.Println()
	log.Info().Msg("server stopped gracefully")
}

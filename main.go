package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"photoorg/internal/api"
	"photoorg/internal/config"
	"photoorg/internal/database"
	"photoorg/internal/engine"
	"photoorg/internal/sse"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed all:frontend/dist
var frontendFS embed.FS

func main() {
	// Load config
	cfg := config.Load()

	// Setup logging
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	log.Info().Str("env", cfg.Env).Int("port", cfg.APIPort).Msg("starting photoorg")

	// Initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database")
	}
	defer db.Close()

	// Create SSE broker
	broker := sse.NewBroker()

	// Create thumbnail directory
	if err := os.MkdirAll(cfg.ThumbnailDir, 0o755); err != nil {
		log.Fatal().Err(err).Msg("failed to create thumbnail directory")
	}

	// Create thumbnail cache
	thumbCache := engine.NewThumbnailCache(cfg.ThumbnailDir, 256)

	// Build router
	router := api.NewRouter(db, cfg, broker, thumbCache, frontendFS)

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.APIHost, cfg.APIPort),
		Handler: router,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}
	log.Info().Msg("server stopped")
}

package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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

// version is injected at build time via -X main.version=... (GoReleaser).
// Falls back to "dev" for local builds.
var version = "dev"

func main() {
	var dataDir string
	flag.StringVar(&dataDir, "data-dir", "", "override data directory (default: OS-standard location)")
	flag.Parse()

	// Load config (dataDir "" → OS-standard path)
	cfg := config.Load(dataDir)

	// Setup logging
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	// First-run migration: if ./data/ exists and target data dir doesn't, migrate automatically
	migrateFromCWD(cfg.DataDir)

	// Ensure data dir exists
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Fatal().Err(err).Str("dir", cfg.DataDir).Msg("failed to create data directory")
	}

	log.Info().Str("version", version).Str("env", cfg.Env).Int("port", cfg.APIPort).Str("data", cfg.DataDir).Msg("starting photoorg")

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

// migrateFromCWD moves ./data/ to the OS-standard location on first run,
// if the legacy CWD data directory exists and the new location is empty.
func migrateFromCWD(targetDir string) {
	cwdData := filepath.Join(".", "data")
	info, err := os.Stat(cwdData)
	if err != nil || !info.IsDir() {
		return // nothing to migrate
	}

	// Check if target already has data (db file present)
	targetDB := filepath.Join(targetDir, "photoorg.db")
	if _, err := os.Stat(targetDB); err == nil {
		return // already migrated
	}

	// Confirm the CWD data dir has a database before migrating
	cwdDB := filepath.Join(cwdData, "photoorg.db")
	if _, err := os.Stat(cwdDB); err != nil {
		return // nothing worth migrating
	}

	log.Info().
		Str("from", cwdData).
		Str("to", targetDir).
		Msg("migrating data directory to OS-standard location")

	if err := os.MkdirAll(filepath.Dir(targetDir), 0o755); err != nil {
		log.Warn().Err(err).Msg("migration: could not create parent dir, skipping")
		return
	}

	if err := os.Rename(cwdData, targetDir); err != nil {
		log.Warn().Err(err).
			Str("from", cwdData).
			Str("to", targetDir).
			Msg("migration: rename failed (different filesystems?), data left in place — set --data-dir=./data to continue using it")
		return
	}

	log.Info().Msg("migration complete")
}

package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"os"
	"song-library/internal/config"
	"song-library/internal/http-server/mwlogger"
	"song-library/internal/logger"
	"song-library/internal/storage/postgres"
)

func main() {
	// Config
	cfg := config.MustLoad()

	// Logger
	log := slogger.SetupLogger(cfg.Environment)

	log.Info("Starting songs library REST API server", slog.String("Environment", cfg.Environment))

	// Storage
	log.Debug("Start connect to storage")

	storage, err := postgres.New(cfg, log)
	if err != nil {
		log.Error("Error opening storage", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("Successfully connect to storage")

	_ = storage

	// Router
	router := chi.NewRouter()

	// middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
}

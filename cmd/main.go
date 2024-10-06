package main

import (
	"log/slog"
	"song-library/internal/config"
	"song-library/internal/logger"
)

func main() {
	// Config
	cfg := config.MustLoad()

	// Logger
	log := slogger.SetupLogger(cfg.Environment)

	log.Info("Starting songs library REST API server", slog.String("Environment", cfg.Environment))
}

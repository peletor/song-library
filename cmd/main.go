package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"song-library/internal/config"
	songinfo "song-library/internal/http-server/handlers/info/get"
	songdelete "song-library/internal/http-server/handlers/songs/delete"
	songsget "song-library/internal/http-server/handlers/songs/get"
	songsave "song-library/internal/http-server/handlers/songs/save"
	songtext "song-library/internal/http-server/handlers/songs/text"
	songupdate "song-library/internal/http-server/handlers/songs/update"
	"song-library/internal/http-server/mwlogger"
	"song-library/internal/logger/slogger"
	"song-library/internal/storage/postgres"
	"syscall"
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

	// Paths
	router.Get("/info", songinfo.New(log, storage))
	router.Get("/songs", songsget.New(log, storage))
	router.Post("/songs", songsave.New(log, storage, cfg.Address))
	router.Put("/songs", songupdate.New(log, storage))
	router.Delete("/songs", songdelete.New(log, storage))
	router.Get("/songs/text", songtext.New(log, cfg.Address))

	// Channel to graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Server
	log.Info("Start server", slog.String("Address", cfg.Address))

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Info("Server is not running", slog.Any("reason", err))
			stop <- syscall.SIGINT
		}
	}()

	// Graceful shutdown
	sign := <-stop

	log.Info("Stopping server", slog.String("signal", sign.String()))

	if err := server.Shutdown(context.Background()); err != nil {
		log.Error("Failed to stop server", slog.Any("error", err))
	}

	storage.Close(log)

	log.Info("Server stopped")
}

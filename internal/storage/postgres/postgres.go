package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log/slog"
	"song-library/internal/config"
)

type Storage struct {
	db *sql.DB
}

func New(cfg *config.Config, logger *slog.Logger) (*Storage, error) {
	const op = "storage.postgres.new"

	log := logger.With(slog.String("op", op),
		slog.String("host", cfg.PgHost),
		slog.String("port", cfg.PgPort),
		slog.String("user", cfg.PgUser),
		slog.String("database", cfg.PgDatabase))

	log.Debug("Connecting to postgres database")

	pgConnectString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PgUser, cfg.PgPass, cfg.PgHost, cfg.PgPort, cfg.PgDatabase)

	db, err := sql.Open("postgres", pgConnectString)
	if err != nil {
		log.Error("Unable to open database", slog.Any("error", err))
		return nil, err
	}

	if err = db.Ping(); err != nil {
		log.Error("Unable to connect to database", slog.Any("error", err))
		return nil, err
	}

	log.Debug("Successfully connected to postgres database")

	pgStorage := &Storage{db: db}

	if err = pgStorage.makeMigrations(logger); err != nil {
		log.Error("Unable to make migrations", slog.Any("error", err))
		return nil, err
	}

	return pgStorage, nil
}

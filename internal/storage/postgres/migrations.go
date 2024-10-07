package postgres

import (
	"errors"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	"log/slog"
)

func (s *Storage) makeMigrations(log *slog.Logger) error {
	const op = "storage.postgres.migrations"

	log = log.With(slog.String("op", op))

	log.Debug("Start migrations")

	driver, err := postgres.WithInstance(s.db, &postgres.Config{})
	if err != nil {
		log.Error("Unable to create database driver for migrations", slog.Any("error", err))
		return err
	}

	migrationsPath := "file://./migrations"

	migrations, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres", driver)
	if err != nil {
		log.Error("Unable to create migrate instance",
			slog.String("migrations path", migrationsPath),
			slog.Any("error", err))
		return err
	}

	if err := migrations.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Debug("No migrations to run")
		} else {
			log.Error("Unable to run migrations", slog.Any("error", err))
			return err
		}
	}

	log.Debug("Migrations successfully completed")

	return nil
}

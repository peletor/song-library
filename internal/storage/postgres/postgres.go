package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log/slog"
	"song-library/internal/config"
	"song-library/internal/storage"
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

func (s *Storage) Close(log *slog.Logger) {
	const op = "storage.postgres.Close"

	log = log.With(slog.String("op", op))

	log.Debug("Starting close database connection")

	err := s.db.Close()
	if err != nil {
		log.Error("Unable to close database connection", slog.Any("error", err))
	} else {
		log.Debug("Database connection was successfully closed")
	}
}

func (s *Storage) SaveSong(groupName string, songName string) (songID int, err error) {
	const op = "storage.postgres.SaveSong"

	groupID, err := s.getGroupID(groupName)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get group id: %w", op, err)
	}

	sqlStr := `INSERT INTO songs (name, group_id) 
				VALUES ($1, $2) 
				RETURNING id`
	err = s.db.QueryRow(sqlStr,
		songName, groupID).Scan(&songID)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to add new song: %w", op, err)
	}
	return songID, nil
}

// findGroupID find the group ID based on the group name.
func (s *Storage) findGroupID(groupName string) (groupID int, err error) {
	sqlStr := `SELECT id 
				FROM groups 
				WHERE name = ($1)`

	err = s.db.QueryRow(sqlStr, groupName).Scan(&groupID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return 0, storage.ErrGroupNotFound
	}
	return groupID, err
}

// getGroupID get the group ID based on the group name.
// If the group name is not already in the database, a new record will be added.
func (s *Storage) getGroupID(groupName string) (groupID int, err error) {
	const op = "storage.postgres.getGroupID"

	groupID, err = s.findGroupID(groupName)
	if err != nil {
		if errors.Is(err, storage.ErrGroupNotFound) {
			sqlStr := `INSERT INTO groups (name) 
						VALUES ($1)
						RETURNING id`
			err = s.db.QueryRow(sqlStr, groupName).Scan(&groupID)
			if err != nil {
				return 0, fmt.Errorf("%s: failed to add new group: %s. Err: %w", op, groupName, err)
			}

			return groupID, nil

		} else {
			return 0, fmt.Errorf("%s: failed to find group: %s. Err: %w", op, groupName, err)
		}
	}

	return groupID, nil
}

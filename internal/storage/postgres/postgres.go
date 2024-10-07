package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"log/slog"
	"song-library/internal/config"
	"song-library/internal/models"
	"song-library/internal/storage"
	"strings"
	"time"
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

func (s *Storage) SongInfo(groupName string, songName string) (songDetail models.SongDetail, err error) {
	const op = "storage.postgres.SongDetail"

	var releaseDate time.Time
	var link string
	textSlice := make([]string, 0)

	sqlStr := ` 
			SELECT s.release_date,
			       s.text,
			       s.link
			FROM songs s
			JOIN groups g ON s.group_id = s.group_id
			WHERE g.name = ($1) AND s.name = ($2)`

	err = s.db.QueryRow(sqlStr, groupName, songName).
		Scan(&releaseDate, pq.Array(&textSlice), &link)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.SongDetail{}, storage.ErrSongNotFound
		}

		return models.SongDetail{}, fmt.Errorf("%s: failed to update song: %w", op, err)
	}

	return models.SongDetail{
			ReleaseDate: dateToString(releaseDate),
			Text:        strings.Join(textSlice, "\n"),
			Link:        link},
		nil
}

func (s *Storage) SongUpdate(groupName string, songName string, songDetail models.SongDetail) error {
	const op = "storage.postgres.SongUpdate"

	releaseDate, err := time.Parse("02.01.2006", songDetail.ReleaseDate)

	textSlice := strings.Split(songDetail.Text, "\n")

	sqlStr := ` 
			UPDATE songs
			SET release_date = ($1), 
    			text = ($2),
    			link = ($3)
			FROM groups
			WHERE songs.group_id = groups.id
  				AND songs.name = ($4)
  				AND groups.name = ($5)`

	result, err := s.db.Exec(sqlStr,
		releaseDate, pq.Array(textSlice), songDetail.Link,
		songName, groupName)
	if err != nil {
		return fmt.Errorf("%s: failed to update song: %w", op, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return storage.ErrSongNotFound
	}

	return nil
}

func (s *Storage) SongDelete(groupName string, songName string) (songID int, err error) {
	const op = "storage.postgres.SongDelete"

	sqlStr := `
  			DELETE FROM songs
			WHERE name = ($1)
  			    AND group_id IN (SELECT id FROM groups WHERE name = ($2))
  			RETURNING id`

	err = s.db.QueryRow(sqlStr, songName, groupName).Scan(&songID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, storage.ErrSongNotFound
		}

		return 0, fmt.Errorf("%s: failed to delete song: %w", op, err)
	}

	// Song was successfully deleted

	// Try to delete group

	sqlStr = `
		DELETE FROM groups
  		WHERE name = ($1)`

	// If there are other songs by this group,
	// then Exec() will return a not-nil error
	// and the group will not be deleted.

	_, _ = s.db.Exec(sqlStr, groupName)

	return songID, nil
}

func (s *Storage) SongsGet(filter models.SongWithDetail, page int, limit int) (songs []models.SongWithDetail, err error) {
	const op = "storage.postgres.SongGet"

	sqlStr := ` 
			SELECT	g.name,
			    	s.name,
			    	s.release_date,
			       	s.text,
			       	s.link
			FROM songs s
			JOIN groups g ON s.group_id = g.id
			WHERE true
				`

	arguments := make([]interface{}, 0)

	if filter.GroupName != "" {
		arguments = append(arguments, filter.GroupName)
		sqlStr += fmt.Sprintf("AND g.name = ($%d) ", len(arguments))
	}

	if filter.SongName != "" {
		arguments = append(arguments, filter.SongName)
		sqlStr += fmt.Sprintf("AND s.name = ($%d) ", len(arguments))
	}

	releaseDate, err := time.Parse("02.01.2006", filter.SongDetail.ReleaseDate)
	if err == nil && !releaseDate.IsZero() {
		arguments = append(arguments, releaseDate)
		sqlStr += fmt.Sprintf("AND s.release_date = ($%d) ", len(arguments))
	}

	offset := (page - 1) * limit
	arguments = append(arguments, offset)
	sqlStr += fmt.Sprintf(`
			OFFSET ($%d) `, len(arguments))

	arguments = append(arguments, limit)
	sqlStr += fmt.Sprintf("LIMIT ($%d) ", len(arguments))

	rows, err := s.db.Query(sqlStr, arguments...)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to query songs: %w", op, err)
	}

	defer rows.Close()

	for rows.Next() {
		var song models.SongWithDetail
		var relDate time.Time
		textSlice := make([]string, 0)

		err = rows.Scan(
			&song.GroupName,
			&song.SongName,
			&relDate,
			pq.Array(&textSlice),
			&song.SongDetail.Link)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to query songs: %w", op, err)
		}

		song.SongDetail.ReleaseDate = dateToString(relDate)
		song.SongDetail.Text = strings.Join(textSlice, "\n")

		songs = append(songs, song)

	}

	return songs, nil
}

func dateToString(date time.Time) (dateString string) {
	if !date.IsZero() {
		dateString = date.Format("02.01.2006")
	}
	return dateString
}

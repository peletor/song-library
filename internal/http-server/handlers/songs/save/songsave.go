package songsave

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"song-library/internal/models"
	"song-library/internal/storage"
)

type SongSaver interface {
	SaveSong(groupName string, songName string) (songId int, err error)
}

func New(log *slog.Logger, songSaver SongSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.songs.save"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req models.Song

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Filed to decode request body", slog.Any("error", err))

			w.WriteHeader(http.StatusBadRequest)

			return
		}

		log.Info("Request body decoded", slog.Any("request", req))

		if req.Group == "" || req.Song == "" {
			log.Info("Cannot save song, group or song name is missing",
				slog.String("song", req.Song),
				slog.String("group", req.Group))

			w.WriteHeader(http.StatusBadRequest)

			return
		}

		songId, err := songSaver.SaveSong(req.Group, req.Song)
		if err != nil {
			if errors.Is(err, storage.ErrSongExists) {
				log.Info("Song already exists",
					slog.String("song", req.Song),
					slog.String("group", req.Group))

				w.WriteHeader(http.StatusAlreadyReported)

				return
			}

			log.Error("Failed to save song", slog.Any("error", err))

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		log.Info("Song successfully saved",
			slog.String("group", req.Group),
			slog.String("song", req.Song),
			slog.Int("song_id", songId),
		)

		w.WriteHeader(http.StatusCreated)
	}
}

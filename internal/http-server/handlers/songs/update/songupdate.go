package songupdate

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"song-library/internal/models"
	"song-library/internal/storage"
)

type Request struct {
	models.Song
	//	Group      string `json:"group" validate:"required"`
	//	Song       string `json:"song" validate:"required"`
	SongDetail models.SongDetail
}

type SongUpdater interface {
	SongUpdate(groupName string, songName string, songDetail models.SongDetail) error
}

func New(log *slog.Logger, songUpdater SongUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.songs.update"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Filed to decode request body", slog.Any("error", err))

			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if req.Group == "" || req.Song.Song == "" {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		log.Info("Request body decoded", slog.Any("request", req))

		err = songUpdater.SongUpdate(req.Group, req.Song.Song, req.SongDetail)
		if err != nil {
			if errors.Is(err, storage.ErrSongNotFound) {
				log.Info("Song not found",
					slog.String("group", req.Group),
					slog.String("song", req.Song.Song))

				w.WriteHeader(http.StatusNotFound)

				return
			}

			log.Error("Failed to update song", slog.Any("error", err))

			w.WriteHeader(http.StatusInternalServerError)

			return

		}

		log.Info("Song successfully updated",
			slog.String("group", req.Group),
			slog.String("song", req.Song.Song),
		)

		w.WriteHeader(http.StatusOK)
	}
}

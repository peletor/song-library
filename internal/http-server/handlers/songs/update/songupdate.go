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
	SongDetail models.SongDetail `json:"songDetail" validate:"required"`
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

		if req.GroupName == "" || req.SongName == "" {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		log.Info("Request body decoded", slog.Any("request", req))

		err = songUpdater.SongUpdate(req.GroupName, req.SongName, req.SongDetail)
		if err != nil {
			if errors.Is(err, storage.ErrSongNotFound) {
				log.Info("SongName not found",
					slog.String("group", req.GroupName),
					slog.String("song", req.SongName))

				w.WriteHeader(http.StatusNotFound)

				return
			}

			log.Error("Failed to update song", slog.Any("error", err))

			w.WriteHeader(http.StatusInternalServerError)

			return

		}

		log.Info("SongName successfully updated",
			slog.String("group", req.GroupName),
			slog.String("song", req.SongName),
		)

		w.WriteHeader(http.StatusOK)
	}
}

package songdelete

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"song-library/internal/models"
	"song-library/internal/storage"
)

type SongDeleter interface {
	DeleteSong(groupName string, songName string) (songId int, err error)
}

func New(log *slog.Logger, songDeleter SongDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.songs.delete"

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

		if req.GroupName == "" || req.SongName == "" {
			log.Info("Cannot delete song, group or song name is missing",
				slog.String("song", req.SongName),
				slog.String("group", req.GroupName))

			w.WriteHeader(http.StatusBadRequest)

			return
		}

		songId, err := songDeleter.DeleteSong(req.GroupName, req.SongName)
		if err != nil {
			if errors.Is(err, storage.ErrSongNotFound) {
				log.Info("SongName not found",
					slog.String("song", req.SongName),
					slog.String("group", req.GroupName))

				w.WriteHeader(http.StatusNoContent)

				return
			}

			log.Error("Failed to delete song", slog.Any("error", err))

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		log.Info("SongName successfully deleted",
			slog.String("group", req.GroupName),
			slog.String("song", req.SongName),
			slog.Int("song_id", songId),
		)

		w.WriteHeader(http.StatusOK)
	}
}

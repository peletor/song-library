package songinfo

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"song-library/internal/models"
	"song-library/internal/storage"
)

type SongInformer interface {
	SongInfo(groupName string, songName string) (models.SongDetail, error)
}

func New(log *slog.Logger, songInformer SongInformer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.info.get.songDetail"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		groupName := r.URL.Query().Get("group")
		songName := r.URL.Query().Get("song")

		if groupName == "" || songName == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Info("Start request GET /info",
			slog.String("group", groupName),
			slog.String("song", songName))

		songDetail, err := songInformer.SongInfo(groupName, songName)
		if err != nil {
			if errors.Is(err, storage.ErrSongNotFound) {
				log.Info("Song not found",
					slog.String("group", groupName),
					slog.String("song", songName))

				w.WriteHeader(http.StatusNoContent)

				return
			}

			log.Error("Failed to find song", slog.Any("error", err))

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		render.JSON(w, r, songDetail)
	}
}

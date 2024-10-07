package songsget

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"song-library/internal/models"
	"song-library/internal/storage"
	"strconv"
)

type SongsResponse struct {
	Songs []models.SongWithDetail `json:"songs"`
	Page  int                     `json:"page"`
	Limit int                     `json:"limit"`
	Items int                     `json:"items"` // len(songs)
}

type SongsGetter interface {
	SongsGet(filter models.SongWithDetail, page int, limit int) ([]models.SongWithDetail, error)
}

func New(log *slog.Logger, songsGetter SongsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.songs.get"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		groupName := r.URL.Query().Get("group")
		songName := r.URL.Query().Get("song")
		releaseDate := r.URL.Query().Get("date")
		page := r.URL.Query().Get("page")
		limit := r.URL.Query().Get("limit")

		log.Info("Start request GET /songs",
			slog.String("group", groupName),
			slog.String("song", songName),
			slog.String("releaseDate", releaseDate),
			slog.String("page", page),
			slog.String("limit", limit))

		pageNumber, err := strconv.Atoi(page)
		if err != nil || pageNumber < 1 {
			log.Info("Bad request: get parameter 'page' is incorrect",
				slog.String("page", page))

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		intLimit, err := strconv.Atoi(limit)
		if err != nil || intLimit < 1 {
			log.Info("Bad request: get parameter 'limit' is incorrect",
				slog.String("limit", limit))

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var filter models.SongWithDetail

		filter.GroupName = groupName
		filter.SongName = songName
		filter.SongDetail.ReleaseDate = releaseDate

		songs, err := songsGetter.SongsGet(filter, pageNumber, intLimit)
		if err != nil {
			if errors.Is(err, storage.ErrSongNotFound) {
				log.Info("SongName not found",
					slog.String("group", groupName),
					slog.String("song", songName))

				w.WriteHeader(http.StatusNoContent)

				return
			}

			log.Error("Failed to get songs", slog.Any("error", err))

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		if len(songs) == 0 {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		render.JSON(w, r, SongsResponse{
			Songs: songs,
			Page:  pageNumber,
			Limit: intLimit,
			Items: len(songs),
		})
	}
}

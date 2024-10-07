package songinfo

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"net/url"
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
			log.Info("Bad request: get parameter 'group' or 'song' is missing",
				slog.String("group", groupName),
				slog.String("song", songName))

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

var ErrSongNotFound = errors.New("song not found")

func GetInfoSongDetail(cfgHost string, groupName string, songName string) (songDetail models.SongDetail, err error) {
	queryParameters := url.Values{}
	queryParameters.Add("group", groupName)
	queryParameters.Add("song", songName)

	getURL := url.URL{
		Scheme:   "http",
		Host:     cfgHost,
		Path:     "info",
		RawQuery: queryParameters.Encode(),
	}

	resp, err := http.Get(getURL.String())
	if err != nil {
		return songDetail, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusOK {

		err := json.NewDecoder(resp.Body).Decode(&songDetail)
		if err != nil {
			return songDetail, err
		}

		return songDetail, nil

	} else if resp.StatusCode == http.StatusNoContent {

		return songDetail, ErrSongNotFound

	} else {

		return songDetail, errors.New("undefined return code from GET /info")
	}
}

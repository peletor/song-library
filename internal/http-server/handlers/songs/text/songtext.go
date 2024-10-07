package songtext

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	songinfo "song-library/internal/http-server/handlers/info/get"
	"strconv"
	"strings"
)

type Response struct {
	GroupName string `json:"group"`
	SongName  string `json:"song"`
	SongText  string `json:"text"`
	Page      int    `json:"page"`
}

func New(log *slog.Logger, cfgHost string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.songs.songText"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		groupName := r.URL.Query().Get("group")
		songName := r.URL.Query().Get("song")
		page := r.URL.Query().Get("page")

		log.Info("Start request GET /songs/text",
			slog.String("group", groupName),
			slog.String("song", songName),
			slog.String("page", page))

		pageNumber, err := strconv.Atoi(page)
		if err != nil || pageNumber < 1 {
			log.Info("Bad request: get parameter 'page' is incorrect",
				slog.String("page", page))

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if groupName == "" || songName == "" {
			log.Info("Bad request: get parameter 'group' or 'song' is missing",
				slog.String("group", groupName),
				slog.String("song", songName))

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		songDetail, err := songinfo.GetInfoSongDetail(cfgHost, groupName, songName)
		if err != nil {
			if errors.Is(err, songinfo.ErrSongNotFound) {
				log.Info("Song not found",
					slog.String("group", groupName),
					slog.String("song", songName))

				w.WriteHeader(http.StatusNoContent)
				return
			}

			log.Error("Failed GET /info",
				slog.String("group", groupName),
				slog.String("song", songName),
				slog.Any("error", err))

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Song text pagination
		verseSlice := strings.Split(songDetail.Text, "\n\n")

		if pageNumber > len(verseSlice) {
			log.Info("The song text has no this page",
				slog.String("group", groupName),
				slog.String("song", songName),
				slog.Int("page", pageNumber))

			w.WriteHeader(http.StatusNoContent)
			return
		}

		render.JSON(w, r, Response{
			GroupName: groupName,
			SongName:  songName,
			SongText:  verseSlice[pageNumber-1],
			Page:      pageNumber,
		})
	}
}

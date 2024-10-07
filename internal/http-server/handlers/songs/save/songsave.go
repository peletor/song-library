package songsave

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	songinfo "song-library/internal/http-server/handlers/info/get"
	"song-library/internal/models"
)

type SongSaver interface {
	SaveSong(groupName string, songName string) (songId int, err error)
}

func New(log *slog.Logger, songSaver SongSaver, cfgHost string) http.HandlerFunc {
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

		if req.GroupName == "" || req.SongName == "" {
			log.Info("Cannot save song, group or song name is missing",
				slog.String("song", req.SongName),
				slog.String("group", req.GroupName))

			w.WriteHeader(http.StatusBadRequest)

			return
		}

		_, err = songinfo.GetInfoSongDetail(cfgHost, req.GroupName, req.SongName)

		if err == nil {
			log.Info("Song already exists",
				slog.String("group", req.GroupName),
				slog.String("song", req.SongName))

			w.WriteHeader(http.StatusAlreadyReported)

			return

		}

		if !errors.Is(err, songinfo.ErrSongNotFound) {
			log.Error("Failed GET /info",
				slog.String("group", req.GroupName),
				slog.String("song", req.SongName),
				slog.Any("error", err))

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		log.Debug("Start to save new song",
			slog.String("group", req.GroupName),
			slog.String("song", req.SongName))

		songId, err := songSaver.SaveSong(req.GroupName, req.SongName)
		if err != nil {

			log.Error("Failed to save song", slog.Any("error", err))

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		log.Info("Song successfully saved",
			slog.String("group", req.GroupName),
			slog.String("song", req.SongName),
			slog.Int("song_id", songId),
		)

		w.WriteHeader(http.StatusCreated)
	}
}

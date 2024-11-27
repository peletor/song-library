package songdelete

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"song-library/internal/http-server/handlers/songs/delete/mocks"
	"song-library/internal/logger/slogdiscard"
	"song-library/internal/models"
	"song-library/internal/storage"
	"testing"
)

func TestSongDeleteHandler(t *testing.T) {
	cases := []struct {
		name       string
		groupName  string
		songName   string
		mockError  error
		httpStatus int
	}{
		{
			name:       "Success",
			groupName:  "test_group",
			songName:   "test_song",
			mockError:  nil,
			httpStatus: http.StatusOK,
		},
		{
			name:       "Empty group",
			groupName:  "",
			songName:   "test_song",
			mockError:  nil,
			httpStatus: http.StatusBadRequest,
		},
		{
			name:       "Empty song",
			groupName:  "test_group",
			songName:   "",
			mockError:  nil,
			httpStatus: http.StatusBadRequest,
		},
		{
			name:       "Empty group and song",
			groupName:  "",
			songName:   "",
			mockError:  nil,
			httpStatus: http.StatusBadRequest,
		},
		{
			name:       "Song not found",
			groupName:  "test_group",
			songName:   "test_song",
			mockError:  storage.ErrSongNotFound,
			httpStatus: http.StatusNoContent,
		},
		{
			name:       "Storage error",
			groupName:  "test_group",
			songName:   "test_song",
			mockError:  errors.New("internal error"),
			httpStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			songDeleterMock := mocks.NewSongDeleter(t)

			songDeleterMock.On("SongDelete", tc.groupName, tc.songName).
				Return(0, tc.mockError).Maybe()

			handler := New(slogdiscard.NewDiscardLogger(), songDeleterMock)

			song := models.Song{GroupName: tc.groupName, SongName: tc.songName}

			// make io.Reader from struct models.Song{}
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(song)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodDelete, "/songs", &buf)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, tc.httpStatus)
		})
	}
}

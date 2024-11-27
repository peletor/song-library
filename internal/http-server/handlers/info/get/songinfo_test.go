package songinfo

import (
	"errors"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"song-library/internal/http-server/handlers/info/get/mocks"
	"song-library/internal/logger/slogdiscard"
	"song-library/internal/models"
	"song-library/internal/storage"
	"testing"
)

func TestSongInfoHandler(t *testing.T) {
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

			songInformerMock := mocks.NewSongInformer(t)

			songInformerMock.On("SongInfo", tc.groupName, tc.songName).
				Return(models.SongDetail{}, tc.mockError).Maybe()

			handler := New(slogdiscard.NewDiscardLogger(), songInformerMock)

			queryParameters := url.Values{}
			queryParameters.Add("group", tc.groupName)
			queryParameters.Add("song", tc.songName)

			songURL := url.URL{Path: "/songs",
				RawQuery: queryParameters.Encode()}
			urlString := songURL.String()

			req, err := http.NewRequest(http.MethodGet, urlString, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, tc.httpStatus)
		})
	}
}

package tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"net/http"
	"song-library/internal/models"
	"testing"
)

func TestSongSave_HappyPath(t *testing.T) {
	e := httpExpect(t)

	e.POST("/songs").
		WithJSON(models.Song{
			Group: gofakeit.AppAuthor(),
			Song:  gofakeit.BookTitle(),
		}).
		Expect().Status(201)
}

func TestSongSave(t *testing.T) {
	testCases := []struct {
		name   string
		group  string
		song   string
		status int
	}{
		{
			name:   "Empty Group",
			group:  "",
			song:   "Song",
			status: http.StatusBadRequest,
		},
		{
			name:   "Empty Song",
			group:  "Group",
			song:   "",
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			e := httpExpect(t)

			e.POST("/songs").
				WithJSON(models.Song{
					Group: tc.group,
					Song:  tc.song,
				}).
				Expect().Status(tc.status)
		})
	}
}

func TestSongInfo_HappyPath(t *testing.T) {
	e := httpExpect(t)

	group := gofakeit.AppAuthor()
	song := gofakeit.BookTitle()

	e.POST("/songs").
		WithJSON(models.Song{
			Group: group,
			Song:  song,
		}).
		Expect().Status(201)

	e.GET("/info").
		WithQuery("group", group).
		WithQuery("song", song).
		Expect().Status(200).
		JSON().Object().
		ContainsKey("releaseDate").
		ContainsKey("text").
		ContainsKey("link")
}

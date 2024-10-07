package tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"net/http"
	"song-library/internal/models"
	"strings"
	"testing"
)

func TestSongSave_HappyPath(t *testing.T) {
	e := httpExpect(t)

	group := gofakeit.AppAuthor() + " " + gofakeit.Animal()
	song := gofakeit.BookTitle() + " " + gofakeit.Animal()

	e.POST("/songs").
		WithJSON(models.Song{
			GroupName: group,
			SongName:  song,
		}).
		Expect().Status(201)

	// Song already exists

	e.POST("/songs").
		WithJSON(models.Song{
			GroupName: group,
			SongName:  song,
		}).
		Expect().Status(208)
}

func TestSongSave(t *testing.T) {
	testCases := []struct {
		name   string
		group  string
		song   string
		status int
	}{
		{
			name:   "Empty GroupName",
			group:  "",
			song:   "SongName",
			status: http.StatusBadRequest,
		},
		{
			name:   "Empty SongName",
			group:  "GroupName",
			song:   "",
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			e := httpExpect(t)

			e.POST("/songs").
				WithJSON(models.Song{
					GroupName: tc.group,
					SongName:  tc.song,
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
			GroupName: group,
			SongName:  song,
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

func TestSongDelete_HappyPath(t *testing.T) {
	e := httpExpect(t)

	group := gofakeit.AppAuthor()
	song := gofakeit.BookTitle()

	e.POST("/songs").
		WithJSON(models.Song{
			GroupName: group,
			SongName:  song,
		}).
		Expect().Status(201)

	e.DELETE("/songs").
		WithJSON(models.Song{
			GroupName: group,
			SongName:  song,
		}).
		Expect().Status(200)
}

func TestSongDelete_NotFound(t *testing.T) {
	e := httpExpect(t)

	group := gofakeit.AppAuthor()
	song := gofakeit.BookTitle()

	e.DELETE("/songs").
		WithJSON(models.Song{
			GroupName: group,
			SongName:  song,
		}).
		Expect().Status(204)
}

func TestSongsGet_HappyPath(t *testing.T) {
	e := httpExpect(t)

	group := gofakeit.AppAuthor() + " " + gofakeit.Animal()
	song := gofakeit.BookTitle() + " " + gofakeit.Animal()

	page := 1
	limit := 3

	e.POST("/songs").
		WithJSON(models.Song{
			GroupName: group,
			SongName:  song,
		}).
		Expect().Status(201)

	var songObject models.SongWithDetail
	songObject.GroupName = group
	songObject.SongName = song

	songs := []models.SongWithDetail{songObject}

	// Filter by song name and group name
	e.GET("/songs").
		WithQuery("page", page).
		WithQuery("limit", limit).
		WithQuery("group", group).
		WithQuery("song", song).
		Expect().Status(200).
		JSON().Object().
		HasValue("songs", songs).
		HasValue("page", page).
		HasValue("limit", limit).
		HasValue("items", 1)

	// Filter only by song name
	e.GET("/songs").
		WithQuery("page", page).
		WithQuery("limit", limit).
		WithQuery("song", song).
		Expect().Status(200).
		JSON().Object().
		HasValue("songs", songs).
		HasValue("page", page).
		HasValue("limit", limit).
		HasValue("items", 1)

	// Filter only by group name
	e.GET("/songs").
		WithQuery("page", page).
		WithQuery("limit", limit).
		WithQuery("group", group).
		Expect().Status(200).
		JSON().Object().
		HasValue("songs", songs).
		HasValue("page", page).
		HasValue("limit", limit).
		HasValue("items", 1)
}

func TestSongTextGet_HappyPath(t *testing.T) {
	e := httpExpect(t)

	group := gofakeit.AppAuthor() + " " + gofakeit.Animal()
	song := gofakeit.BookTitle() + " " + gofakeit.Animal()

	e.POST("/songs").
		WithJSON(models.Song{
			GroupName: group,
			SongName:  song,
		}).
		Expect().Status(201)

	testCount := 10

	pages := make([]string, 0, testCount)

	for page := 1; page <= testCount; page++ {
		pageText := gofakeit.Paragraph(4, 1, 6, "\n")
		pages = append(pages, pageText)
	}

	delimiter := "\n\n"
	songText := strings.Join(pages, delimiter)

	var songObject models.SongWithDetail
	songObject.GroupName = group
	songObject.SongName = song
	songObject.SongDetail = models.SongDetail{Text: songText}

	e.PUT("/songs").
		WithJSON(songObject).
		Expect().Status(200)

	for page := 1; page <= testCount; page++ {

		e.GET("/songs/text").
			WithQuery("group", group).
			WithQuery("song", song).
			WithQuery("page", page).
			Expect().Status(200).
			JSON().Object().
			HasValue("group", group).
			HasValue("song", song).
			HasValue("page", page).
			HasValue("text", pages[page-1])
	}

	e.GET("/songs/text").
		WithQuery("group", group).
		WithQuery("song", song).
		WithQuery("page", testCount+1).
		Expect().Status(204)
}

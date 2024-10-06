package models

type Song struct {
	GroupName string `json:"group" validate:"required"`
	SongName  string `json:"song" validate:"required"`
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate" validate:"required"`
	Text        string `json:"text" validate:"required"`
	Link        string `json:"link" validate:"required"`
}

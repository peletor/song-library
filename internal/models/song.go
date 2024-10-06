package models

type Song struct {
	Group string `json:"group" validate:"required"`
	Song  string `json:"song" validate:"required"`
}

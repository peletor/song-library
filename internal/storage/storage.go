package storage

import "errors"

var (
	ErrSongExists    = errors.New("song already exists")
	ErrSongNotFound  = errors.New("song not found")
	ErrGroupNotFound = errors.New("group not found")
)

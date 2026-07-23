package playlist

import "errors"

const (
	ErrPlaylistNotExist = "Плейлист не найден или недоступен"
	ErrVideoNotExist    = "Видео не найдено"
)

// ErrVideoNotFound is the sentinel returned when a playlist references a video
// that does not exist, so handlers can map it to 404 via errors.Is.
var ErrVideoNotFound = errors.New(ErrVideoNotExist)

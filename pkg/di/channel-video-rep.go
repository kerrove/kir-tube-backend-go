package di

import "time"

// IChannelVideoRepository is the port the channel domain needs from the video
// domain. Like IVideoRepository it lives in di so channel can depend on it
// without importing video — video already imports channel, so a channel ->
// video edge would form an import cycle.
type IChannelVideoRepository interface {
	FindAllByChannelID(channelID string) ([]ChannelVideo, error)
}

// ChannelVideo is a video of a single channel, returned by
// IChannelVideoRepository. The owning channel is not part of the DTO: the
// channel domain already has that aggregate loaded and attaches it itself.
type ChannelVideo struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	PublicId string `json:"publicId"`

	Title       string `json:"title"`
	Description string `json:"description"`

	ThumbnailUrl  string `json:"thumbnailUrl"`
	VideoFileName string `json:"videoFileName"`
	MaxResolution string `json:"maxResolution"`

	ViewsCount int  `json:"viewsCount"`
	IsPublic   bool `json:"isPublic"`

	ChannelID string `json:"channelId"`
}

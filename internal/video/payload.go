package video

type ToggleLikeReq struct {
	VideoId string `json:"videoId" validate:"required"`
}

// CreateVideoInput is the data the studio (channel owner) supplies to publish a
// new video. Tags are matched-or-created by name; the public id, visibility and
// channel ownership are set by the repository, not the caller.
type CreateVideoInput struct {
	Title         string   `json:"title" validate:"required"`
	Description   string   `json:"description"`
	ThumbnailUrl  string   `json:"thumbnailUrl"`
	VideoFileName string   `json:"videoFileName"`
	MaxResolution string   `json:"maxResolution"`
	Tags          []string `json:"tags"`
}

// UpdateVideoInput is a partial update of a video. A nil pointer field is left
// untouched; a non-nil pointer overwrites the column. Tags, when provided,
// replace the video's whole tag set (matched-or-created by name).
type UpdateVideoInput struct {
	Title         *string  `json:"title"`
	Description   *string  `json:"description"`
	ThumbnailUrl  *string  `json:"thumbnailUrl"`
	VideoFileName *string  `json:"videoFileName"`
	MaxResolution *string  `json:"maxResolution"`
	IsPublic      *bool    `json:"isPublic"`
	Tags          []string `json:"tags"`
}

package playlist

type PlaylistRequest struct {
	Title         string `json:"title" validate:"required"`
	VideoPublicId string `json:"videoPublicId" validate:"required"`
}
type ToggleVideoRequest struct {
	VideoId string `json:"videoId" validate:"required"`
}
type ToggleVideoResponse struct {
	Message string `json:"string"`
}

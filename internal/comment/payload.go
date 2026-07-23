package comment

type CreateCommentReq struct {
	Text    *string `json:"text" validate:"required"`
	VideoId *string `json:"videoId" validate:"required"`
}
type UpdateCommentReq struct {
	Text *string `json:"text" validate:"required"`
}

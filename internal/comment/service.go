package comment

import (
	"go/kir-tube/internal/channel"
	"go/kir-tube/internal/user"
	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/di"
)

// CommentUser is a comment author together with their channel. user.User has no
// channel relation of its own (the FK lives on channel), so the channel is
// attached here — the read-model equivalent of Prisma's user:{ include:{ channel } }.
type CommentUser struct {
	user.User
	Channel *channel.Channel `json:"channel"`
}

// CommentWithUser is a video comment with its author (and the author's channel).
// The nested User field shadows the embedded VideoComment.User in JSON, so the
// comment carries the enriched author instead of the bare one.
type CommentWithUser struct {
	video.VideoComment
	User CommentUser `json:"user"`
}

type CommentRepositoryPort interface {
	Create(userId string, body *CreateCommentReq) (*video.VideoComment, error)
	Update(commentId, userId string, body *UpdateCommentReq) (*video.VideoComment, error)
	Delete(commentId, userId string) (bool, error)
	FindByVideo(publicId string) ([]CommentWithUser, error)
}

type CommentServiceDeps struct {
	CommentRepository CommentRepositoryPort
	VideoRepository   di.IVideoRepository
}
type CommentService struct {
	CommentRepository CommentRepositoryPort
	VideoRepository   di.IVideoRepository
}

func NewCommentService(deps *CommentServiceDeps) *CommentService {
	return &CommentService{
		CommentRepository: deps.CommentRepository,
		VideoRepository:   deps.VideoRepository,
	}
}

func (s *CommentService) CreateComment(userId string, body *CreateCommentReq) (*video.VideoComment, error) {
	return s.CommentRepository.Create(userId, body)
}
func (s *CommentService) UpdateComment(commentId, userId string, body *UpdateCommentReq) (*video.VideoComment, error) {
	return s.CommentRepository.Update(commentId, userId, body)
}
func (s *CommentService) DeleteComment(commentId, userId string) (bool, error) {
	return s.CommentRepository.Delete(commentId, userId)
}

// GetByVideo returns every comment on the video with the given publicId, newest
// first, each with its author and the author's channel.
func (s *CommentService) GetByVideo(publicId string) ([]CommentWithUser, error) {
	return s.CommentRepository.FindByVideo(publicId)
}

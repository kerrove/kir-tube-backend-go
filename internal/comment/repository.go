package comment

import (
	"go/kir-tube/internal/channel"
	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/db"
)

type CommentRepository struct {
	Database *db.Db
}

func NewCommentRepository(database *db.Db) *CommentRepository {
	return &CommentRepository{Database: database}
}

func (repo *CommentRepository) Create(userId string, body *CreateCommentReq) (*video.VideoComment, error) {
	comment := &video.VideoComment{
		UserID:  userId,
		Text:    *body.Text,
		VideoID: *body.VideoId,
	}

	if err := repo.Database.DB.Create(&comment).Error; err != nil {
		return nil, err
	}

	return comment, nil
}

func (repo *CommentRepository) Update(commentId, userId string, body *UpdateCommentReq) (*video.VideoComment, error) {
	comment, err := repo.FindById(commentId)
	if err != nil {
		return nil, err
	}
	if comment.UserID != userId {
		return nil, ErrCommentForbidden
	}

	updates := map[string]any{}
	if body.Text != nil {
		updates["text"] = *body.Text
	}

	if len(updates) > 0 {
		if err := repo.Database.DB.Model(comment).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return repo.FindById(commentId)
}
func (repo *CommentRepository) Delete(commentId, userId string) (bool, error) {
	comment, err := repo.FindById(commentId)
	if err != nil {
		return false, err
	}
	if comment.UserID != userId {
		return false, ErrCommentForbidden
	}
	if err := repo.Database.DB.Delete(&video.VideoComment{}, "id = ?", commentId).Error; err != nil {
		return false, err
	}

	return true, nil
}

func (repo *CommentRepository) FindById(id string) (*video.VideoComment, error) {
	var comment video.VideoComment

	res := repo.Database.DB.First(&comment, "id = ?", id)
	if res.Error != nil {
		return nil, res.Error
	}

	return &comment, nil
}

func (repo *CommentRepository) FindByVideo(publicId string) ([]CommentWithUser, error) {
	var comments []video.VideoComment
	err := repo.Database.DB.
		Select("video_comment.*").
		Joins("JOIN video ON video.id = video_comment.video_id").
		Where("video.public_id = ?", publicId).
		Preload("User").
		Order("video_comment.created_at DESC").
		Find(&comments).Error

	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(comments))
	for _, c := range comments {
		if c.User != nil {
			userIDs = append(userIDs, c.User.ID)
		}
	}

	channelsByUser := make(map[string]*channel.Channel, len(userIDs))
	if len(userIDs) > 0 {
		var channels []channel.Channel
		if err := repo.Database.DB.Where("user_id IN ?", userIDs).Find(&channels).Error; err != nil {
			return nil, err
		}
		for i := range channels {
			channelsByUser[channels[i].UserID] = &channels[i]
		}
	}

	result := make([]CommentWithUser, 0, len(comments))
	for i := range comments {
		c := comments[i]

		author := CommentUser{}
		if c.User != nil {
			author.User = *c.User
			author.Channel = channelsByUser[c.User.ID]
		}
		c.User = nil

		result = append(result, CommentWithUser{VideoComment: c, User: author})
	}

	return result, nil
}

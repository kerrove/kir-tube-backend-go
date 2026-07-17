package video

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/di"
)

type VideoRepository struct {
	Database *db.Db
}

func NewVideoRepository(database *db.Db) *VideoRepository {
	return &VideoRepository{Database: database}
}

func (repo *VideoRepository) FindSubscribedVideos(userID string) ([]di.SubscribedVideo, error) {
	var videos []Video

	result := repo.Database.DB.
		Joins("JOIN channel ON channel.id = video.channel_id").
		Joins("JOIN channel_subscribers ON channel_subscribers.channel_id = channel.id").
		Where("channel_subscribers.user_id = ?", userID).
		Preload("Channel.User").
		Preload("Likes").
		Order("video.created_at desc").
		Find(&videos)

	if result.Error != nil {
		return nil, result.Error
	}

	out := make([]di.SubscribedVideo, 0, len(videos))
	for _, v := range videos {
		sv := di.SubscribedVideo{
			ID:           v.ID,
			PublicId:     v.PublicId,
			Title:        v.Title,
			ThumbnailUrl: v.ThumbnailUrl,
			ViewsCount:   v.ViewsCount,
			ChannelID:    v.ChannelID,
		}

		if v.Channel != nil {
			sv.Channel = &di.SubscribedChannel{
				ID:         v.Channel.ID,
				Slug:       v.Channel.Slug,
				AvatarUrl:  v.Channel.AvatarUrl,
				IsVerified: v.Channel.IsVerified,
			}
			if v.Channel.User != nil {
				sv.Channel.User = &di.SubscribedChannelUser{
					ID:    v.Channel.User.ID,
					Name:  v.Channel.User.Name,
					Email: v.Channel.User.Email,
				}
			}
		}

		sv.Likes = make([]di.SubscribedLike, 0, len(v.Likes))
		for _, l := range v.Likes {
			sv.Likes = append(sv.Likes, di.SubscribedLike{
				ID:     l.ID,
				UserID: l.UserID,
			})
		}

		out = append(out, sv)
	}

	return out, nil
}

// FindAllByChannelID returns every video of a channel, newest first. It is the
// di.IChannelVideoRepository implementation used by the channel domain.
func (repo *VideoRepository) FindAllByChannelID(channelID string) ([]di.ChannelVideo, error) {
	var videos []Video

	res := repo.Database.DB.
		Where("channel_id = ?", channelID).
		Order("created_at desc").
		Find(&videos)

	if res.Error != nil {
		return nil, res.Error
	}

	out := make([]di.ChannelVideo, 0, len(videos))
	for _, v := range videos {
		out = append(out, di.ChannelVideo{
			ID:            v.ID,
			CreatedAt:     v.CreatedAt,
			UpdatedAt:     v.UpdatedAt,
			PublicId:      v.PublicId,
			Title:         v.Title,
			Description:   v.Description,
			ThumbnailUrl:  v.ThumbnailUrl,
			VideoFileName: v.VideoFileName,
			MaxResolution: v.MaxResolution,
			ViewsCount:    v.ViewsCount,
			IsPublic:      v.IsPublic,
			ChannelID:     v.ChannelID,
		})
	}

	return out, nil
}

func (repo *VideoRepository) FindById(id string) (*Video, error) {
	var video Video
	res := repo.Database.DB.First(&video, "id = ?", id)

	if res.Error != nil {
		return nil, res.Error
	}

	return &video, nil
}
func (repo *VideoRepository) FindByPublicId(publicId string) (*Video, error) {
	var video Video
	res := repo.Database.DB.First(&video, "public_id = ?", publicId)

	if res.Error != nil {
		return nil, res.Error
	}

	return &video, nil
}

func (repo *VideoRepository) FindRecentPublicVideos(since time.Time) ([]Video, error) {
	var videos []Video

	res := repo.Database.DB.
		Where("is_public = ? AND created_at >= ?", true, since).
		Preload("Tags").
		Preload("Likes").
		Preload("Comments").
		Preload("Channel.User").
		Find(&videos)

	if res.Error != nil {
		return nil, res.Error
	}

	return videos, nil
}

func (repo *VideoRepository) FindPublicVideosByViews(excludeIDs []string, limit int) ([]Video, error) {
	var videos []Video

	query := repo.Database.DB.Where("is_public = ?", true)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}

	res := query.
		Order("views_count desc").
		Limit(limit).
		Preload("Tags").
		Preload("Likes").
		Preload("Comments").
		Preload("Channel.User").
		Find(&videos)

	if res.Error != nil {
		return nil, res.Error
	}

	return videos, nil
}

// ToggleLike adds or removes the user's like on a video and returns the new
// state: true if the video is now liked, false if the like was removed.
func (repo *VideoRepository) ToggleLike(userId, videoId string) (bool, error) {
	if _, err := repo.FindById(videoId); err != nil {
		return false, err
	}

	var videoLike VideoLike

	res := repo.Database.DB.
		First(&videoLike, "user_id = ? AND video_id = ?", userId, videoId)

	// A like already exists -> remove it (toggle off).
	if res.Error == nil {
		if err := repo.Database.DB.Delete(&videoLike).Error; err != nil {
			return false, err
		}
		return false, nil
	}

	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return false, res.Error
	}

	// No like yet -> create one (toggle on).
	like := VideoLike{UserID: userId, VideoID: videoId}
	if err := repo.Database.DB.Create(&like).Error; err != nil {
		return false, err
	}

	return true, nil
}

package video

import (
	"strings"

	"gorm.io/gorm"
)

func (repo *VideoRepository) applyPublicSearch(query *gorm.DB, searchTerm string) *gorm.DB {
	query = query.Where("video.is_public = ?", true)

	term := strings.TrimSpace(searchTerm)
	if term == "" {
		return query
	}

	like := "%" + term + "%"
	return query.Where(
		`(
			video.title ILIKE ?
			OR video.description ILIKE ?
			OR EXISTS (
				SELECT 1 FROM video_tags vt
				JOIN video_tag t ON t.id = vt.video_tag_id
				WHERE vt.video_id = video.id AND t.name ILIKE ?
			)
			OR EXISTS (
				SELECT 1 FROM channel c
				JOIN "user" u ON u.id = c.user_id
				WHERE c.id = video.channel_id AND u.name ILIKE ?
			)
		)`,
		like, like, like, like,
	)
}

func (repo *VideoRepository) FindPublicVideos(searchTerm string, skip, limit int) ([]Video, error) {
	var videos []Video

	res := repo.applyPublicSearch(repo.Database.DB.Model(&Video{}), searchTerm).
		Preload("Channel.User").
		Preload("Tags").
		Preload("Comments").
		Preload("Likes").
		Order("video.created_at desc").
		Offset(skip).
		Limit(limit).
		Find(&videos)

	if res.Error != nil {
		return nil, res.Error
	}

	return videos, nil
}

func (repo *VideoRepository) CountPublicVideos(searchTerm string) (int64, error) {
	var count int64

	err := repo.applyPublicSearch(repo.Database.DB.Model(&Video{}), searchTerm).
		Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (repo *VideoRepository) FindByPublicIdFull(publicId string) (*Video, error) {
	var video Video

	res := repo.Database.DB.
		Preload("Channel.User").
		Preload("Tags").
		Preload("Likes").
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Order("video_comment.created_at desc")
		}).
		Preload("Comments.User").
		First(&video, "public_id = ?", publicId)

	if res.Error != nil {
		return nil, res.Error
	}

	return &video, nil
}

func (repo *VideoRepository) FindSimilarVideos(excludeVideoID, channelID string, tagIDs, titleWords []string, limit int) ([]Video, error) {
	var videos []Video

	condition, args := similarOr(tagIDs, channelID, titleWords)

	res := repo.Database.DB.Model(&Video{}).
		Where("video.id <> ?", excludeVideoID).
		Where("video.is_public = ?", true).
		Where(condition, args...).
		Preload("Channel.User").
		Order("video.created_at desc").
		Limit(limit).
		Find(&videos)

	if res.Error != nil {
		return nil, res.Error
	}

	return videos, nil
}

func (repo *VideoRepository) FindRecommendedVideos(excludeIDs, tagIDs, channelIDs, titleWords []string, skip, limit int) ([]Video, error) {
	var videos []Video

	condition, args := recommendationOr(tagIDs, channelIDs, titleWords)

	query := repo.Database.DB.Model(&Video{}).Where("video.is_public = ?", true)
	if len(excludeIDs) > 0 {
		query = query.Where("video.id NOT IN ?", excludeIDs)
	}

	res := query.
		Where(condition, args...).
		Preload("Channel.User").
		Preload("Tags").
		Order("video.created_at desc").
		Offset(skip).
		Limit(limit).
		Find(&videos)

	if res.Error != nil {
		return nil, res.Error
	}

	return videos, nil
}

func (repo *VideoRepository) CountRecommendedVideos(excludeIDs, tagIDs, channelIDs, titleWords []string) (int64, error) {
	var count int64

	condition, args := recommendationOr(tagIDs, channelIDs, titleWords)

	query := repo.Database.DB.Model(&Video{}).Where("video.is_public = ?", true)
	if len(excludeIDs) > 0 {
		query = query.Where("video.id NOT IN ?", excludeIDs)
	}

	if err := query.Where(condition, args...).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (repo *VideoRepository) FindGeneralRecommendations(excludeIDs []string, skip, limit int) ([]Video, error) {
	var videos []Video

	query := repo.Database.DB.Model(&Video{}).Where("is_public = ?", true)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}

	res := query.
		Preload("Channel.User").
		Preload("Tags").
		Order("created_at desc").
		Offset(skip).
		Limit(limit).
		Find(&videos)

	if res.Error != nil {
		return nil, res.Error
	}

	return videos, nil
}

func (repo *VideoRepository) CountAllPublicVideos() (int64, error) {
	var count int64
	if err := repo.Database.DB.Model(&Video{}).Where("is_public = ?", true).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *VideoRepository) FindByChannel(channelID string, skip, limit int) ([]Video, error) {
	var videos []Video

	res := repo.Database.DB.
		Where("channel_id = ? AND is_public = ?", channelID, true).
		Preload("Channel.User").
		Order("created_at desc").
		Offset(skip).
		Limit(limit).
		Find(&videos)

	if res.Error != nil {
		return nil, res.Error
	}

	return videos, nil
}

func (repo *VideoRepository) CountByChannel(channelID string) (int64, error) {
	var count int64
	err := repo.Database.DB.Model(&Video{}).
		Where("channel_id = ? AND is_public = ?", channelID, true).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *VideoRepository) CountChannelSubscribers(channelID string) (int64, error) {
	var count int64
	err := repo.Database.DB.Table("channel_subscribers").
		Where("channel_id = ?", channelID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *VideoRepository) IncrementViewsCount(publicId string) (*Video, error) {
	res := repo.Database.DB.Model(&Video{}).
		Where("public_id = ?", publicId).
		Update("views_count", gorm.Expr("views_count + 1"))

	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return repo.FindByPublicId(publicId)
}

func (repo *VideoRepository) FindWatchedVideoIDs(userID string) ([]string, error) {
	var ids []string
	err := repo.Database.DB.Table("watch_history").
		Where("user_id = ?", userID).
		Pluck("video_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (repo *VideoRepository) FindLikedVideoIDs(userID string) ([]string, error) {
	var ids []string
	err := repo.Database.DB.Table("video_like").
		Where("user_id = ?", userID).
		Pluck("video_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (repo *VideoRepository) FindSubscriptionChannelIDs(userID string) ([]string, error) {
	var ids []string
	err := repo.Database.DB.Table("channel_subscribers").
		Where("user_id = ?", userID).
		Pluck("channel_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (repo *VideoRepository) FindTagIDsForVideos(videoIDs []string) ([]string, error) {
	if len(videoIDs) == 0 {
		return []string{}, nil
	}
	var ids []string
	err := repo.Database.DB.Table("video_tags").
		Where("video_id IN ?", videoIDs).
		Distinct().
		Pluck("video_tag_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (repo *VideoRepository) FindTitlesForVideos(videoIDs []string) ([]string, error) {
	if len(videoIDs) == 0 {
		return []string{}, nil
	}
	var titles []string
	err := repo.Database.DB.Model(&Video{}).
		Where("id IN ?", videoIDs).
		Pluck("title", &titles).Error
	if err != nil {
		return nil, err
	}
	return titles, nil
}

func similarOr(tagIDs []string, channelID string, titleWords []string) (string, []any) {
	var parts []string
	var args []any

	if len(tagIDs) > 0 {
		parts = append(parts, "EXISTS (SELECT 1 FROM video_tags vt WHERE vt.video_id = video.id AND vt.video_tag_id IN ?)")
		args = append(args, tagIDs)
	}
	if channelID != "" {
		parts = append(parts, "video.channel_id = ?")
		args = append(args, channelID)
	}
	for _, word := range titleWords {
		parts = append(parts, "video.title ILIKE ?")
		args = append(args, "%"+word+"%")
	}
	if len(titleWords) > 0 {
		lowered := make([]string, len(titleWords))
		for i, word := range titleWords {
			lowered[i] = strings.ToLower(word)
		}
		parts = append(parts, "EXISTS (SELECT 1 FROM video_tags vt JOIN video_tag t ON t.id = vt.video_tag_id WHERE vt.video_id = video.id AND LOWER(t.name) IN ?)")
		args = append(args, lowered)
	}

	if len(parts) == 0 {
		return "1 = 0", nil
	}
	return "(" + strings.Join(parts, " OR ") + ")", args
}

func recommendationOr(tagIDs, channelIDs, titleWords []string) (string, []any) {
	var parts []string
	var args []any

	if len(tagIDs) > 0 {
		parts = append(parts, "EXISTS (SELECT 1 FROM video_tags vt WHERE vt.video_id = video.id AND vt.video_tag_id IN ?)")
		args = append(args, tagIDs)
	}
	if len(channelIDs) > 0 {
		parts = append(parts, "video.channel_id IN ?")
		args = append(args, channelIDs)
	}
	for _, word := range titleWords {
		parts = append(parts, "video.title ILIKE ?")
		args = append(args, "%"+word+"%")
	}

	if len(parts) == 0 {
		return "1 = 0", nil
	}
	return "(" + strings.Join(parts, " OR ") + ")", args
}

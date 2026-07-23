package video

import (
	"errors"
	"strings"
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

// ErrVideoNotFound is returned when an operation references a video that does
// not exist. Handlers map it to 404.
var ErrVideoNotFound = errors.New("video not found")

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

// ExistsById reports whether a video with the given id exists. It is a cheap
// primary-key lookup used for validation, not for loading the row.
func (repo *VideoRepository) ExistsById(id string) (bool, error) {
	var count int64
	err := repo.Database.DB.Model(&Video{}).
		Where("video.id = ?", id).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
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
		Where("video.is_public = ? AND video.created_at >= ?", true, since).
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

	query := repo.Database.DB.Where("video.is_public = ?", true)
	if len(excludeIDs) > 0 {
		query = query.Where("video.id NOT IN ?", excludeIDs)
	}

	res := query.
		Order("video.views_count desc").
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

func (repo *VideoRepository) ToggleLike(userId, videoId string) (bool, error) {
	exists, err := repo.ExistsById(videoId)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, ErrVideoNotFound
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

	query := repo.Database.DB.Model(&Video{}).Where("video.is_public = ?", true)
	if len(excludeIDs) > 0 {
		query = query.Where("video.id NOT IN ?", excludeIDs)
	}

	res := query.
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

func (repo *VideoRepository) CountAllPublicVideos() (int64, error) {
	var count int64
	if err := repo.Database.DB.Model(&Video{}).Where("video.is_public = ?", true).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *VideoRepository) FindByChannel(channelID string, skip, limit int) ([]Video, error) {
	var videos []Video

	res := repo.Database.DB.
		Where("video.channel_id = ? AND video.is_public = ?", channelID, true).
		Preload("Channel.User").
		Order("video.created_at desc").
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
		Where("video.channel_id = ? AND video.is_public = ?", channelID, true).
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

// --- Studio (channel-owner) CRUD -------------------------------------------

// applyChannelSearch scopes a query to a single channel's videos and, when a
// search term is given, narrows it to titles or descriptions matching it. Unlike
// applyPublicSearch it does not restrict to public videos: the owner sees drafts
// and unlisted videos too.
func (repo *VideoRepository) applyChannelSearch(query *gorm.DB, channelID, searchTerm string) *gorm.DB {
	query = query.Where("video.channel_id = ?", channelID)

	term := strings.TrimSpace(searchTerm)
	if term == "" {
		return query
	}

	like := "%" + term + "%"
	return query.Where("(video.title ILIKE ? OR video.description ILIKE ?)", like, like)
}

// FindChannelVideos lists a channel's own videos (newest first), optionally
// filtered by a search term, with the associations the studio view renders.
func (repo *VideoRepository) FindChannelVideos(channelID, searchTerm string, skip, limit int) ([]Video, error) {
	var videos []Video

	res := repo.applyChannelSearch(repo.Database.DB.Model(&Video{}), channelID, searchTerm).
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

// CountChannelVideos counts a channel's videos matching the same filter used by
// FindChannelVideos, for pagination totals.
func (repo *VideoRepository) CountChannelVideos(channelID, searchTerm string) (int64, error) {
	var count int64

	err := repo.applyChannelSearch(repo.Database.DB.Model(&Video{}), channelID, searchTerm).
		Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

// FindById loads a single video by primary key with its associations. It returns
// gorm.ErrRecordNotFound when no such video exists.
func (repo *VideoRepository) FindById(id string) (*Video, error) {
	var video Video

	res := repo.Database.DB.
		Preload("Channel").
		Preload("Tags").
		Preload("Comments").
		Preload("Likes").
		First(&video, "video.id = ?", id)

	if res.Error != nil {
		return nil, res.Error
	}

	return &video, nil
}

// resolveTags matches each name to an existing tag or creates it, returning the
// persisted tags. It is the Go equivalent of Prisma's connectOrCreate.
func (repo *VideoRepository) resolveTags(names []string) ([]VideoTag, error) {
	if len(names) == 0 {
		return nil, nil
	}

	tags := make([]VideoTag, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		var tag VideoTag
		if err := repo.Database.DB.
			Where(VideoTag{Name: name}).
			FirstOrCreate(&tag).Error; err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// Create publishes a new video for the given channel. The public id is generated
// here and the video is made public on creation, mirroring the NestJS backend.
func (repo *VideoRepository) Create(channelID string, input CreateVideoInput) (*Video, error) {
	tags, err := repo.resolveTags(input.Tags)
	if err != nil {
		return nil, err
	}

	publicID, err := newPublicID()
	if err != nil {
		return nil, err
	}

	maxResolution := input.MaxResolution
	if maxResolution == "" {
		maxResolution = "1080p"
	}

	video := Video{
		PublicId:      publicID,
		Title:         input.Title,
		Description:   input.Description,
		ThumbnailUrl:  input.ThumbnailUrl,
		VideoFileName: input.VideoFileName,
		MaxResolution: maxResolution,
		IsPublic:      true,
		ChannelID:     channelID,
	}

	if err := repo.Database.DB.Create(&video).Error; err != nil {
		return nil, err
	}

	// Associate tags after the insert so GORM only writes join rows and never
	// tries to re-create the (already persisted) tags.
	if len(tags) > 0 {
		if err := repo.Database.DB.Model(&video).Association("Tags").Append(tags); err != nil {
			return nil, err
		}
		video.Tags = tags
	}

	return &video, nil
}

// Update applies a partial change to a video and, when tags are supplied,
// replaces its whole tag set. It returns gorm.ErrRecordNotFound if the video is
// missing. Matching the NestJS backend, an update always rewrites the tag set:
// omitting tags clears them.
func (repo *VideoRepository) Update(id string, input UpdateVideoInput) (*Video, error) {
	video, err := repo.FindById(id)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.ThumbnailUrl != nil {
		updates["thumbnail_url"] = *input.ThumbnailUrl
	}
	if input.VideoFileName != nil {
		updates["video_file_name"] = *input.VideoFileName
	}
	if input.MaxResolution != nil {
		updates["max_resolution"] = *input.MaxResolution
	}
	if input.IsPublic != nil {
		updates["is_public"] = *input.IsPublic
	}

	if len(updates) > 0 {
		if err := repo.Database.DB.Model(video).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	tags, err := repo.resolveTags(input.Tags)
	if err != nil {
		return nil, err
	}
	if err := repo.Database.DB.Model(video).Association("Tags").Replace(tags); err != nil {
		return nil, err
	}

	return repo.FindById(id)
}

// Delete removes a video by id, returning the row as it was just before removal.
// It returns gorm.ErrRecordNotFound if the video does not exist.
func (repo *VideoRepository) Delete(id string) (*Video, error) {
	video, err := repo.FindById(id)
	if err != nil {
		return nil, err
	}

	if err := repo.Database.DB.Delete(&Video{}, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return video, nil
}

// appendTagClause matches videos carrying any of the given tag IDs.
func appendTagClause(parts []string, args []any, tagIDs []string) ([]string, []any) {
	if len(tagIDs) > 0 {
		parts = append(parts, "EXISTS (SELECT 1 FROM video_tags vt WHERE vt.video_id = video.id AND vt.video_tag_id IN ?)")
		args = append(args, tagIDs)
	}
	return parts, args
}

// appendTitleClauses matches videos whose title contains any of the words.
func appendTitleClauses(parts []string, args []any, titleWords []string) ([]string, []any) {
	for _, word := range titleWords {
		parts = append(parts, "video.title ILIKE ?")
		args = append(args, "%"+word+"%")
	}
	return parts, args
}

// joinOr OR-joins the accumulated clauses. With no clauses it returns a
// never-matching condition so callers fetch nothing rather than everything.
func joinOr(parts []string, args []any) (string, []any) {
	if len(parts) == 0 {
		return "1 = 0", nil
	}
	return "(" + strings.Join(parts, " OR ") + ")", args
}

func similarOr(tagIDs []string, channelID string, titleWords []string) (string, []any) {
	var parts []string
	var args []any

	parts, args = appendTagClause(parts, args, tagIDs)
	if channelID != "" {
		parts = append(parts, "video.channel_id = ?")
		args = append(args, channelID)
	}
	parts, args = appendTitleClauses(parts, args, titleWords)
	if len(titleWords) > 0 {
		lowered := make([]string, len(titleWords))
		for i, word := range titleWords {
			lowered[i] = strings.ToLower(word)
		}
		parts = append(parts, "EXISTS (SELECT 1 FROM video_tags vt JOIN video_tag t ON t.id = vt.video_tag_id WHERE vt.video_id = video.id AND LOWER(t.name) IN ?)")
		args = append(args, lowered)
	}

	return joinOr(parts, args)
}

func recommendationOr(tagIDs, channelIDs, titleWords []string) (string, []any) {
	var parts []string
	var args []any

	parts, args = appendTagClause(parts, args, tagIDs)
	if len(channelIDs) > 0 {
		parts = append(parts, "video.channel_id IN ?")
		args = append(args, channelIDs)
	}
	parts, args = appendTitleClauses(parts, args, titleWords)

	return joinOr(parts, args)
}

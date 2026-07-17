package video

import "time"

type IVideoRepository interface {
	ToggleLike(userId, videoId string) (bool, error)

	FindByPublicIdFull(publicId string) (*Video, error)
	IncrementViewsCount(publicId string) (*Video, error)

	FindPublicVideos(searchTerm string, skip, limit int) ([]Video, error)
	CountPublicVideos(searchTerm string) (int64, error)

	FindSimilarVideos(excludeVideoID, channelID string, tagIDs, titleWords []string, limit int) ([]Video, error)

	FindRecentPublicVideos(since time.Time) ([]Video, error)
	FindPublicVideosByViews(excludeIDs []string, limit int) ([]Video, error)

	FindRecommendedVideos(excludeIDs, tagIDs, channelIDs, titleWords []string, skip, limit int) ([]Video, error)
	CountRecommendedVideos(excludeIDs, tagIDs, channelIDs, titleWords []string) (int64, error)
	FindGeneralRecommendations(excludeIDs []string, skip, limit int) ([]Video, error)
	CountAllPublicVideos() (int64, error)

	FindByChannel(channelID string, skip, limit int) ([]Video, error)
	CountByChannel(channelID string) (int64, error)
	CountChannelSubscribers(channelID string) (int64, error)

	FindWatchedVideoIDs(userID string) ([]string, error)
	FindLikedVideoIDs(userID string) ([]string, error)
	FindSubscriptionChannelIDs(userID string) ([]string, error)
	FindTagIDsForVideos(videoIDs []string) ([]string, error)
	FindTitlesForVideos(videoIDs []string) ([]string, error)
}

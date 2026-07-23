package video

import (
	"time"

	"go/kir-tube/pkg/di"
)

// IVideoRepository is the full contract of VideoRepository. VideoRepository is
// a single type whose methods live in repository.go; this interface is the one
// place that lists everything it offers.
//
// The cross-context ports di.IVideoRepository and di.IChannelVideoRepository
// stay separate on purpose: user, channel and playlist depend on those to look
// videos up without importing this package, which would form an import cycle.
// VideoRepository satisfies all three.
type IVideoRepository interface {
	FindSubscribedVideos(userID string) ([]di.SubscribedVideo, error)
	FindLikedVideos(userID string) ([]di.SubscribedVideo, error)
	FindSubscriptions(userID string) ([]di.SubscriptionChannel, error)
	FindAllByChannelID(channelID string) ([]di.ChannelVideo, error)

	ExistsById(id string) (bool, error)
	FindByPublicId(publicId string) (*Video, error)
	FindIdByPublicId(publicId string) (id string, found bool, err error)
	FindByPublicIdFull(publicId string) (*Video, error)
	IncrementViewsCount(publicId string) (*Video, error)

	ToggleLike(userId, videoId string) (bool, error)

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

	FindChannelVideos(channelID, searchTerm string, skip, limit int) ([]Video, error)
	CountChannelVideos(channelID, searchTerm string) (int64, error)
	FindById(id string) (*Video, error)
	Create(channelID string, input CreateVideoInput) (*Video, error)
	Update(id string, input UpdateVideoInput) (*Video, error)
	Delete(id string) (*Video, error)

	FindWatchedVideoIDs(userID string) ([]string, error)
	FindLikedVideoIDs(userID string) ([]string, error)
	FindSubscriptionChannelIDs(userID string) ([]string, error)
	FindTagIDsForVideos(videoIDs []string) ([]string, error)
	FindTitlesForVideos(videoIDs []string) ([]string, error)
}

// Compile-time assertions that VideoRepository implements every port.
var (
	_ IVideoRepository            = (*VideoRepository)(nil)
	_ di.IVideoRepository         = (*VideoRepository)(nil)
	_ di.IChannelVideoRepository  = (*VideoRepository)(nil)
	_ di.IPlaylistVideoRepository = (*VideoRepository)(nil)
)

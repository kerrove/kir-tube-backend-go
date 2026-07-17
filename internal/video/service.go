package video

import (
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

type VideoServiceDeps struct {
	VideoRepository IVideoRepository
}

type TrendingVideo struct {
	Video
	EngagementScore float64 `json:"engagementScore"`
}

type PaginatedVideos struct {
	Videos     []Video `json:"videos"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	TotalCount int64   `json:"totalCount"`
	TotalPages int     `json:"totalPages"`
}

type VideoWithSimilar struct {
	Video
	SubscribersCount int64   `json:"subscribersCount"`
	SimilarVideos    []Video `json:"similarVideos"`
}

type VideoService struct {
	VideoRepository IVideoRepository
}

func NewVideoService(deps *VideoServiceDeps) *VideoService {
	return &VideoService{
		VideoRepository: deps.VideoRepository,
	}
}

func (s *VideoService) ToggleLike(userId, videoId string) (bool, error) {
	return s.VideoRepository.ToggleLike(userId, videoId)
}

func (s *VideoService) GetAll(searchTerm string, page, limit int) (*PaginatedVideos, error) {
	page, limit = normalizePage(page, limit, 10)
	skip := (page - 1) * limit

	videos, err := s.VideoRepository.FindPublicVideos(searchTerm, skip, limit)
	if err != nil {
		return nil, err
	}

	totalCount, err := s.VideoRepository.CountPublicVideos(searchTerm)
	if err != nil {
		return nil, err
	}

	return &PaginatedVideos{
		Videos:     videos,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: pagesCeil(totalCount, limit),
	}, nil
}

func (s *VideoService) GetVideoByPublicId(publicId string) (*VideoWithSimilar, error) {
	video, err := s.findVideoByPublicId(publicId)
	if err != nil {
		return nil, err
	}

	similarVideos, err := s.getSimilarVideos(video)
	if err != nil {
		return nil, err
	}

	subscribersCount, err := s.VideoRepository.CountChannelSubscribers(video.ChannelID)
	if err != nil {
		return nil, err
	}

	return &VideoWithSimilar{
		Video:            *video,
		SubscribersCount: subscribersCount,
		SimilarVideos:    similarVideos,
	}, nil
}

func (s *VideoService) findVideoByPublicId(publicId string) (*Video, error) {
	return s.VideoRepository.FindByPublicIdFull(publicId)
}

func (s *VideoService) GetRecommendations(userId string, page, limit int, excludeIds []string) (*PaginatedVideos, error) {
	if userId != "" {
		return s.getPersonalizedRecommendations(userId, page, limit, excludeIds)
	}
	return s.getGeneralRecommendations(page, limit, excludeIds)
}

func (s *VideoService) getPersonalizedRecommendations(userId string, page, limit int, excludeIds []string) (*PaginatedVideos, error) {
	page, limit = normalizePage(page, limit, 30)
	skip := (page - 1) * limit

	var (
		watched                []string
		liked                  []string
		subscriptionChannelIDs []string
	)

	interactions := new(errgroup.Group)
	interactions.Go(func() (err error) {
		watched, err = s.VideoRepository.FindWatchedVideoIDs(userId)
		return err
	})
	interactions.Go(func() (err error) {
		liked, err = s.VideoRepository.FindLikedVideoIDs(userId)
		return err
	})
	interactions.Go(func() (err error) {
		subscriptionChannelIDs, err = s.VideoRepository.FindSubscriptionChannelIDs(userId)
		return err
	})

	if err := interactions.Wait(); err != nil {
		return nil, err
	}

	interactedVideoIDs := unique(append(append([]string{}, watched...), liked...))

	var (
		tagIDs     []string
		titleWords []string
	)
	signals := new(errgroup.Group)
	signals.Go(func() (err error) {
		tagIDs, err = s.VideoRepository.FindTagIDsForVideos(interactedVideoIDs)
		return err
	})
	signals.Go(func() (err error) {
		titleWords, err = s.getTitleWordsFromVideos(interactedVideoIDs)
		return err
	})
	if err := signals.Wait(); err != nil {
		return nil, err
	}

	totalCount, err := s.VideoRepository.CountRecommendedVideos(excludeIds, tagIDs, subscriptionChannelIDs, titleWords)

	if err != nil {
		return nil, err
	}

	excludeForFind := unique(append(append([]string{}, interactedVideoIDs...), excludeIds...))
	recommended, err := s.VideoRepository.FindRecommendedVideos(excludeForFind, tagIDs, subscriptionChannelIDs, titleWords, skip, limit)

	if err != nil {
		return nil, err
	}

	if len(recommended) < limit {
		additionalLimit := limit - len(recommended)
		additionalExclude := append(append([]string{}, excludeIds...), videoIDs(recommended)...)

		additional, err := s.getGeneralRecommendations(1, additionalLimit, additionalExclude)
		if err != nil {
			return nil, err
		}
		recommended = append(recommended, additional.Videos...)
	}

	return &PaginatedVideos{
		Videos:     recommended,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: pagesCeil(totalCount, limit),
	}, nil
}

func (s *VideoService) getGeneralRecommendations(page, limit int, excludeIds []string) (*PaginatedVideos, error) {
	page, limit = normalizePage(page, limit, 30)
	skip := (page - 1) * limit

	totalCount, err := s.VideoRepository.CountAllPublicVideos()
	if err != nil {
		return nil, err
	}

	videos, err := s.VideoRepository.FindGeneralRecommendations(excludeIds, skip, limit)
	if err != nil {
		return nil, err
	}

	return &PaginatedVideos{
		Videos:     videos,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: pagesCeil(totalCount, limit),
	}, nil
}

func (s *VideoService) getTitleWordsFromVideos(videoIDs []string) ([]string, error) {
	titles, err := s.VideoRepository.FindTitlesForVideos(videoIDs)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var words []string
	for _, title := range titles {
		for _, word := range s.extractTitleWords(title) {
			if _, ok := seen[word]; ok {
				continue
			}
			seen[word] = struct{}{}
			words = append(words, word)
		}
	}
	return words, nil
}

func (s *VideoService) getSimilarVideos(video *Video) ([]Video, error) {
	tagIDs := make([]string, 0, len(video.Tags))
	for _, tag := range video.Tags {
		tagIDs = append(tagIDs, tag.ID)
	}

	titleWords := s.extractTitleWords(video.Title)

	similar, err := s.VideoRepository.FindSimilarVideos(video.ID, video.ChannelID, tagIDs, titleWords, 6)
	if err != nil {
		return nil, err
	}

	rand.Shuffle(len(similar), func(i, j int) {
		similar[i], similar[j] = similar[j], similar[i]
	})

	return similar, nil
}

func (s *VideoService) extractTitleWords(title string) []string {
	words := strings.Fields(strings.ToLower(title))
	out := make([]string, 0, len(words))
	for _, word := range words {
		if len(word) > 2 {
			out = append(out, word)
		}
	}
	return out
}

func (s *VideoService) ByChannel(channelId string, page, limit int) (*PaginatedVideos, error) {
	page, limit = normalizePage(page, limit, 10)
	skip := (page - 1) * limit

	videos, err := s.VideoRepository.FindByChannel(channelId, skip, limit)
	if err != nil {
		return nil, err
	}

	totalCount, err := s.VideoRepository.CountByChannel(channelId)
	if err != nil {
		return nil, err
	}

	return &PaginatedVideos{
		Videos:     videos,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: pagesCeil(totalCount, limit),
	}, nil
}
func (s *VideoService) GetTrendingVideos() ([]TrendingVideo, error) {
	const trendingLimit = 6

	oneWeekAgo := time.Now().AddDate(0, 0, -7)

	recentVideos, err := s.VideoRepository.FindRecentPublicVideos(oneWeekAgo)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	trending := make([]TrendingVideo, 0, len(recentVideos))
	for _, v := range recentVideos {
		hoursSincePublished := now.Sub(v.CreatedAt).Hours()
		engagementScore := (float64(v.ViewsCount)*1 +
			float64(len(v.Likes))*2 +
			float64(len(v.Comments))*3) /
			math.Pow(hoursSincePublished+2, 1.5)

		trending = append(trending, TrendingVideo{
			Video:           v,
			EngagementScore: engagementScore,
		})
	}

	sort.Slice(trending, func(i, j int) bool {
		return trending[i].EngagementScore > trending[j].EngagementScore
	})

	if len(trending) > trendingLimit {
		trending = trending[:trendingLimit]
	}

	if len(trending) < trendingLimit {
		needed := trendingLimit - len(trending)

		excludeIDs := make([]string, len(trending))
		for i, t := range trending {
			excludeIDs[i] = t.ID
		}

		additionalVideos, err := s.VideoRepository.FindPublicVideosByViews(excludeIDs, needed)
		if err != nil {
			return nil, err
		}

		for _, v := range additionalVideos {
			trending = append(trending, TrendingVideo{
				Video:           v,
				EngagementScore: 0,
			})
		}
	}

	return trending, nil
}

func (s *VideoService) UpdateViewsCount(publicId string) (*Video, error) {
	return s.VideoRepository.IncrementViewsCount(publicId)
}

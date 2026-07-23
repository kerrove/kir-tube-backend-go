package studio

import (
	"errors"
	"math"

	"gorm.io/gorm"

	"go/kir-tube/internal/video"
)

type VideoRepository interface {
	FindChannelVideos(channelID, searchTerm string, skip, limit int) ([]video.Video, error)
	CountChannelVideos(channelID, searchTerm string) (int64, error)
	FindById(id string) (*video.Video, error)
	Create(channelID string, input video.CreateVideoInput) (*video.Video, error)
	Update(id string, input video.UpdateVideoInput) (*video.Video, error)
	Delete(id string) (*video.Video, error)
}

type StudioServiceDeps struct {
	VideoRepository VideoRepository
}
type StudioService struct {
	VideoRepository VideoRepository
}

func NewStudioService(deps *StudioServiceDeps) *StudioService {
	return &StudioService{
		VideoRepository: deps.VideoRepository,
	}
}

const defaultStudioLimit = 6

func (s *StudioService) GetAll(channelID, searchTerm string, page, limit int) (*video.PaginatedVideos, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultStudioLimit
	}
	skip := (page - 1) * limit

	videos, err := s.VideoRepository.FindChannelVideos(channelID, searchTerm, skip, limit)
	if err != nil {
		return nil, err
	}

	totalCount, err := s.VideoRepository.CountChannelVideos(channelID, searchTerm)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(limit)))
	}

	return &video.PaginatedVideos{
		Videos:     videos,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}, nil
}

func (s *StudioService) ById(id string) (*video.Video, error) {
	found, err := s.VideoRepository.FindById(id)
	if err != nil {
		return nil, notFoundOr(err)
	}
	return found, nil
}

func (s *StudioService) Create(channelID string, input video.CreateVideoInput) (string, error) {
	created, err := s.VideoRepository.Create(channelID, input)
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

func (s *StudioService) Update(id string, input video.UpdateVideoInput) (*video.Video, error) {
	updated, err := s.VideoRepository.Update(id, input)
	if err != nil {
		return nil, notFoundOr(err)
	}
	return updated, nil
}

func (s *StudioService) Delete(id string) (*video.Video, error) {
	deleted, err := s.VideoRepository.Delete(id)
	if err != nil {
		return nil, notFoundOr(err)
	}
	return deleted, nil
}

func notFoundOr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return video.ErrVideoNotFound
	}
	return err
}

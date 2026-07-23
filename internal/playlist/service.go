package playlist

import (
	"errors"

	"gorm.io/gorm"
)

type PlaylistServiceDeps struct {
	PlaylistRepository *PlaylistRepository
}

type PlaylistService struct {
	PlaylistRepository *PlaylistRepository
}

func NewPlaylistService(deps *PlaylistServiceDeps) *PlaylistService {
	return &PlaylistService{
		PlaylistRepository: deps.PlaylistRepository,
	}
}

func (s *PlaylistService) GetUserPlaylist(userId string) (*[]Playlist, error) {
	playlist, err := s.PlaylistRepository.FindByUserId(userId)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrPlaylistNotExist)
		}
		return nil, err
	}

	return playlist, nil
}

func (s *PlaylistService) GetPlaylistById(playlistId string) (*Playlist, error) {
	playlist, err := s.PlaylistRepository.FindById(playlistId)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrPlaylistNotExist)
		}
		return nil, err
	}

	return playlist, nil
}
func (s *PlaylistService) ToggleVideo(userId, playlistId, videoId string) (*ToggleVideoResponse, error) {
	return s.PlaylistRepository.ToggleVideo(playlistId, videoId, userId)
}
func (s *PlaylistService) Create(userId string, body *PlaylistRequest) {

}

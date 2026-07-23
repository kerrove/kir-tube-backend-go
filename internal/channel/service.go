package channel

import (
	"errors"

	"gorm.io/gorm"
)

type ToggleSubscribeRes struct {
	Message      string `json:"message"`
	IsSubscribed bool   `json:"isSubscribed"`
}

// ChannelRepositoryPort is the persistence contract the channel service depends
// on. *ChannelRepository satisfies it.
type ChannelRepositoryPort interface {
	FindAll() *[]Channel
	FindBySlug(slug string) (*ChannelDetails, error)
	ToggleSubscribe(channelSlug, userId string) (bool, error)
}

type ChannelServiceDeps struct {
	ChannelRepository ChannelRepositoryPort
}

type ChannelService struct {
	ChannelRepository ChannelRepositoryPort
}

func NewChannelService(deps *ChannelServiceDeps) *ChannelService {
	return &ChannelService{
		ChannelRepository: deps.ChannelRepository,
	}
}

func (s *ChannelService) GetAll() *[]Channel {
	return s.ChannelRepository.FindAll()
}

func (s *ChannelService) GetBySlug(slug string) (*ChannelDetails, error) {
	return s.ChannelRepository.FindBySlug(slug)
}

func (s *ChannelService) ToggleSubscribe(slug, userId string) (*ToggleSubscribeRes, error) {
	isSubscribed, err := s.ChannelRepository.ToggleSubscribe(slug, userId)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrChannelNotExist)
		}
		return nil, err
	}

	if isSubscribed {
		return &ToggleSubscribeRes{Message: "Subscribed successfully", IsSubscribed: true}, nil
	}

	return &ToggleSubscribeRes{Message: "Unsubscribed successfully", IsSubscribed: false}, nil
}

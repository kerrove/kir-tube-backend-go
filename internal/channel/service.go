package channel

import (
	"errors"

	"gorm.io/gorm"

	"go/kir-tube/configs"
)

type ToggleSubscribeRes struct {
	Message      string `json:"message"`
	IsSubscribed bool   `json:"isSubscribed"`
}

type ChannelServiceDeps struct {
	Config            *configs.Config
	ChannelRepository *ChannelRepository
}

type ChannelService struct {
	Config            *configs.Config
	ChannelRepository *ChannelRepository
}

func NewChannelService(deps *ChannelServiceDeps) *ChannelService {
	return &ChannelService{
		Config:            deps.Config,
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

package channel

import (
	"slices"

	"go/kir-tube/internal/user"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/gormx"
)

type ChannelRepository struct {
	Database        *db.Db
	VideoRepository di.IChannelVideoRepository
}

func NewChannelRepository(database *db.Db, videoRepository di.IChannelVideoRepository) *ChannelRepository {
	return &ChannelRepository{Database: database, VideoRepository: videoRepository}
}

func (repo *ChannelRepository) FindAll() *[]Channel {
	var channels []Channel

	repo.Database.Table("channel").
		Order("created_at desc").
		Scan(&channels)

	return &channels
}

func (repo *ChannelRepository) FindBySlug(slug string) (*ChannelDetails, error) {
	var channel Channel

	res := repo.Database.DB.
		Preload("User").
		Preload("Subscribers").
		First(&channel, "slug = ?", slug)

	if res.Error != nil {
		return nil, res.Error
	}

	videos, err := repo.VideoRepository.FindAllByChannelID(channel.ID)

	if err != nil {
		return nil, err
	}

	owner := channel
	owner.Subscribers = nil

	details := &ChannelDetails{
		Channel: channel,
		Videos:  make([]ChannelVideo, 0, len(videos)),
	}

	for _, video := range videos {
		details.Videos = append(details.Videos, ChannelVideo{
			ChannelVideo: video,
			Channel:      &owner,
		})
	}

	return details, nil
}

func (repo *ChannelRepository) FindById(id string) (*Channel, error) {
	var channel Channel

	err := repo.Database.DB.
		Preload("User").
		Preload("Subscribers").
		First(&channel, "id = ?", id).
		Error

	if err != nil {
		return nil, err
	}

	return &channel, nil
}

// findEntityBySlug loads the channel entity with its subscribers. It returns a
// *Channel rather than the *ChannelDetails read model, because only the entity
// maps to a table: handing ChannelDetails to GORM's Model() makes it try to
// resolve the Videos field as a relation and fail.
func (repo *ChannelRepository) findEntityBySlug(slug string) (*Channel, error) {
	var channel Channel

	err := repo.Database.DB.
		Preload("Subscribers").
		First(&channel, "slug = ?", slug).
		Error

	if err != nil {
		return nil, err
	}

	return &channel, nil
}

func (repo *ChannelRepository) ToggleSubscribe(channelSlug, userId string) (bool, error) {
	channel, err := repo.findEntityBySlug(channelSlug)
	if err != nil {
		return false, err
	}

	isSubscribed := slices.ContainsFunc(channel.Subscribers, func(subscriber user.User) bool {
		return subscriber.ID == userId
	})

	subscriber := user.User{
		TimestampedBase: gormx.TimestampedBase{
			Base: gormx.Base{Identifier: gormx.Identifier{ID: userId}},
		},
	}

	subscribers := repo.Database.DB.Model(channel).Association("Subscribers")

	if isSubscribed {
		if err := subscribers.Delete(&subscriber); err != nil {
			return false, err
		}

		return false, nil
	}

	if err := subscribers.Append(&subscriber); err != nil {
		return false, err
	}

	return true, nil
}

package channel

import (
	"go/kir-tube/internal/user"
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/gormx"

	"gorm.io/gorm"
)

// Channel is a user's public presence that hosts videos and gathers
// subscribers. Mirrors the Prisma "channel" model.
type Channel struct {
	gormx.TimestampedBase

	Slug        string  `json:"slug" gorm:"uniqueIndex;not null"`
	Description *string `json:"description"`

	IsVerified bool `json:"isVerified" gorm:"not null;default:false"`

	AvatarUrl *string `json:"avatarUrl"`
	BannerUrl *string `json:"bannerUrl"`

	// Owner: one channel per user (unique FK), cascade-deleted with the user.
	UserID string     `json:"userId" gorm:"uniqueIndex;not null"`
	User   *user.User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	// Subscribers is the many-to-many opposite of User.subscriptions in Prisma.
	Subscribers []user.User `json:"subscribers" gorm:"many2many:channel_subscribers"`
}

// TableName keeps the table name aligned with the Prisma @@map("channel").
func (Channel) TableName() string { return "channel" }

// AfterFind guarantees Subscribers serializes as an array ([]) instead of null
// when the association was not preloaded, keeping the JSON shape stable wherever
// a channel is read — the channel, playlist and studio read models all embed it.
func (c *Channel) AfterFind(*gorm.DB) error {
	if c.Subscribers == nil {
		c.Subscribers = []user.User{}
	}
	return nil
}

// ChannelDetails is a channel read model: the channel itself with its owner and
// subscribers, plus its videos. It is not a table — the videos come from the
// video domain through di.IChannelVideoRepository.
type ChannelDetails struct {
	Channel

	Videos []ChannelVideo `json:"videos"`
}

// ChannelVideo is a video of the channel with its owning channel attached, so
// clients can read video.channel.user without a second request.
type ChannelVideo struct {
	di.ChannelVideo

	Channel *Channel `json:"channel,omitempty"`
}

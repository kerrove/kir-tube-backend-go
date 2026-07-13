package video

import (
	"go/kir-tube/internal/channel"
	"go/kir-tube/internal/user"
	"go/kir-tube/pkg/gormx"
)

// Video is an uploaded clip belonging to a channel. Mirrors the Prisma
// "video" model.
type Video struct {
	gormx.TimestampedBase

	PublicId string `json:"publicId" gorm:"uniqueIndex;not null"`

	Title       string `json:"title" gorm:"not null"`
	Description string `json:"description" gorm:"not null"`

	ThumbnailUrl  string `json:"thumbnailUrl" gorm:"not null"`
	VideoFileName string `json:"videoFileName" gorm:"not null"`
	MaxResolution string `json:"maxResolution" gorm:"not null;default:1080p"`

	ViewsCount int  `json:"viewsCount" gorm:"not null;default:0"`
	IsPublic   bool `json:"isPublic" gorm:"not null;default:false"`

	// Owning channel, cascade-deleted with the channel.
	ChannelID string           `json:"channelId" gorm:"index;not null"`
	Channel   *channel.Channel `json:"channel,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	Comments []VideoComment `json:"comments,omitempty"`
	Likes    []VideoLike    `json:"likes,omitempty"`
	Tags     []VideoTag     `json:"tags,omitempty" gorm:"many2many:video_tags"`
}

// TableName keeps the table name aligned with the Prisma @@map("video").
func (Video) TableName() string { return "video" }

// VideoComment is a user's comment on a video. Mirrors the Prisma
// "video_comment" model.
type VideoComment struct {
	gormx.TimestampedBase

	Text string `json:"text" gorm:"not null"`

	UserID string     `json:"userId" gorm:"index;not null"`
	User   *user.User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	VideoID string `json:"videoId" gorm:"index;not null"`
	Video   *Video `json:"video,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// TableName keeps the table name aligned with the Prisma @@map("video_comment").
func (VideoComment) TableName() string { return "video_comment" }

// VideoLike is a user's like on a video. A user can like a video at most once.
// Mirrors the Prisma "video_like" model (created_at only, no updated_at).
type VideoLike struct {
	gormx.Base

	UserID string     `json:"userId" gorm:"uniqueIndex:idx_video_like_user_video;not null"`
	User   *user.User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	VideoID string `json:"videoId" gorm:"uniqueIndex:idx_video_like_user_video;not null"`
	Video   *Video `json:"video,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// TableName keeps the table name aligned with the Prisma @@map("video_like").
func (VideoLike) TableName() string { return "video_like" }

// VideoTag is a label that can be attached to many videos. Mirrors the Prisma
// "video_tag" model.
type VideoTag struct {
	gormx.TimestampedBase

	Name string `json:"name" gorm:"uniqueIndex;not null"`

	Videos []Video `json:"videos,omitempty" gorm:"many2many:video_tags"`
}

// TableName keeps the table name aligned with the Prisma @@map("video_tag").
func (VideoTag) TableName() string { return "video_tag" }

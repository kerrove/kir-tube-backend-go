package video

import (
	"go/kir-tube/internal/channel"
	"go/kir-tube/internal/user"
	"go/kir-tube/pkg/gormx"
)

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

	ChannelID string           `json:"channelId" gorm:"index;not null"`
	Channel   *channel.Channel `json:"channel,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	Comments []VideoComment `json:"comments,omitempty"`
	Likes    []VideoLike    `json:"likes,omitempty"`
	Tags     []VideoTag     `json:"tags,omitempty" gorm:"many2many:video_tags"`
}

func (Video) TableName() string { return "video" }

type VideoComment struct {
	gormx.TimestampedBase

	Text string `json:"text" gorm:"not null"`

	UserID string     `json:"userId" gorm:"index;not null"`
	User   *user.User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	VideoID string `json:"videoId" gorm:"index;not null"`
	Video   *Video `json:"video,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

func (VideoComment) TableName() string { return "video_comment" }

type VideoLike struct {
	gormx.Base

	UserID string     `json:"userId" gorm:"uniqueIndex:idx_video_like_user_video;not null"`
	User   *user.User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	VideoID string `json:"videoId" gorm:"uniqueIndex:idx_video_like_user_video;not null"`
	Video   *Video `json:"video,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

func (VideoLike) TableName() string { return "video_like" }

type VideoTag struct {
	gormx.TimestampedBase

	Name string `json:"name" gorm:"uniqueIndex;not null"`

	Videos []Video `json:"videos,omitempty" gorm:"many2many:video_tags"`
}

func (VideoTag) TableName() string { return "video_tag" }

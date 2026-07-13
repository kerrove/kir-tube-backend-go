package playlist

import (
	"go/kir-tube/internal/user"
	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/gormx"
)

// Playlist is a user-curated, ordered collection of videos. Mirrors the Prisma
// "playlist" model.
type Playlist struct {
	gormx.TimestampedBase

	Title string `json:"title" gorm:"not null"`

	UserID string     `json:"userId" gorm:"index;not null"`
	User   *user.User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	Videos []video.Video `json:"videos,omitempty" gorm:"many2many:playlist_videos"`
}

// TableName keeps the table name aligned with the Prisma @@map("playlist").
func (Playlist) TableName() string { return "playlist" }

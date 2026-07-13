package history

import (
	"time"

	"go/kir-tube/internal/user"
	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/gormx"
)

// WatchHistory records that a user watched a video. A single row per
// user/video pair is kept, refreshed on re-watch. Mirrors the Prisma
// "watch_history" model (tracks watched_at instead of created/updated_at).
type WatchHistory struct {
	gormx.Identifier

	UserID string     `json:"userId" gorm:"uniqueIndex:idx_watch_history_user_video;not null"`
	User   *user.User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	VideoID string       `json:"videoId" gorm:"uniqueIndex:idx_watch_history_user_video;not null"`
	Video   *video.Video `json:"video,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	WatchedAt time.Time `json:"watchedAt" gorm:"autoCreateTime"`
}

// TableName keeps the table name aligned with the Prisma @@map("watch_history").
func (WatchHistory) TableName() string { return "watch_history" }

package testutil

import (
	"testing"

	"go/kir-tube/internal/channel"
	"go/kir-tube/internal/user"
	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/gormx"
	"go/kir-tube/pkg/password"
)

// CreateUser inserts a user with the given email and a bcrypt hash of plainPass
// (so it can log in through the auth service). It fails the test on error.
func CreateUser(t *testing.T, database *db.Db, email, plainPass string) *user.User {
	t.Helper()
	hash, err := password.Encode(plainPass)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	u := &user.User{Email: email, Password: hash}
	if err := database.Create(u).Error; err != nil {
		t.Fatalf("create user %q: %v", email, err)
	}
	return u
}

// CreateChannel inserts a channel owned by userID.
func CreateChannel(t *testing.T, database *db.Db, userID, slug string) *channel.Channel {
	t.Helper()
	c := &channel.Channel{Slug: slug, UserID: userID}
	if err := database.Create(c).Error; err != nil {
		t.Fatalf("create channel %q: %v", slug, err)
	}
	return c
}

// CreateVideo inserts a public video for channelID with sensible defaults.
func CreateVideo(t *testing.T, database *db.Db, channelID, title string) *video.Video {
	t.Helper()
	v := &video.Video{
		PublicId:      gormx.NewID(),
		Title:         title,
		Description:   "description",
		ThumbnailUrl:  "/uploads/thumbnails/thumb.jpg",
		VideoFileName: "video.mp4",
		MaxResolution: "1080p",
		IsPublic:      true,
		ChannelID:     channelID,
	}
	if err := database.Create(v).Error; err != nil {
		t.Fatalf("create video %q: %v", title, err)
	}
	return v
}

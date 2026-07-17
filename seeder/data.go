package seeder

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed data/channels.json data/videos.json
var dataFS embed.FS

// ChannelSeed is one row of data/channels.json. Name belongs to the channel's
// owner (user.User), the rest to channel.Channel.
type ChannelSeed struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	AvatarUrl   *string `json:"avatarUrl"`
	BannerUrl   *string `json:"bannerUrl"`
	IsVerified  bool    `json:"isVerified"`
}

// VideoSeed is one row of data/videos.json. Slug carries the public id used in
// video URLs; when empty the seeder generates one.
type VideoSeed struct {
	Title         string   `json:"title"`
	Slug          string   `json:"slug"`
	Description   string   `json:"description"`
	ThumbnailUrl  string   `json:"thumbnailUrl"`
	VideoFileName string   `json:"videoFileName"`
	MaxResolution string   `json:"maxResolution"`
	ViewsCount    int      `json:"viewsCount"`
	IsPublic      bool     `json:"isPublic"`
	ChannelSlug   string   `json:"channelSlug"`
	Tags          []string `json:"tags"`
}

func loadChannels() ([]ChannelSeed, error) {
	var channels []ChannelSeed
	if err := loadJSON("data/channels.json", &channels); err != nil {
		return nil, err
	}
	return channels, nil
}

func loadVideos() ([]VideoSeed, error) {
	var videos []VideoSeed
	if err := loadJSON("data/videos.json", &videos); err != nil {
		return nil, err
	}
	return videos, nil
}

func loadJSON(name string, dst any) error {
	raw, err := dataFS.ReadFile(name)
	if err != nil {
		return fmt.Errorf("read %s: %w", name, err)
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("parse %s: %w", name, err)
	}
	return nil
}

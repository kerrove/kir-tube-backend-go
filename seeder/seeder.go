// Package seeder fills the database with the demo channels and videos used for
// local development.
package seeder

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"gorm.io/gorm"

	"go/kir-tube/internal/channel"
	"go/kir-tube/internal/user"
	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/password"
)

// seedEmailDomain marks the accounts this package owns. Clean only removes
// users under this domain, so real accounts survive a re-seed.
const seedEmailDomain = "kir-tube.seed"

// seedPassword is the plain-text password shared by every demo channel owner.
const seedPassword = "123456"

// cleanupTables lists the tables the seeder writes to, ordered so that child
// rows go before the rows they reference.
var cleanupTables = []string{
	"video_tags",
	"playlist_videos",
	"watch_history",
	"video_like",
	"video_comment",
	"video",
	"video_tag",
	"channel_subscribers",
	"channel",
}

type Seeder struct {
	Database *db.Db
}

func NewSeeder(database *db.Db) *Seeder {
	return &Seeder{Database: database}
}

// Report counts what a Run inserted.
type Report struct {
	Channels int
	Videos   int
	Tags     int
}

// Run seeds the database inside a single transaction. When clean is true the
// previously seeded rows are removed first, making the run repeatable.
func (s *Seeder) Run(clean bool) (Report, error) {
	channels, err := loadChannels()
	if err != nil {
		return Report{}, err
	}
	videos, err := loadVideos()
	if err != nil {
		return Report{}, err
	}

	var report Report
	err = s.Database.DB.Transaction(func(tx *gorm.DB) error {
		if clean {
			if err := cleanup(tx); err != nil {
				return err
			}
		}

		channelIDs, err := seedChannels(tx, channels)
		if err != nil {
			return err
		}

		tags, err := seedVideos(tx, videos, channelIDs)
		if err != nil {
			return err
		}

		report = Report{Channels: len(channels), Videos: len(videos), Tags: tags}
		return nil
	})

	return report, err
}

func cleanup(tx *gorm.DB) error {
	for _, table := range cleanupTables {
		if !tx.Migrator().HasTable(table) {
			continue
		}
		if err := tx.Exec(`DELETE FROM ` + tx.Statement.Quote(table)).Error; err != nil {
			return fmt.Errorf("clean %s: %w", table, err)
		}
	}

	if err := tx.Where("email LIKE ?", "%@"+seedEmailDomain).Delete(&user.User{}).Error; err != nil {
		return fmt.Errorf("clean seeded users: %w", err)
	}
	return nil
}

// seedChannels creates an owner account per channel and returns the channel id
// for every slug, so videos can be attached to them.
func seedChannels(tx *gorm.DB, seeds []ChannelSeed) (map[string]string, error) {
	hashed, err := password.Encode(seedPassword)
	if err != nil {
		return nil, fmt.Errorf("hash seed password: %w", err)
	}

	ids := make(map[string]string, len(seeds))
	for _, seed := range seeds {
		name := seed.Name
		owner := user.User{
			Name:     &name,
			Email:    seed.Slug + "@" + seedEmailDomain,
			Password: hashed,
		}
		if err := tx.Create(&owner).Error; err != nil {
			return nil, fmt.Errorf("create owner of %s: %w", seed.Slug, err)
		}

		ch := channel.Channel{
			Slug:        seed.Slug,
			Description: seed.Description,
			AvatarUrl:   seed.AvatarUrl,
			BannerUrl:   seed.BannerUrl,
			IsVerified:  seed.IsVerified,
			UserID:      owner.ID,
		}
		if err := tx.Create(&ch).Error; err != nil {
			return nil, fmt.Errorf("create channel %s: %w", seed.Slug, err)
		}

		ids[seed.Slug] = ch.ID
	}
	return ids, nil
}

// seedVideos creates the videos and their tags, returning the number of
// distinct tags inserted.
func seedVideos(tx *gorm.DB, seeds []VideoSeed, channelIDs map[string]string) (int, error) {
	tags := map[string]*video.VideoTag{}

	for _, seed := range seeds {
		channelID, ok := channelIDs[seed.ChannelSlug]
		if !ok {
			return 0, fmt.Errorf("video %q references unknown channel %q", seed.Title, seed.ChannelSlug)
		}

		publicID := seed.Slug
		if publicID == "" {
			generated, err := newPublicID()
			if err != nil {
				return 0, err
			}
			publicID = generated
		}

		v := video.Video{
			PublicId:      publicID,
			Title:         seed.Title,
			Description:   seed.Description,
			ThumbnailUrl:  seed.ThumbnailUrl,
			VideoFileName: seed.VideoFileName,
			MaxResolution: seed.MaxResolution,
			ViewsCount:    seed.ViewsCount,
			IsPublic:      seed.IsPublic,
			ChannelID:     channelID,
		}
		if err := tx.Create(&v).Error; err != nil {
			return 0, fmt.Errorf("create video %q: %w", seed.Title, err)
		}

		attached, err := resolveTags(tx, tags, seed.Tags)
		if err != nil {
			return 0, fmt.Errorf("tags of video %q: %w", seed.Title, err)
		}
		if len(attached) == 0 {
			continue
		}
		if err := tx.Model(&v).Association("Tags").Append(attached); err != nil {
			return 0, fmt.Errorf("attach tags to video %q: %w", seed.Title, err)
		}
	}

	return len(tags), nil
}

// resolveTags returns the tag rows for the given names, creating the ones that
// do not exist yet. The cache keeps a tag shared across videos.
func resolveTags(tx *gorm.DB, cache map[string]*video.VideoTag, names []string) ([]*video.VideoTag, error) {
	resolved := make([]*video.VideoTag, 0, len(names))
	for _, name := range names {
		tag, ok := cache[name]
		if !ok {
			tag = &video.VideoTag{}
			if err := tx.Where(video.VideoTag{Name: name}).FirstOrCreate(tag).Error; err != nil {
				return nil, fmt.Errorf("resolve tag %q: %w", name, err)
			}
			cache[name] = tag
		}
		resolved = append(resolved, tag)
	}
	return resolved, nil
}

// newPublicID mirrors the 12-character hex ids the existing seed data uses for
// videos that ship without one.
func newPublicID() (string, error) {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate public id: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

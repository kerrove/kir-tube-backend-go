package playlist

import (
	"errors"
	"slices"

	"gorm.io/gorm"

	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/di"
)

type PlaylistRepository struct {
	Database        *db.Db
	VideoRepository di.IPlaylistVideoRepository
}

func NewPlaylistRepository(database *db.Db, videoRepository di.IPlaylistVideoRepository) *PlaylistRepository {
	return &PlaylistRepository{Database: database, VideoRepository: videoRepository}
}

// FindById loads a playlist together with its videos (newest first), each video
// carrying its channel and the channel's owner.
func (rep *PlaylistRepository) FindById(id string) (*Playlist, error) {
	var playlist Playlist
	err := rep.Database.DB.
		Preload("Videos", func(db *gorm.DB) *gorm.DB {
			return db.Order("video.created_at desc")
		}).
		Preload("Videos.Channel.User").
		First(&playlist, "id = ?", id).
		Error

	if err != nil {
		return nil, err
	}

	return &playlist, nil
}
func (rep *PlaylistRepository) FindByUserId(userId string) (*[]Playlist, error) {
	var playlists []Playlist
	err := rep.Database.DB.
		Preload("Videos").
		Where("user_id = ?", userId).
		Order("created_at desc").
		Find(&playlists).
		Error

	if err != nil {
		return nil, err
	}

	return &playlists, nil
}
func (rep *PlaylistRepository) ToggleVideo(playlistId, videoId, userId string) (*ToggleVideoResponse, error) {
	var playlist Playlist
	err := rep.Database.DB.
		Preload("Videos").
		First(&playlist, "user_id = ? AND id = ?", userId, playlistId).
		Error

	if err != nil {
		return nil, errors.New(ErrPlaylistNotExist)
	}

	exists, err := rep.VideoRepository.ExistsById(videoId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New(ErrVideoNotExist)
	}

	// Videos was preloaded, so the current membership is known without another
	// query.
	isVideoInPlaylist := slices.ContainsFunc(playlist.Videos, func(video video.Video) bool {
		return video.ID == videoId
	})

	if isVideoInPlaylist {
		err := rep.Database.DB.
			Exec("DELETE FROM playlist_videos WHERE playlist_id = ? AND video_id = ?", playlistId, videoId).
			Error

		if err != nil {
			return nil, err
		}

		return &ToggleVideoResponse{Message: "Видео удалено из плейлиста"}, nil
	}

	// ON CONFLICT DO NOTHING keeps concurrent toggles from failing on the
	// composite primary key.
	err = rep.Database.DB.
		Exec("INSERT INTO playlist_videos (playlist_id, video_id) VALUES (?, ?) ON CONFLICT DO NOTHING", playlistId, videoId).
		Error

	if err != nil {
		return nil, err
	}

	return &ToggleVideoResponse{Message: "Видео добавлено в плейлист"}, nil
}

// Create makes a new playlist for the user. When body.VideoPublicId is set the
// referenced video is attached to the playlist; a missing video is reported as
// ErrVideoNotExist (mapped to 404), mirroring the Prisma createPlaylist method.
// The returned playlist has its videos loaded (Prisma's include: { videos }).
func (rep *PlaylistRepository) Create(userId string, body *PlaylistRequest) (*Playlist, error) {
	var videoId string
	if body.VideoPublicId != "" {
		id, found, err := rep.VideoRepository.FindIdByPublicId(body.VideoPublicId)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, ErrVideoNotFound
		}
		videoId = id
	}

	playlist := &Playlist{
		UserID: userId,
		Title:  body.Title,
	}

	err := rep.Database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(playlist).Error; err != nil {
			return err
		}

		if videoId != "" {
			// Insert straight into the join table (like ToggleVideo) so GORM
			// never touches the video row itself.
			if err := tx.Exec(
				"INSERT INTO playlist_videos (playlist_id, video_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
				playlist.ID, videoId,
			).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reload with the videos attached, matching the Prisma include.
	if err := rep.Database.DB.
		Preload("Videos").
		First(playlist, "id = ?", playlist.ID).Error; err != nil {
		return nil, err
	}

	return playlist, nil
}

package playlist_test

import (
	"testing"

	"go/kir-tube/internal/playlist"
	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPlaylistRepo(t *testing.T) (*playlist.PlaylistRepository, *testutilCtx) {
	db := testutil.RequireDB(t)
	repo := playlist.NewPlaylistRepository(db, video.NewVideoRepository(db))
	owner := testutil.CreateUser(t, db, "owner@b.c", "p")
	ch := testutil.CreateChannel(t, db, owner.ID, "redgroup")
	v := testutil.CreateVideo(t, db, ch.ID, "Clip")
	return repo, &testutilCtx{ownerID: owner.ID, videoID: v.ID, videoPublicID: v.PublicId}
}

type testutilCtx struct {
	ownerID       string
	videoID       string
	videoPublicID string
}

func TestPlaylistRepository_CreateWithoutVideo(t *testing.T) {
	repo, ctx := newPlaylistRepo(t)

	pl, err := repo.Create(ctx.ownerID, &playlist.PlaylistRequest{Title: "Favourites"})
	require.NoError(t, err)
	assert.Equal(t, "Favourites", pl.Title)
	assert.Len(t, pl.Videos, 0)

	byUser, err := repo.FindByUserId(ctx.ownerID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(*byUser), 1)

	byID, err := repo.FindById(pl.ID)
	require.NoError(t, err)
	assert.Equal(t, "Favourites", byID.Title)
}

func TestPlaylistRepository_CreateWithVideo(t *testing.T) {
	repo, ctx := newPlaylistRepo(t)

	pl, err := repo.Create(ctx.ownerID, &playlist.PlaylistRequest{
		Title:         "With video",
		VideoPublicId: ctx.videoPublicID,
	})
	require.NoError(t, err)
	assert.Len(t, pl.Videos, 1)
}

func TestPlaylistRepository_CreateMissingVideo(t *testing.T) {
	repo, ctx := newPlaylistRepo(t)

	_, err := repo.Create(ctx.ownerID, &playlist.PlaylistRequest{
		Title:         "Broken",
		VideoPublicId: "no-such-public-id",
	})
	assert.ErrorIs(t, err, playlist.ErrVideoNotFound)
}

func TestPlaylistRepository_ToggleVideo(t *testing.T) {
	repo, ctx := newPlaylistRepo(t)

	pl, err := repo.Create(ctx.ownerID, &playlist.PlaylistRequest{Title: "Toggle"})
	require.NoError(t, err)

	added, err := repo.ToggleVideo(pl.ID, ctx.videoID, ctx.ownerID)
	require.NoError(t, err)
	assert.Contains(t, added.Message, "добавлено")

	removed, err := repo.ToggleVideo(pl.ID, ctx.videoID, ctx.ownerID)
	require.NoError(t, err)
	assert.Contains(t, removed.Message, "удалено")
}

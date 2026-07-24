package video_test

import (
	"testing"

	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVideoRepository_CreateFindSearch(t *testing.T) {
	db := testutil.RequireDB(t)
	repo := video.NewVideoRepository(db)

	owner := testutil.CreateUser(t, db, "owner@b.c", "p")
	ch := testutil.CreateChannel(t, db, owner.ID, "redgroup")

	created, err := repo.Create(ch.ID, video.CreateVideoInput{
		Title:       "Go Tutorial",
		Description: "learn go",
		Tags:        []string{"golang", "tutorial"},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.PublicId)
	assert.True(t, created.IsPublic)
	assert.Equal(t, "1080p", created.MaxResolution)
	assert.Len(t, created.Tags, 2)

	found, err := repo.FindByPublicId(created.PublicId)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)

	// ILIKE search matches title case-insensitively.
	hits, err := repo.FindPublicVideos("TUTORIAL", 0, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(hits), 1)

	count, err := repo.CountPublicVideos("tutorial")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(1))

	misses, err := repo.FindPublicVideos("zzz-no-such-term", 0, 10)
	require.NoError(t, err)
	assert.Len(t, misses, 0)
}

func TestVideoRepository_ViewsAndLikes(t *testing.T) {
	db := testutil.RequireDB(t)
	repo := video.NewVideoRepository(db)

	owner := testutil.CreateUser(t, db, "owner@b.c", "p")
	ch := testutil.CreateChannel(t, db, owner.ID, "redgroup")
	v := testutil.CreateVideo(t, db, ch.ID, "Clip")

	inc, err := repo.IncrementViewsCount(v.PublicId)
	require.NoError(t, err)
	assert.Equal(t, v.ViewsCount+1, inc.ViewsCount)

	liked, err := repo.ToggleLike(owner.ID, v.ID)
	require.NoError(t, err)
	assert.True(t, liked, "first toggle should like")

	liked, err = repo.ToggleLike(owner.ID, v.ID)
	require.NoError(t, err)
	assert.False(t, liked, "second toggle should unlike")
}

func TestVideoRepository_UpdateAndDelete(t *testing.T) {
	db := testutil.RequireDB(t)
	repo := video.NewVideoRepository(db)

	owner := testutil.CreateUser(t, db, "owner@b.c", "p")
	ch := testutil.CreateChannel(t, db, owner.ID, "redgroup")
	v := testutil.CreateVideo(t, db, ch.ID, "Original")

	newTitle := "Updated Title"
	updated, err := repo.Update(v.ID, video.UpdateVideoInput{Title: &newTitle})
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)

	_, err = repo.Delete(v.ID)
	require.NoError(t, err)

	_, err = repo.FindById(v.ID)
	assert.Error(t, err, "deleted video should not be found")
}

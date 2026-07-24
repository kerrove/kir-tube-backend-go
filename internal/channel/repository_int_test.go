package channel_test

import (
	"testing"

	"go/kir-tube/internal/channel"
	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelRepository_FindBySlugAndAll(t *testing.T) {
	db := testutil.RequireDB(t)
	repo := channel.NewChannelRepository(db, video.NewVideoRepository(db))

	owner := testutil.CreateUser(t, db, "owner@b.c", "p")
	testutil.CreateChannel(t, db, owner.ID, "redgroup")

	details, err := repo.FindBySlug("redgroup")
	require.NoError(t, err)
	assert.Equal(t, "redgroup", details.Slug)
	assert.NotNil(t, details.Videos) // never nil

	all := repo.FindAll()
	require.NotNil(t, all)
	assert.GreaterOrEqual(t, len(*all), 1)

	_, err = repo.FindBySlug("missing-slug")
	assert.Error(t, err)
}

func TestChannelRepository_ToggleSubscribe(t *testing.T) {
	db := testutil.RequireDB(t)
	repo := channel.NewChannelRepository(db, video.NewVideoRepository(db))

	owner := testutil.CreateUser(t, db, "owner@b.c", "p")
	testutil.CreateChannel(t, db, owner.ID, "redgroup")
	subscriber := testutil.CreateUser(t, db, "sub@b.c", "p")

	subscribed, err := repo.ToggleSubscribe("redgroup", subscriber.ID)
	require.NoError(t, err)
	assert.True(t, subscribed, "first toggle should subscribe")

	subscribed, err = repo.ToggleSubscribe("redgroup", subscriber.ID)
	require.NoError(t, err)
	assert.False(t, subscribed, "second toggle should unsubscribe")
}

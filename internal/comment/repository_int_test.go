package comment_test

import (
	"testing"

	"go/kir-tube/internal/comment"
	"go/kir-tube/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strptr(s string) *string { return &s }

func TestCommentRepository_Lifecycle(t *testing.T) {
	db := testutil.RequireDB(t)
	repo := comment.NewCommentRepository(db)

	owner := testutil.CreateUser(t, db, "owner@b.c", "p")
	ch := testutil.CreateChannel(t, db, owner.ID, "redgroup")
	v := testutil.CreateVideo(t, db, ch.ID, "Clip")

	created, err := repo.Create(owner.ID, &comment.CreateCommentReq{
		Text:    strptr("nice video"),
		VideoId: strptr(v.ID),
	})
	require.NoError(t, err)
	assert.Equal(t, "nice video", created.Text)

	list, err := repo.FindByVideo(v.PublicId)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "nice video", list[0].Text)

	updated, err := repo.Update(created.ID, owner.ID, &comment.UpdateCommentReq{Text: strptr("edited")})
	require.NoError(t, err)
	assert.Equal(t, "edited", updated.Text)
}

func TestCommentRepository_ForbidsNonOwner(t *testing.T) {
	db := testutil.RequireDB(t)
	repo := comment.NewCommentRepository(db)

	owner := testutil.CreateUser(t, db, "owner@b.c", "p")
	ch := testutil.CreateChannel(t, db, owner.ID, "redgroup")
	v := testutil.CreateVideo(t, db, ch.ID, "Clip")
	stranger := testutil.CreateUser(t, db, "stranger@b.c", "p")

	created, err := repo.Create(owner.ID, &comment.CreateCommentReq{
		Text:    strptr("mine"),
		VideoId: strptr(v.ID),
	})
	require.NoError(t, err)

	_, err = repo.Update(created.ID, stranger.ID, &comment.UpdateCommentReq{Text: strptr("hijack")})
	assert.ErrorIs(t, err, comment.ErrCommentForbidden)

	_, err = repo.Delete(created.ID, stranger.ID)
	assert.ErrorIs(t, err, comment.ErrCommentForbidden)

	ok, err := repo.Delete(created.ID, owner.ID)
	require.NoError(t, err)
	assert.True(t, ok)
}

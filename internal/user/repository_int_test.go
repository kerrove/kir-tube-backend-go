package user_test

import (
	"testing"

	"go/kir-tube/internal/user"
	"go/kir-tube/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CRUD(t *testing.T) {
	db := testutil.RequireDB(t)
	repo := user.NewUserRepository(db)

	created, err := repo.Create(&user.User{Email: "a@b.c", Password: "hash"})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	// BeforeCreate assigns a verification token.
	require.NotNil(t, created.VerificationToken)
	require.NotEmpty(t, *created.VerificationToken)

	byID, err := repo.FindById(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "a@b.c", byID.Email)

	byEmail, err := repo.FindByEmail("a@b.c")
	require.NoError(t, err)
	assert.Equal(t, created.ID, byEmail.ID)

	byToken, err := repo.FindByVerifyToken(*created.VerificationToken)
	require.NoError(t, err)
	assert.Equal(t, created.ID, byToken.ID)

	name := "Kir"
	created.Name = &name
	updated, err := repo.Update(created)
	require.NoError(t, err)
	require.NotNil(t, updated.Name)
	assert.Equal(t, "Kir", *updated.Name)
}

func TestUserRepository_FindMissing(t *testing.T) {
	db := testutil.RequireDB(t)
	repo := user.NewUserRepository(db)

	_, err := repo.FindByEmail("nobody@nowhere.test")
	assert.Error(t, err)

	_, err = repo.FindById("does-not-exist")
	assert.Error(t, err)
}

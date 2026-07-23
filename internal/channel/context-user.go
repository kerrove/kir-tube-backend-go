package channel

import (
	"errors"

	"gorm.io/gorm"

	"go/kir-tube/internal/user"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/di"
)

// ContextUserRepository loads the authenticated user together with their channel
// for the auth middleware (di.IUserProvider). It lives in the channel package
// because it needs both the user and channel tables, and channel already
// depends on user — so user cannot depend on channel without an import cycle.
type ContextUserRepository struct {
	Database *db.Db
}

func NewContextUserRepository(database *db.Db) *ContextUserRepository {
	return &ContextUserRepository{Database: database}
}

// ContextUserRepository is the concrete user+channel loader for the middleware.
var _ di.IUserProvider = (*ContextUserRepository)(nil)

// FindContextUser loads a user by id and their channel (if any), projecting both
// into the di.ContextUser the auth middleware stores in the request context.
func (repo *ContextUserRepository) FindContextUser(id string) (*di.ContextUser, error) {
	var u user.User
	if err := repo.Database.DB.First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}

	ctxUser := &di.ContextUser{
		ID:                u.ID,
		Name:              u.Name,
		Email:             u.Email,
		Password:          u.Password,
		VerificationToken: u.VerificationToken,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}

	var ch Channel
	err := repo.Database.DB.First(&ch, "user_id = ?", id).Error
	switch {
	case err == nil:
		ctxUser.Channel = &di.ContextChannel{
			ID:          ch.ID,
			Slug:        ch.Slug,
			Description: ch.Description,
			IsVerified:  ch.IsVerified,
			AvatarUrl:   ch.AvatarUrl,
			BannerUrl:   ch.BannerUrl,
			UserID:      ch.UserID,
			CreatedAt:   ch.CreatedAt,
			UpdatedAt:   ch.UpdatedAt,
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		// User has no channel yet — leave ContextUser.Channel nil.
	default:
		return nil, err
	}

	return ctxUser, nil
}

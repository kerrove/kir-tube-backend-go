package di

import "time"

// IUserProvider is the port the auth middleware needs to load the full user
// (with their channel) behind a token by id. It lives in di (which imports no
// domain package) so that pkg/middleware and pkg/req can resolve the
// authenticated user without importing internal/user — pkg must not depend on
// internal. channel.ContextUserRepository satisfies it (it reads both the user
// and channel tables, which user itself cannot without an import cycle).
type IUserProvider interface {
	FindContextUser(id string) (*ContextUser, error)
}

// ContextUser is the authenticated user the auth middleware puts into the
// request context. It carries the full user record plus the caller's channel,
// so handlers can read the caller's identity without an extra database round
// trip. The bearer token is unchanged — it still only carries the id; the
// middleware loads the rest.
//
// Password and VerificationToken are kept out of JSON (json:"-") so the record
// never leaks if a handler serialises it, mirroring the user.User model.
type ContextUser struct {
	ID                string          `json:"id"`
	Name              *string         `json:"name"`
	Email             string          `json:"email"`
	Password          string          `json:"-"`
	VerificationToken *string         `json:"-"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	Channel           *ContextChannel `json:"channel,omitempty"`
}

// ContextChannel is the caller's channel projected into ContextUser. It is nil
// when the user has not created a channel yet.
type ContextChannel struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	IsVerified  bool      `json:"isVerified"`
	AvatarUrl   *string   `json:"avatarUrl"`
	BannerUrl   *string   `json:"bannerUrl"`
	UserID      string    `json:"userId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

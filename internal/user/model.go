package user

import (
	"gorm.io/gorm"

	"go/kir-tube/pkg/gormx"
)

// User is the account that owns a channel, playlists, comments and likes.
// Mirrors the Prisma "user" model.
type User struct {
	gormx.TimestampedBase

	Name     *string `json:"name"`
	Email    string  `json:"email" gorm:"uniqueIndex;not null"`
	Password string  `json:"-" gorm:"not null"`

	VerificationToken *string `json:"verificationToken,omitempty"`
}

// TableName keeps the table name aligned with the Prisma @@map("user").
func (User) TableName() string { return "user" }

// BeforeCreate generates the primary key and seeds a verification token,
// matching Prisma's cuid() id and uuid() default for verification_token.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if err := u.Identifier.BeforeCreate(tx); err != nil {
		return err
	}
	if u.VerificationToken == nil || *u.VerificationToken == "" {
		token := gormx.NewID()
		u.VerificationToken = &token
	}
	return nil
}

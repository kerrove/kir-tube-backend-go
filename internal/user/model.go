package user

import (
	"gorm.io/gorm"

	"go/kir-tube/pkg/gormx"
)

type User struct {
	gormx.TimestampedBase

	Name     *string `json:"name"`
	Email    string  `json:"email" gorm:"uniqueIndex;not null"`
	Password string  `json:"-" gorm:"not null"`

	VerificationToken *string `json:"-"`
}

func (User) TableName() string { return "user" }

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

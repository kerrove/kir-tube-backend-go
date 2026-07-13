// Package gormx contains shared building blocks for the project's GORM models.
package gormx

import (
	"crypto/rand"
	"time"

	"gorm.io/gorm"
)

// idAlphabet is the character set used for generated identifiers (base36).
const idAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

// idBodyLength is the number of random characters appended after the "c" prefix.
const idBodyLength = 24

// newID returns a collision-resistant, cuid-like identifier such as
// "cf3k9x2...". It mirrors the string IDs produced by Prisma's cuid() so that
// data migrated from the NestJS backend keeps the same primary-key shape.
func newID() (string, error) {
	buf := make([]byte, idBodyLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i, b := range buf {
		buf[i] = idAlphabet[int(b)%len(idAlphabet)]
	}
	return "c" + string(buf), nil
}

// NewID returns a fresh identifier and panics only if the system CSPRNG fails,
// which is treated as unrecoverable. Use it outside of GORM hooks (e.g. to seed
// tokens) where returning an error is inconvenient.
func NewID() string {
	id, err := newID()
	if err != nil {
		panic("gormx: failed to generate id: " + err.Error())
	}
	return id
}

// Identifier is embedded by every model to provide a string primary key that is
// auto-generated on insert when left empty.
type Identifier struct {
	ID string `gorm:"primaryKey;type:varchar(30)" json:"id"`
}

// BeforeCreate populates the primary key if the caller did not set one.
func (m *Identifier) BeforeCreate(*gorm.DB) error {
	if m.ID != "" {
		return nil
	}
	id, err := newID()
	if err != nil {
		return err
	}
	m.ID = id
	return nil
}

// Base is for records that only track their creation time
// (Prisma models with just a created_at column).
type Base struct {
	Identifier
	CreatedAt time.Time `json:"createdAt"`
}

// TimestampedBase is for records that track both creation and update times
// (Prisma models with created_at + updated_at).
type TimestampedBase struct {
	Base
	UpdatedAt time.Time `json:"updatedAt"`
}

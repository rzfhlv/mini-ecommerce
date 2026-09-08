package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrEmptyName = errors.New("user: name cannot be empty")

type User struct {
	id                       uuid.UUID
	name                     string
	email                    Email
	passwordHash             PasswordHash
	createdAt, updatedAt     time.Time
}

func NewUser(name string, email Email, passwordHash PasswordHash) (*User, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	now := time.Now()
	return &User{id: uuid.New(), name: name, email: email, passwordHash: passwordHash, createdAt: now, updatedAt: now}, nil
}

func ReconstructUser(id uuid.UUID, name string, email Email, passwordHash PasswordHash, createdAt, updatedAt time.Time) *User {
	return &User{id: id, name: name, email: email, passwordHash: passwordHash, createdAt: createdAt, updatedAt: updatedAt}
}

func (u *User) ID() uuid.UUID             { return u.id }
func (u *User) Name() string              { return u.name }
func (u *User) Email() Email              { return u.email }
func (u *User) PasswordHash() PasswordHash { return u.passwordHash }
func (u *User) CreatedAt() time.Time      { return u.createdAt }

func (u *User) VerifyPassword(plainPassword string, hasher PasswordHasher) bool {
	return hasher.Compare(u.passwordHash, plainPassword)
}

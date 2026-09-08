package domain

import "errors"

var ErrEmptyPasswordHash = errors.New("password hash: cannot be empty")

type PasswordHash struct{ value string }

func NewPasswordHash(hashed string) (PasswordHash, error) {
	if hashed == "" {
		return PasswordHash{}, ErrEmptyPasswordHash
	}
	return PasswordHash{value: hashed}, nil
}

func (p PasswordHash) String() string { return p.value }

// PasswordHasher -- port milik domain, diimplementasikan infrastructure (bcrypt).
type PasswordHasher interface {
	Hash(plainPassword string) (PasswordHash, error)
	Compare(hash PasswordHash, plainPassword string) bool
}

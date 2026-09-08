package domain

import (
	"context"
	"errors"
)

var (
	ErrEmailAlreadyRegistered = errors.New("registration: email already registered")
	ErrPasswordTooShort       = errors.New("registration: password must be at least 8 characters")
)

const minPasswordLength = 8

type RegistrationService struct {
	userRepo UserRepository
	hasher   PasswordHasher
}

func NewRegistrationService(userRepo UserRepository, hasher PasswordHasher) *RegistrationService {
	return &RegistrationService{userRepo: userRepo, hasher: hasher}
}

func (s *RegistrationService) Register(ctx context.Context, name, rawEmail, plainPassword string) (*User, error) {
	email, err := NewEmail(rawEmail)
	if err != nil {
		return nil, err
	}
	if len(plainPassword) < minPasswordLength {
		return nil, ErrPasswordTooShort
	}
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyRegistered
	}
	passwordHash, err := s.hasher.Hash(plainPassword)
	if err != nil {
		return nil, err
	}
	user, err := NewUser(name, email, passwordHash)
	if err != nil {
		return nil, err
	}
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

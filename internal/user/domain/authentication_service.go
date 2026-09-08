package domain

import (
	"context"
	"errors"
)

var ErrInvalidCredentials = errors.New("authentication: invalid email or password")

type AuthenticationService struct {
	userRepo UserRepository
	hasher   PasswordHasher
}

func NewAuthenticationService(userRepo UserRepository, hasher PasswordHasher) *AuthenticationService {
	return &AuthenticationService{userRepo: userRepo, hasher: hasher}
}

func (s *AuthenticationService) Authenticate(ctx context.Context, rawEmail, plainPassword string) (*User, error) {
	email, err := NewEmail(rawEmail)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.VerifyPassword(plainPassword, s.hasher) {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

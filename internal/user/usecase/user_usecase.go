package usecase

import (
	"context"

	"github.com/google/uuid"
	"mini-ecommerce/internal/user/domain"
	"mini-ecommerce/pkg/jwt"
)

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type RegisterUseCase interface {
	Execute(ctx context.Context, req RegisterRequest) (*UserResponse, error)
}
type registerUseCase struct{ registrationService *domain.RegistrationService }

func NewRegisterUseCase(s *domain.RegistrationService) RegisterUseCase { return &registerUseCase{registrationService: s} }
func (uc *registerUseCase) Execute(ctx context.Context, req RegisterRequest) (*UserResponse, error) {
	user, err := uc.registrationService.Register(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

type LoginUseCase interface {
	Execute(ctx context.Context, req LoginRequest) (*AuthResponse, error)
}
type loginUseCase struct {
	authService  *domain.AuthenticationService
	tokenManager *jwt.TokenManager
}

func NewLoginUseCase(authService *domain.AuthenticationService, tokenManager *jwt.TokenManager) LoginUseCase {
	return &loginUseCase{authService: authService, tokenManager: tokenManager}
}
func (uc *loginUseCase) Execute(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := uc.authService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	token, err := uc.tokenManager.Generate(user.ID().String(), user.Email().String())
	if err != nil {
		return nil, err
	}
	return &AuthResponse{AccessToken: token, User: *toUserResponse(user)}, nil
}

type GetMeUseCase interface {
	Execute(ctx context.Context, userID string) (*UserResponse, error)
}
type getMeUseCase struct{ userRepo domain.UserRepository }

func NewGetMeUseCase(userRepo domain.UserRepository) GetMeUseCase { return &getMeUseCase{userRepo: userRepo} }
func (uc *getMeUseCase) Execute(ctx context.Context, userIDStr string) (*UserResponse, error) {
	id, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func toUserResponse(user *domain.User) *UserResponse {
	return &UserResponse{ID: user.ID().String(), Name: user.Name(), Email: user.Email().String()}
}

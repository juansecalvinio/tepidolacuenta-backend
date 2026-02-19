package usecase

import (
	"context"
	"errors"

	"juansecalvinio/tepidolacuenta/internal/auth/domain"
	"juansecalvinio/tepidolacuenta/internal/auth/repository"
	"juansecalvinio/tepidolacuenta/internal/pkg"
)

// UseCase defines the interface for authentication use cases
type UseCase interface {
	Register(ctx context.Context, input domain.RegisterInput) (*domain.User, error)
	Login(ctx context.Context, input domain.LoginInput) (*domain.LoginResponse, error)
}

type authUseCase struct {
	repo       repository.Repository
	jwtService *pkg.JWTService
}

// NewAuthUseCase creates a new authentication use case
func NewAuthUseCase(repo repository.Repository, jwtService *pkg.JWTService) UseCase {
	return &authUseCase{
		repo:       repo,
		jwtService: jwtService,
	}
}

// Register registers a new user
func (uc *authUseCase) Register(ctx context.Context, input domain.RegisterInput) (*domain.User, error) {
	// Validate input
	if !pkg.IsValidEmail(input.Email) {
		return nil, pkg.ErrInvalidInput
	}

	if !pkg.IsValidPassword(input.Password) {
		return nil, errors.New("password must be at least 8 characters")
	}

	// Create new user
	user := domain.NewUser(input.Email, input.Password)

	// Hash password
	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	// Save to database
	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (uc *authUseCase) Login(ctx context.Context, input domain.LoginInput) (*domain.LoginResponse, error) {
	// Find user by email
	user, err := uc.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pkg.ErrUserNotFound) {
			return nil, pkg.ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if err := user.ComparePassword(input.Password); err != nil {
		return nil, pkg.ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := uc.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

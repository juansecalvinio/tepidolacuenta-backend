package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"juansecalvinio/tepidolacuenta/internal/auth/domain"
	"juansecalvinio/tepidolacuenta/internal/auth/repository"
	"juansecalvinio/tepidolacuenta/internal/pkg"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

// UseCase defines the interface for authentication use cases
type UseCase interface {
	Register(ctx context.Context, input domain.RegisterInput) (*domain.User, error)
	Login(ctx context.Context, input domain.LoginInput) (*domain.LoginResponse, error)
	ForgotPassword(ctx context.Context, input domain.ForgotPasswordInput) error
	ResetPassword(ctx context.Context, input domain.ResetPasswordInput) error
	GoogleAuthURL(state string) string
	HandleGoogleCallback(ctx context.Context, code string) (*domain.LoginResponse, error)
}

type authUseCase struct {
	repo            repository.Repository
	jwtService      *pkg.JWTService
	emailService    *pkg.EmailService
	frontendBaseURL string
	googleOAuth     *oauth2.Config
}

// NewAuthUseCase creates a new authentication use case
func NewAuthUseCase(repo repository.Repository, jwtService *pkg.JWTService, emailService *pkg.EmailService, frontendBaseURL string, googleOAuth *oauth2.Config) UseCase {
	return &authUseCase{
		repo:            repo,
		jwtService:      jwtService,
		emailService:    emailService,
		frontendBaseURL: frontendBaseURL,
		googleOAuth:     googleOAuth,
	}
}

// Register registers a new user
func (uc *authUseCase) Register(ctx context.Context, input domain.RegisterInput) (*domain.User, error) {
	if !pkg.IsValidEmail(input.Email) {
		return nil, pkg.ErrInvalidInput
	}

	if !pkg.IsValidPassword(input.Password) {
		return nil, errors.New("password must be at least 8 characters")
	}

	user := domain.NewUser(input.Email, input.Password)

	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Send welcome email (non-blocking — failure doesn't affect registration)
	loginLink := uc.frontendBaseURL + "/login"
	if err := uc.emailService.SendWelcomeEmail(user.Email, loginLink); err != nil {
		log.Printf("[Register] error sending welcome email to %s: %v", user.Email, err)
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (uc *authUseCase) Login(ctx context.Context, input domain.LoginInput) (*domain.LoginResponse, error) {
	user, err := uc.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pkg.ErrUserNotFound) {
			return nil, pkg.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := user.ComparePassword(input.Password); err != nil {
		return nil, pkg.ErrInvalidCredentials
	}

	token, err := uc.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

// ForgotPassword generates a reset token and sends it by email
func (uc *authUseCase) ForgotPassword(ctx context.Context, input domain.ForgotPasswordInput) error {
	user, err := uc.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		// Always return nil to avoid email enumeration attacks
		return nil
	}

	token, err := generateResetToken()
	if err != nil {
		return pkg.ErrInternalServer
	}

	expiry := time.Now().Add(1 * time.Hour)

	if err := uc.repo.SaveResetToken(ctx, user.ID, token, expiry); err != nil {
		return err
	}

	resetLink := uc.frontendBaseURL + "/reset-password?token=" + token

	return uc.emailService.SendPasswordResetEmail(user.Email, resetLink)
}

// ResetPassword validates the token and updates the password
func (uc *authUseCase) ResetPassword(ctx context.Context, input domain.ResetPasswordInput) error {
	if !pkg.IsValidPassword(input.Password) {
		return errors.New("password must be at least 8 characters")
	}

	user, err := uc.repo.FindByResetToken(ctx, input.Token)
	if err != nil {
		if errors.Is(err, pkg.ErrUserNotFound) {
			return pkg.ErrInvalidToken
		}
		return err
	}

	if time.Now().After(user.ResetPasswordExpiry) {
		return pkg.ErrResetTokenExpired
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return uc.repo.UpdatePasswordAndClearToken(ctx, user.ID, string(hashed))
}

// GoogleAuthURL returns the Google OAuth consent page URL
func (uc *authUseCase) GoogleAuthURL(state string) string {
	return uc.googleOAuth.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// HandleGoogleCallback exchanges the auth code for a token, fetches the user info,
// and returns a JWT — creating or linking the account as needed.
func (uc *authUseCase) HandleGoogleCallback(ctx context.Context, code string) (*domain.LoginResponse, error) {
	token, err := uc.googleOAuth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google oauth exchange: %w", err)
	}

	googleUser, err := fetchGoogleUserInfo(ctx, uc.googleOAuth, token)
	if err != nil {
		return nil, err
	}

	user, err := uc.findOrProvisionGoogleUser(ctx, googleUser)
	if err != nil {
		return nil, err
	}

	jwtToken, err := uc.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &domain.LoginResponse{Token: jwtToken, User: *user}, nil
}

// findOrProvisionGoogleUser finds an existing user by Google ID or email,
// linking the Google account if needed, or creates a new one.
func (uc *authUseCase) findOrProvisionGoogleUser(ctx context.Context, googleUser *domain.GoogleUserInfo) (*domain.User, error) {
	// 1. Already linked with this Google ID
	user, err := uc.repo.FindByGoogleID(ctx, googleUser.ID)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, pkg.ErrUserNotFound) {
		return nil, err
	}

	// 2. Account exists with same email — link Google ID
	user, err = uc.repo.FindByEmail(ctx, googleUser.Email)
	if err == nil {
		if linkErr := uc.repo.LinkGoogleID(ctx, user.ID, googleUser.ID); linkErr != nil {
			log.Printf("[GoogleLogin] error linking google_id for user %s: %v", user.ID.Hex(), linkErr)
		}
		user.GoogleID = googleUser.ID
		return user, nil
	}
	if !errors.Is(err, pkg.ErrUserNotFound) {
		return nil, err
	}

	// 3. New user — create account
	newUser := domain.NewGoogleUser(googleUser.Email, googleUser.ID)
	if err := uc.repo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	loginLink := uc.frontendBaseURL + "/login"
	if err := uc.emailService.SendWelcomeEmail(newUser.Email, loginLink); err != nil {
		log.Printf("[GoogleLogin] error sending welcome email to %s: %v", newUser.Email, err)
	}

	return newUser, nil
}

// fetchGoogleUserInfo calls the Google userinfo API with the given OAuth token.
func fetchGoogleUserInfo(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (*domain.GoogleUserInfo, error) {
	client := cfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("fetch google userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read google userinfo body: %w", err)
	}

	var info domain.GoogleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse google userinfo: %w", err)
	}

	if info.ID == "" || info.Email == "" {
		return nil, pkg.ErrInvalidInput
	}

	return &info, nil
}

func generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

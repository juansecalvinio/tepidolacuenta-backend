package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"

	"juansecalvinio/tepidolacuenta/internal/auth/domain"
	"juansecalvinio/tepidolacuenta/internal/auth/usecase"
	"juansecalvinio/tepidolacuenta/internal/pkg"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	useCase         usecase.UseCase
	jwtSecret       string
	frontendBaseURL string
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(useCase usecase.UseCase, jwtSecret, frontendBaseURL string) *Handler {
	return &Handler{
		useCase:         useCase,
		jwtSecret:       jwtSecret,
		frontendBaseURL: frontendBaseURL,
	}
}

// Register handles user registration
// @Summary Register a new restaurant user
// @Tags auth
// @Accept json
// @Produce json
// @Param input body domain.RegisterInput true "Registration data"
// @Success 201 {object} pkg.Response{data=domain.User}
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var input domain.RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	user, err := h.useCase.Register(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, pkg.ErrUserAlreadyExists) {
			pkg.BadRequestResponse(c, "User with this email already exists", err)
			return
		}
		if errors.Is(err, pkg.ErrInvalidInput) {
			pkg.BadRequestResponse(c, "Invalid input data", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to register user", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "User registered successfully", user)
}

// Login handles user authentication
// @Summary Login a restaurant user
// @Tags auth
// @Accept json
// @Produce json
// @Param input body domain.LoginInput true "Login credentials"
// @Success 200 {object} pkg.Response{data=domain.LoginResponse}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var input domain.LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	loginResponse, err := h.useCase.Login(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, pkg.ErrInvalidCredentials) {
			pkg.UnauthorizedResponse(c, "Invalid email or password", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to login", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Login successful", loginResponse)
}

// ForgotPassword handles password reset requests
func (h *Handler) ForgotPassword(c *gin.Context) {
	var input domain.ForgotPasswordInput

	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	// Log error internally but always return success to avoid email enumeration
	if err := h.useCase.ForgotPassword(c.Request.Context(), input); err != nil {
		log.Printf("[ForgotPassword] error sending reset email to %s: %v", input.Email, err)
	}

	pkg.SuccessResponse(c, http.StatusOK, "If the email exists, a reset link has been sent", nil)
}

// ResetPassword handles password reset with token
func (h *Handler) ResetPassword(c *gin.Context) {
	var input domain.ResetPasswordInput

	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	if err := h.useCase.ResetPassword(c.Request.Context(), input); err != nil {
		if errors.Is(err, pkg.ErrInvalidToken) {
			pkg.BadRequestResponse(c, "Invalid or expired reset token", err)
			return
		}
		if errors.Is(err, pkg.ErrResetTokenExpired) {
			pkg.BadRequestResponse(c, "Reset token has expired", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to reset password", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Password reset successfully", nil)
}

// GoogleRedirect generates a signed CSRF state and redirects the user to Google's consent page.
func (h *Handler) GoogleRedirect(c *gin.Context) {
	state, err := generateOAuthState(h.jwtSecret)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "Failed to generate auth state", err)
		return
	}

	url := h.useCase.GoogleAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the redirect from Google, exchanges the code for a JWT,
// and redirects the user to the frontend with the token.
func (h *Handler) GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	if !verifyOAuthState(state, h.jwtSecret) {
		pkg.BadRequestResponse(c, "Invalid OAuth state", nil)
		return
	}

	code := c.Query("code")
	if code == "" {
		pkg.BadRequestResponse(c, "Missing authorization code", nil)
		return
	}

	loginResponse, err := h.useCase.HandleGoogleCallback(c.Request.Context(), code)
	if err != nil {
		log.Printf("[GoogleCallback] error: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.frontendBaseURL+"/login?error=google_auth_failed")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, h.frontendBaseURL+"/auth/callback?token="+loginResponse.Token)
}

// RegisterEmployee handles employee registration via invitation code
func (h *Handler) RegisterEmployee(c *gin.Context) {
	var input domain.EmployeeRegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	loginResponse, err := h.useCase.RegisterEmployee(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, pkg.ErrInvalidToken) {
			pkg.BadRequestResponse(c, "Invalid or expired invitation code", err)
			return
		}
		if errors.Is(err, pkg.ErrResetTokenExpired) {
			pkg.BadRequestResponse(c, "Invitation code has expired", err)
			return
		}
		if errors.Is(err, pkg.ErrUserAlreadyExists) {
			pkg.BadRequestResponse(c, "User with this email already exists", err)
			return
		}
		if errors.Is(err, pkg.ErrInvalidInput) {
			pkg.BadRequestResponse(c, "Invalid input data", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to register employee", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "Employee registered successfully", loginResponse)
}

// RegisterRoutes registers all auth routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/register/employee", h.RegisterEmployee)
		auth.POST("/login", h.Login)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/reset-password", h.ResetPassword)
		auth.GET("/google", h.GoogleRedirect)
		auth.GET("/google/callback", h.GoogleCallback)
	}
}

// generateOAuthState creates a CSRF-safe state token: <nonce>.<HMAC(nonce, secret)>
func generateOAuthState(secret string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(b)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(nonce))
	sig := hex.EncodeToString(mac.Sum(nil))
	return nonce + "." + sig, nil
}

// verifyOAuthState validates the HMAC signature of the state token.
func verifyOAuthState(state, secret string) bool {
	parts := strings.SplitN(state, ".", 2)
	if len(parts) != 2 {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0]))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(parts[1]))
}

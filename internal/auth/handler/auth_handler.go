package handler

import (
	"errors"
	"net/http"

	"juansecalvinio/tepidolacuenta/internal/auth/domain"
	"juansecalvinio/tepidolacuenta/internal/auth/usecase"
	"juansecalvinio/tepidolacuenta/internal/pkg"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	useCase usecase.UseCase
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(useCase usecase.UseCase) *Handler {
	return &Handler{
		useCase: useCase,
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

// RegisterRoutes registers all auth routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}
}

package handler

import (
	"errors"
	"net/http"

	"juansecalvinio/tepidolacuenta/internal/middleware"
	"juansecalvinio/tepidolacuenta/internal/pkg"
	"juansecalvinio/tepidolacuenta/internal/setup/domain"
	"juansecalvinio/tepidolacuenta/internal/setup/usecase"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	useCase usecase.UseCase
}

// NewSetupHandler creates a new setup handler
func NewSetupHandler(useCase usecase.UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

// SetupRestaurant handles the unified restaurant setup after registration
// @Summary Setup a new restaurant with branch and tables
// @Tags setup
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body domain.SetupRestaurantInput true "Setup data"
// @Success 201 {object} pkg.Response{data=domain.SetupRestaurantOutput}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/setup/restaurant [post]
func (h *Handler) SetupRestaurant(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		pkg.UnauthorizedResponse(c, "User not authenticated", pkg.ErrUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid user ID", err)
		return
	}

	var input domain.SetupRestaurantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	result, err := h.useCase.SetupRestaurant(c.Request.Context(), userID, input)
	if err != nil {
		if errors.Is(err, pkg.ErrInvalidInput) {
			pkg.BadRequestResponse(c, "Invalid input data", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to setup restaurant", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "Restaurant setup completed successfully", result)
}

// RegisterRoutes registers setup routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	setup := router.Group("/setup")
	{
		setup.POST("/restaurant", h.SetupRestaurant)
	}
}

package handler

import (
	"errors"
	"net/http"

	invitationDomain "juansecalvinio/tepidolacuenta/internal/invitation/domain"
	"juansecalvinio/tepidolacuenta/internal/invitation/usecase"
	"juansecalvinio/tepidolacuenta/internal/middleware"
	"juansecalvinio/tepidolacuenta/internal/pkg"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	useCase usecase.UseCase
}

func NewInvitationHandler(uc usecase.UseCase) *Handler {
	return &Handler{useCase: uc}
}

// GenerateCode creates a new single-use invitation code for a restaurant.
// POST /api/v1/invitations/generate
func (h *Handler) GenerateCode(c *gin.Context) {
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

	var input invitationDomain.GenerateInvitationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	restaurantID, err := primitive.ObjectIDFromHex(input.RestaurantID)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid restaurant ID", err)
		return
	}

	resp, err := h.useCase.Generate(c.Request.Context(), userID, restaurantID)
	if err != nil {
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't own this restaurant", err)
			return
		}
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant not found", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to generate invitation", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "Invitation code generated", resp)
}

// RegisterRoutes attaches the owner-protected generate endpoint.
func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	invitations := protected.Group("/invitations")
	invitations.Use(middleware.OwnerOnly())
	{
		invitations.POST("/generate", h.GenerateCode)
	}
}

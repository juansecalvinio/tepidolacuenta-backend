package handler

import (
	"errors"
	"net/http"

	authUseCase "juansecalvinio/tepidolacuenta/internal/auth/usecase"
	invitationDomain "juansecalvinio/tepidolacuenta/internal/invitation/domain"
	"juansecalvinio/tepidolacuenta/internal/invitation/usecase"
	"juansecalvinio/tepidolacuenta/internal/middleware"
	"juansecalvinio/tepidolacuenta/internal/pkg"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	useCase  usecase.UseCase
	authCase authUseCase.UseCase
}

func NewInvitationHandler(uc usecase.UseCase, authUC authUseCase.UseCase) *Handler {
	return &Handler{useCase: uc, authCase: authUC}
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

	branchID, err := primitive.ObjectIDFromHex(input.BranchID)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid branch ID", err)
		return
	}

	resp, err := h.useCase.Generate(c.Request.Context(), userID, restaurantID, branchID)
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

// AcceptInvitation links an authenticated Google user to a restaurant as an employee.
// POST /api/v1/invitations/accept
func (h *Handler) AcceptInvitation(c *gin.Context) {
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

	var input invitationDomain.AcceptInvitationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	resp, err := h.authCase.AcceptInvitation(c.Request.Context(), userID, input.Code)
	if err != nil {
		if errors.Is(err, pkg.ErrInvalidToken) || errors.Is(err, pkg.ErrResetTokenExpired) {
			pkg.BadRequestResponse(c, "Invalid or expired invitation code", err)
			return
		}
		if errors.Is(err, pkg.ErrUserNotFound) {
			pkg.NotFoundResponse(c, "User not found", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to accept invitation", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Invitation accepted", resp)
}

// RegisterRoutes attaches invitation endpoints to the protected router.
func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	invitations := protected.Group("/invitations")
	{
		invitations.POST("/accept", h.AcceptInvitation)
	}

	ownerInvitations := invitations.Group("")
	ownerInvitations.Use(middleware.OwnerOnly())
	{
		ownerInvitations.POST("/generate", h.GenerateCode)
	}
}

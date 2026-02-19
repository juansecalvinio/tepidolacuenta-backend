package handler

import (
	"errors"
	"net/http"

	"juansecalvinio/tepidolacuenta/internal/branch/domain"
	"juansecalvinio/tepidolacuenta/internal/branch/usecase"
	"juansecalvinio/tepidolacuenta/internal/middleware"
	"juansecalvinio/tepidolacuenta/internal/pkg"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	useCase usecase.UseCase
}

// NewBranchHandler creates a new branch handler
func NewBranchHandler(useCase usecase.UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

// Create handles branch creation
func (h *Handler) Create(c *gin.Context) {
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

	var input domain.CreateBranchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	branch, err := h.useCase.Create(c.Request.Context(), userID, input)
	if err != nil {
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this restaurant", err)
			return
		}
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant not found", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to create branch", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "Branch created successfully", branch)
}

// GetByID handles retrieving a branch by ID
func (h *Handler) GetByID(c *gin.Context) {
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

	branchIDStr := c.Param("id")
	branchID, err := primitive.ObjectIDFromHex(branchIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid branch ID", err)
		return
	}

	branch, err := h.useCase.GetByID(c.Request.Context(), branchID, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Branch not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this branch", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to get branch", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Branch retrieved successfully", branch)
}

// ListByRestaurant handles retrieving all branches for a restaurant
func (h *Handler) ListByRestaurant(c *gin.Context) {
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

	restaurantIDStr := c.Param("restaurantId")
	restaurantID, err := primitive.ObjectIDFromHex(restaurantIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid restaurant ID", err)
		return
	}

	branches, err := h.useCase.GetByRestaurantID(c.Request.Context(), restaurantID, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this restaurant", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to get branches", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Branches retrieved successfully", branches)
}

// Update handles branch updates
func (h *Handler) Update(c *gin.Context) {
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

	branchIDStr := c.Param("id")
	branchID, err := primitive.ObjectIDFromHex(branchIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid branch ID", err)
		return
	}

	var input domain.UpdateBranchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	branch, err := h.useCase.Update(c.Request.Context(), branchID, userID, input)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Branch not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this branch", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to update branch", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Branch updated successfully", branch)
}

// Delete handles branch deletion
func (h *Handler) Delete(c *gin.Context) {
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

	branchIDStr := c.Param("id")
	branchID, err := primitive.ObjectIDFromHex(branchIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid branch ID", err)
		return
	}

	if err := h.useCase.Delete(c.Request.Context(), branchID, userID); err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Branch not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this branch", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to delete branch", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Branch deleted successfully", nil)
}

// RegisterRoutes registers all branch routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	branches := router.Group("/branches")
	{
		branches.POST("", h.Create)
		branches.GET("/:id", h.GetByID)
		branches.GET("/restaurant/:restaurantId", h.ListByRestaurant)
		branches.PUT("/:id", h.Update)
		branches.DELETE("/:id", h.Delete)
	}
}

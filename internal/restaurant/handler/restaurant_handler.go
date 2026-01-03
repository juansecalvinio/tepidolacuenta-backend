package handler

import (
	"errors"
	"net/http"

	"juansecalvinio/tepidolacuenta/internal/middleware"
	"juansecalvinio/tepidolacuenta/internal/pkg"
	"juansecalvinio/tepidolacuenta/internal/restaurant/domain"
	"juansecalvinio/tepidolacuenta/internal/restaurant/usecase"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	useCase usecase.UseCase
}

// NewRestaurantHandler creates a new restaurant handler
func NewRestaurantHandler(useCase usecase.UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

// Create handles restaurant creation
// @Summary Create a new restaurant
// @Tags restaurants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body domain.CreateRestaurantInput true "Restaurant data"
// @Success 201 {object} pkg.Response{data=domain.Restaurant}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/restaurants [post]
func (h *Handler) Create(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
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

	var input domain.CreateRestaurantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	restaurant, err := h.useCase.Create(c.Request.Context(), userID, input)
	if err != nil {
		if errors.Is(err, pkg.ErrInvalidInput) {
			pkg.BadRequestResponse(c, "Invalid input data", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to create restaurant", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "Restaurant created successfully", restaurant)
}

// GetByID handles retrieving a restaurant by ID
// @Summary Get restaurant by ID
// @Tags restaurants
// @Produce json
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Success 200 {object} pkg.Response{data=domain.Restaurant}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/restaurants/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	// Get user ID from context
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

	// Get restaurant ID from URL
	restaurantIDStr := c.Param("id")
	restaurantID, err := primitive.ObjectIDFromHex(restaurantIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid restaurant ID", err)
		return
	}

	restaurant, err := h.useCase.GetByID(c.Request.Context(), restaurantID, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this restaurant", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to get restaurant", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Restaurant retrieved successfully", restaurant)
}

// List handles retrieving all restaurants for the authenticated user
// @Summary List user's restaurants
// @Tags restaurants
// @Produce json
// @Security BearerAuth
// @Success 200 {object} pkg.Response{data=[]domain.Restaurant}
// @Failure 401 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/restaurants [get]
func (h *Handler) List(c *gin.Context) {
	// Get user ID from context
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

	restaurants, err := h.useCase.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "Failed to get restaurants", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Restaurants retrieved successfully", restaurants)
}

// Update handles restaurant updates
// @Summary Update restaurant
// @Tags restaurants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Param input body domain.UpdateRestaurantInput true "Update data"
// @Success 200 {object} pkg.Response{data=domain.Restaurant}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/restaurants/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	// Get user ID from context
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

	// Get restaurant ID from URL
	restaurantIDStr := c.Param("id")
	restaurantID, err := primitive.ObjectIDFromHex(restaurantIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid restaurant ID", err)
		return
	}

	var input domain.UpdateRestaurantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	restaurant, err := h.useCase.Update(c.Request.Context(), restaurantID, userID, input)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this restaurant", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to update restaurant", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Restaurant updated successfully", restaurant)
}

// Delete handles restaurant deletion
// @Summary Delete restaurant
// @Tags restaurants
// @Produce json
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/restaurants/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	// Get user ID from context
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

	// Get restaurant ID from URL
	restaurantIDStr := c.Param("id")
	restaurantID, err := primitive.ObjectIDFromHex(restaurantIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid restaurant ID", err)
		return
	}

	if err := h.useCase.Delete(c.Request.Context(), restaurantID, userID); err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this restaurant", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to delete restaurant", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Restaurant deleted successfully", nil)
}

// RegisterRoutes registers all restaurant routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	restaurants := router.Group("/restaurants")
	{
		restaurants.POST("", h.Create)
		restaurants.GET("", h.List)
		restaurants.GET("/:id", h.GetByID)
		restaurants.PUT("/:id", h.Update)
		restaurants.DELETE("/:id", h.Delete)
	}
}

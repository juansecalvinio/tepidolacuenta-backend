package handler

import (
	"errors"
	"net/http"

	"juansecalvinio/tepidolacuenta/internal/middleware"
	"juansecalvinio/tepidolacuenta/internal/pkg"
	"juansecalvinio/tepidolacuenta/internal/subscription/domain"
	"juansecalvinio/tepidolacuenta/internal/subscription/usecase"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	useCase usecase.UseCase
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(useCase usecase.UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

// GetAllPlans handles retrieving all available plans
func (h *Handler) GetAllPlans(c *gin.Context) {
	plans, err := h.useCase.GetAllPlans(c.Request.Context())
	if err != nil {
		pkg.InternalServerErrorResponse(c, "Failed to get plans", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Plans retrieved successfully", plans)
}

// GetPlanByID handles retrieving a plan by ID
func (h *Handler) GetPlanByID(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := primitive.ObjectIDFromHex(planIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid plan ID", err)
		return
	}

	plan, err := h.useCase.GetPlanByID(c.Request.Context(), planID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Plan not found", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to get plan", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Plan retrieved successfully", plan)
}

// Create handles subscription creation
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

	var input domain.CreateSubscriptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	subscription, err := h.useCase.Create(c.Request.Context(), userID, input)
	if err != nil {
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this restaurant", err)
			return
		}
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant or plan not found", err)
			return
		}
		if errors.Is(err, pkg.ErrSubscriptionAlreadyExists) {
			pkg.BadRequestResponse(c, "Restaurant already has an active subscription", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to create subscription", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "Subscription created successfully", subscription)
}

// GetByID handles retrieving a subscription by ID
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

	subscriptionIDStr := c.Param("id")
	subscriptionID, err := primitive.ObjectIDFromHex(subscriptionIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid subscription ID", err)
		return
	}

	subscription, err := h.useCase.GetByID(c.Request.Context(), subscriptionID, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Subscription not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this subscription", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to get subscription", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Subscription retrieved successfully", subscription)
}

// GetByUser handles retrieving all subscriptions for the authenticated user
func (h *Handler) GetByUser(c *gin.Context) {
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

	subscriptions, err := h.useCase.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "Failed to get subscriptions", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Subscriptions retrieved successfully", subscriptions)
}

// GetByRestaurant handles retrieving the subscription for a restaurant
func (h *Handler) GetByRestaurant(c *gin.Context) {
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

	subscription, err := h.useCase.GetByRestaurantID(c.Request.Context(), restaurantID, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Subscription not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this restaurant", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to get subscription", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Subscription retrieved successfully", subscription)
}

// Update handles subscription updates
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

	subscriptionIDStr := c.Param("id")
	subscriptionID, err := primitive.ObjectIDFromHex(subscriptionIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid subscription ID", err)
		return
	}

	var input domain.UpdateSubscriptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	subscription, err := h.useCase.Update(c.Request.Context(), subscriptionID, userID, input)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Subscription or plan not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this subscription", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to update subscription", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Subscription updated successfully", subscription)
}

// Cancel handles subscription cancellation
func (h *Handler) Cancel(c *gin.Context) {
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

	subscriptionIDStr := c.Param("id")
	subscriptionID, err := primitive.ObjectIDFromHex(subscriptionIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid subscription ID", err)
		return
	}

	if err := h.useCase.Cancel(c.Request.Context(), subscriptionID, userID); err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Subscription not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this subscription", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to cancel subscription", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Subscription canceled successfully", nil)
}

// RegisterRoutes registers all subscription routes
func (h *Handler) RegisterRoutes(protected *gin.RouterGroup, public *gin.RouterGroup) {
	// Public plan routes (no auth required)
	publicPlans := public.Group("/plans")
	{
		publicPlans.GET("", h.GetAllPlans)
		publicPlans.GET("/:id", h.GetPlanByID)
	}

	// Protected plan routes
	plans := protected.Group("/plans")
	{
		plans.GET("", h.GetAllPlans)
		plans.GET("/:id", h.GetPlanByID)
	}

	// Protected subscription routes
	subscriptions := protected.Group("/subscriptions")
	{
		subscriptions.POST("", h.Create)
		subscriptions.GET("", h.GetByUser)
		subscriptions.GET("/:id", h.GetByID)
		subscriptions.GET("/restaurant/:restaurantId", h.GetByRestaurant)
		subscriptions.PUT("/:id", h.Update)
		subscriptions.DELETE("/:id", h.Cancel)
	}
}

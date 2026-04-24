package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"juansecalvinio/tepidolacuenta/internal/middleware"
	"juansecalvinio/tepidolacuenta/internal/payment/domain"
	"juansecalvinio/tepidolacuenta/internal/payment/usecase"
	"juansecalvinio/tepidolacuenta/internal/pkg"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	useCase usecase.UseCase
}

func NewPaymentHandler(useCase usecase.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

// CreatePreference handles POST /api/v1/payments/create-preference
func (h *Handler) CreatePreference(c *gin.Context) {
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

	var input domain.CreatePreferenceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	result, err := h.useCase.CreatePaymentPreference(c.Request.Context(), userID, input)
	if err != nil {
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.ForbiddenResponse(c, "You don't have access to this restaurant", err)
			return
		}
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant or plan not found", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to create payment preference", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":      true,
		"paymentUrl":   result.PaymentURL,
		"preferenceId": result.PreferenceID,
	})
}

// ProcessWebhook handles POST /api/v1/payments/webhook (no auth — called by MercadoPago)
func (h *Handler) ProcessWebhook(c *gin.Context) {
	xSignature := c.GetHeader("x-signature")
	xRequestID := c.GetHeader("x-request-id")

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		pkg.BadRequestResponse(c, "Failed to read request body", err)
		return
	}

	var body domain.WebhookBody
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		pkg.BadRequestResponse(c, "Invalid webhook body", err)
		return
	}

	if body.Data.ID == "" || body.Action == "" {
		c.Status(http.StatusOK)
		return
	}

	// Only process payment notifications
	if body.Action != "payment.created" && body.Action != "payment.updated" {
		c.Status(http.StatusOK)
		return
	}

	if err := h.useCase.ProcessPaymentWebhook(c.Request.Context(), body.Data.ID, xSignature, xRequestID); err != nil {
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.BadRequestResponse(c, "Invalid webhook signature", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to process webhook", err)
		return
	}

	c.Status(http.StatusOK)
}

// GetByID handles GET /api/v1/payments/:id
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

	paymentID := c.Param("id")
	if paymentID == "" {
		pkg.BadRequestResponse(c, "Payment ID is required", nil)
		return
	}

	payment, err := h.useCase.GetPaymentByID(c.Request.Context(), userID, paymentID)
	if err != nil {
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.ForbiddenResponse(c, "You don't have access to this payment", err)
			return
		}
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Payment not found", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to get payment", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Payment retrieved successfully", payment)
}

// GetHistory handles GET /api/v1/payments/history?restaurantId=...
func (h *Handler) GetHistory(c *gin.Context) {
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

	restaurantID := c.Query("restaurantId")
	if restaurantID == "" {
		pkg.BadRequestResponse(c, "restaurantId query param is required", nil)
		return
	}

	payments, err := h.useCase.GetPaymentHistory(c.Request.Context(), userID, restaurantID)
	if err != nil {
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.ForbiddenResponse(c, "You don't have access to this restaurant", err)
			return
		}
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant not found", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to get payment history", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Payment history retrieved successfully", payments)
}

// RegisterRoutes registers all payment routes
func (h *Handler) RegisterRoutes(protected *gin.RouterGroup, public *gin.RouterGroup) {
	payments := protected.Group("/payments")
	{
		payments.POST("/create-preference", h.CreatePreference)
		payments.GET("/:id", h.GetByID)
		payments.GET("/history", h.GetHistory)
	}

	// Webhook is public — MercadoPago calls it without auth
	public.POST("/payments/webhook", h.ProcessWebhook)
}

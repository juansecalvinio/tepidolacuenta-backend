package handler

import (
	"errors"
	"net/http"

	"juansecalvinio/tepidolacuenta/internal/middleware"
	"juansecalvinio/tepidolacuenta/internal/pkg"
	"juansecalvinio/tepidolacuenta/internal/request/domain"
	"juansecalvinio/tepidolacuenta/internal/request/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, you should check the origin properly
		return true
	},
}

type Handler struct {
	useCase usecase.UseCase
	hub     *pkg.Hub
}

// NewRequestHandler creates a new request handler
func NewRequestHandler(useCase usecase.UseCase, hub *pkg.Hub) *Handler {
	return &Handler{
		useCase: useCase,
		hub:     hub,
	}
}

// Create handles request creation (public endpoint for clients)
// @Summary Create a new account request from QR code
// @Tags requests
// @Accept json
// @Produce json
// @Param input body domain.CreateRequestInput true \"Request data\"
// @Success 201 {object} pkg.Response{data=domain.Request}
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/public/request-account [post]
func (h *Handler) Create(c *gin.Context) {
	var input domain.CreateRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	request, err := h.useCase.Create(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant or table not found", err)
			return
		}
		pkg.BadRequestResponse(c, "Failed to create request", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "Request created successfully", request)
}

// GetByID handles retrieving a request by ID
// @Summary Get request by ID
// @Tags requests
// @Produce json
// @Security BearerAuth
// @Param id path string true \"Request ID\"
// @Success 200 {object} pkg.Response{data=domain.Request}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/requests/{id} [get]
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

	requestIDStr := c.Param("id")
	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid request ID", err)
		return
	}

	request, err := h.useCase.GetByID(c.Request.Context(), requestID, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Request not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this request", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to get request", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Request retrieved successfully", request)
}

// ListByRestaurant handles retrieving all requests for a restaurant
// @Summary List requests by restaurant
// @Tags requests
// @Produce json
// @Security BearerAuth
// @Param restaurantId path string true \"Restaurant ID\"
// @Success 200 {object} pkg.Response{data=[]domain.Request}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/requests/restaurant/{restaurantId} [get]
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

	requests, err := h.useCase.GetByRestaurantID(c.Request.Context(), restaurantID, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this restaurant", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to get requests", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Requests retrieved successfully", requests)
}

// ListPendingByRestaurant handles retrieving pending requests for a restaurant
// @Summary List pending requests by restaurant
// @Tags requests
// @Produce json
// @Security BearerAuth
// @Param restaurantId path string true \"Restaurant ID\"
// @Success 200 {object} pkg.Response{data=[]domain.Request}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/requests/restaurant/{restaurantId}/pending [get]
func (h *Handler) ListPendingByRestaurant(c *gin.Context) {
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

	requests, err := h.useCase.GetPendingByRestaurantID(c.Request.Context(), restaurantID, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this restaurant", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to get pending requests", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Pending requests retrieved successfully", requests)
}

// UpdateStatus handles updating request status
// @Summary Update request status
// @Tags requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true \"Request ID\"
// @Param input body domain.UpdateRequestStatusInput true \"Status data\"
// @Success 200 {object} pkg.Response{data=domain.Request}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/requests/{id}/status [put]
func (h *Handler) UpdateStatus(c *gin.Context) {
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

	requestIDStr := c.Param("id")
	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid request ID", err)
		return
	}

	var input domain.UpdateRequestStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	request, err := h.useCase.UpdateStatus(c.Request.Context(), requestID, userID, input)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Request not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this request", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to update request status", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Request status updated successfully", request)
}

// Delete handles request deletion
// @Summary Delete request
// @Tags requests
// @Produce json
// @Security BearerAuth
// @Param id path string true \"Request ID\"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/requests/{id} [delete]
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

	requestIDStr := c.Param("id")
	requestID, err := primitive.ObjectIDFromHex(requestIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid request ID", err)
		return
	}

	if err := h.useCase.Delete(c.Request.Context(), requestID, userID); err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Request not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this request", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to delete request", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Request deleted successfully", nil)
}

// WebSocket handles WebSocket connections for real-time notifications
// @Summary WebSocket endpoint for real-time notifications
// @Tags requests
// @Security BearerAuth
// @Param restaurantId path string true \"Restaurant ID\"
// @Router /api/v1/requests/ws/{restaurantId} [get]
func (h *Handler) WebSocket(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		pkg.UnauthorizedResponse(c, "User not authenticated", pkg.ErrUnauthorized)
		return
	}

	restaurantIDStr := c.Param("restaurantId")
	restaurantID, err := primitive.ObjectIDFromHex(restaurantIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid restaurant ID", err)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "Failed to upgrade connection", err)
		return
	}

	// Create client
	client := &pkg.Client{
		ID:           uuid.New().String(),
		RestaurantID: restaurantID,
		Conn:         conn,
		Send:         make(chan []byte, 256),
	}

	// Register client
	h.hub.Register(client)

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump(h.hub)
}

// RegisterRoutes registers all request routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, publicRouter *gin.RouterGroup) {
	// Public routes (no authentication required)
	publicRouter.POST("/request-account", h.Create)

	// Protected routes (authentication required)
	requests := router.Group("/requests")
	{
		requests.GET("/:id", h.GetByID)
		requests.GET("/restaurant/:restaurantId", h.ListByRestaurant)
		requests.GET("/restaurant/:restaurantId/pending", h.ListPendingByRestaurant)
		requests.PUT("/:id/status", h.UpdateStatus)
		requests.DELETE("/:id", h.Delete)
		requests.GET("/ws/:restaurantId", h.WebSocket)
	}
}

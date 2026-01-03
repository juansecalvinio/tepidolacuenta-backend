package handler

import (
	"errors"
	"net/http"

	"juansecalvinio/tepidolacuenta/internal/middleware"
	"juansecalvinio/tepidolacuenta/internal/pkg"
	"juansecalvinio/tepidolacuenta/internal/table/domain"
	"juansecalvinio/tepidolacuenta/internal/table/usecase"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	useCase usecase.UseCase
}

// NewTableHandler creates a new table handler
func NewTableHandler(useCase usecase.UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

// Create handles table creation
// @Summary Create a new table
// @Tags tables
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body domain.CreateTableInput true "Table data"
// @Success 201 {object} pkg.Response{data=domain.Table}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/tables [post]
func (h *Handler) Create(c *gin.Context) {
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

	var input domain.CreateTableInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	table, err := h.useCase.Create(c.Request.Context(), userID, input)
	if err != nil {
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this restaurant", err)
			return
		}
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant not found", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to create table", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "Table created successfully", table)
}

// GetByID handles retrieving a table by ID
// @Summary Get table by ID
// @Tags tables
// @Produce json
// @Security BearerAuth
// @Param id path string true "Table ID"
// @Success 200 {object} pkg.Response{data=domain.Table}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/tables/{id} [get]
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

	// Get table ID from URL
	tableIDStr := c.Param("id")
	tableID, err := primitive.ObjectIDFromHex(tableIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid table ID", err)
		return
	}

	table, err := h.useCase.GetByID(c.Request.Context(), tableID, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Table not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this table", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to get table", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Table retrieved successfully", table)
}

// ListByRestaurant handles retrieving all tables for a restaurant
// @Summary List tables by restaurant
// @Tags tables
// @Produce json
// @Security BearerAuth
// @Param restaurantId path string true "Restaurant ID"
// @Success 200 {object} pkg.Response{data=[]domain.Table}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/tables/restaurant/{restaurantId} [get]
func (h *Handler) ListByRestaurant(c *gin.Context) {
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
	restaurantIDStr := c.Param("restaurantId")
	restaurantID, err := primitive.ObjectIDFromHex(restaurantIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid restaurant ID", err)
		return
	}

	tables, err := h.useCase.GetByRestaurantID(c.Request.Context(), restaurantID, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Restaurant not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this restaurant", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to get tables", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Tables retrieved successfully", tables)
}

// Update handles table updates
// @Summary Update table
// @Tags tables
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Table ID"
// @Param input body domain.UpdateTableInput true "Update data"
// @Success 200 {object} pkg.Response{data=domain.Table}
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/tables/{id} [put]
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

	// Get table ID from URL
	tableIDStr := c.Param("id")
	tableID, err := primitive.ObjectIDFromHex(tableIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid table ID", err)
		return
	}

	var input domain.UpdateTableInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.BadRequestResponse(c, "Invalid input", err)
		return
	}

	table, err := h.useCase.Update(c.Request.Context(), tableID, userID, input)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Table not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this table", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to update table", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Table updated successfully", table)
}

// Delete handles table deletion
// @Summary Delete table
// @Tags tables
// @Produce json
// @Security BearerAuth
// @Param id path string true "Table ID"
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Failure 401 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /api/v1/tables/{id} [delete]
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

	// Get table ID from URL
	tableIDStr := c.Param("id")
	tableID, err := primitive.ObjectIDFromHex(tableIDStr)
	if err != nil {
		pkg.BadRequestResponse(c, "Invalid table ID", err)
		return
	}

	if err := h.useCase.Delete(c.Request.Context(), tableID, userID); err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.NotFoundResponse(c, "Table not found", err)
			return
		}
		if errors.Is(err, pkg.ErrUnauthorized) {
			pkg.UnauthorizedResponse(c, "You don't have access to this table", err)
			return
		}
		pkg.InternalServerErrorResponse(c, "Failed to delete table", err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "Table deleted successfully", nil)
}

// RegisterRoutes registers all table routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	tables := router.Group("/tables")
	{
		tables.POST("", h.Create)
		tables.GET("/:id", h.GetByID)
		tables.GET("/restaurant/:restaurantId", h.ListByRestaurant)
		tables.PUT("/:id", h.Update)
		tables.DELETE("/:id", h.Delete)
	}
}

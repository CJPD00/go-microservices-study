package infrastructure

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-micro/internal/orders/application"
	"go-micro/pkg/errors"
	"go-micro/pkg/middleware"
)

// HTTPHandler handles HTTP requests for orders
type HTTPHandler struct {
	useCase *application.OrderUseCase
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(useCase *application.OrderUseCase) *HTTPHandler {
	return &HTTPHandler{useCase: useCase}
}

// RegisterRoutes registers the order routes
func (h *HTTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	orders := r.Group("/orders")
	{
		orders.POST("", h.CreateOrder)
		orders.GET("/:id", h.GetOrder)
	}
}

// CreateOrderRequest is the request body for creating an order
type CreateOrderRequest struct {
	UserID uint    `json:"user_id" binding:"required"`
	Total  float64 `json:"total" binding:"required,gt=0"`
}

// OrderResponse is the response body for order operations
type OrderResponse struct {
	ID        uint    `json:"id"`
	UserID    uint    `json:"user_id"`
	Total     float64 `json:"total"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

// CreateOrder handles POST /orders
func (h *HTTPHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewValidation("invalid request body", err.Error()))
		return
	}

	output, err := h.useCase.CreateOrder(c.Request.Context(), application.CreateOrderInput{
		UserID: req.UserID,
		Total:  req.Total,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": OrderResponse{
			ID:        output.Order.ID,
			UserID:    output.Order.UserID,
			Total:     output.Order.Total,
			Status:    string(output.Order.Status),
			CreatedAt: output.Order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		"trace_id": c.GetString(middleware.TraceIDKey),
	})
}

// GetOrder handles GET /orders/:id
func (h *HTTPHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Error(errors.NewValidation("invalid order id", nil))
		return
	}

	output, err := h.useCase.GetOrder(c.Request.Context(), application.GetOrderInput{
		ID: uint(id),
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": OrderResponse{
			ID:        output.Order.ID,
			UserID:    output.Order.UserID,
			Total:     output.Order.Total,
			Status:    string(output.Order.Status),
			CreatedAt: output.Order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		"trace_id": c.GetString(middleware.TraceIDKey),
	})
}

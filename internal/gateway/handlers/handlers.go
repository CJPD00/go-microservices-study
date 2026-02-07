package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	orderspb "go-micro/api/gen/orders/v1"
	userspb "go-micro/api/gen/users/v1"
	"go-micro/pkg/errors"
	"go-micro/pkg/middleware"
)

// Handler handles all gateway HTTP requests
type Handler struct {
	usersClient  userspb.UserServiceClient
	ordersClient orderspb.OrderServiceClient
}

// NewHandler creates a new gateway handler
func NewHandler(usersClient userspb.UserServiceClient, ordersClient orderspb.OrderServiceClient) *Handler {
	return &Handler{
		usersClient:  usersClient,
		ordersClient: ordersClient,
	}
}

// RegisterRoutes registers all gateway routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Users endpoints
	users := r.Group("/users")
	{
		users.POST("", h.CreateUser)
		users.GET("/:id", h.GetUser)
	}

	// Orders endpoints
	orders := r.Group("/orders")
	{
		orders.POST("", h.CreateOrder)
		orders.GET("/:id", h.GetOrder)
	}
}

// =============================================================================
// Request/Response DTOs
// =============================================================================

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required" example:"John Doe"`
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
}

// UserResponse represents a user in responses
type UserResponse struct {
	ID        uint   `json:"id" example:"1"`
	Name      string `json:"name" example:"John Doe"`
	Email     string `json:"email" example:"john@example.com"`
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// CreateOrderRequest represents the request body for creating an order
type CreateOrderRequest struct {
	UserID uint    `json:"user_id" binding:"required" example:"1"`
	Total  float64 `json:"total" binding:"required,gt=0" example:"99.99"`
}

// OrderResponse represents an order in responses
type OrderResponse struct {
	ID        uint    `json:"id" example:"1"`
	UserID    uint    `json:"user_id" example:"1"`
	Total     float64 `json:"total" example:"99.99"`
	Status    string  `json:"status" example:"pending"`
	CreatedAt string  `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// SuccessResponse is the standard success response
type SuccessResponse struct {
	Data    interface{} `json:"data"`
	TraceID string      `json:"trace_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ErrorResponse is the standard error response
type ErrorResponse struct {
	Error   ErrorBody `json:"error"`
	TraceID string    `json:"trace_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ErrorBody contains error details
type ErrorBody struct {
	Code    string      `json:"code" example:"VALIDATION_ERROR"`
	Message string      `json:"message" example:"Invalid request body"`
	Details interface{} `json:"details,omitempty"`
}

// =============================================================================
// Users Handlers
// =============================================================================

// CreateUser creates a new user
// @Summary Create a new user
// @Description Create a new user with name and email
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User creation request"
// @Success 201 {object} SuccessResponse{data=UserResponse} "User created successfully"
// @Failure 400 {object} ErrorResponse "Validation error"
// @Failure 409 {object} ErrorResponse "Email already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewValidation("invalid request body", err.Error()))
		return
	}

	resp, err := h.usersClient.CreateUser(c.Request.Context(), &userspb.CreateUserRequest{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		c.Error(errors.FromGRPCStatus(err))
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Data: UserResponse{
			ID:        uint(resp.GetId()),
			Name:      resp.GetName(),
			Email:     resp.GetEmail(),
			CreatedAt: resp.GetCreatedAt(),
		},
		TraceID: c.GetString(middleware.TraceIDKey),
	})
}

// GetUser retrieves a user by ID
// @Summary Get a user by ID
// @Description Retrieve user details by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} SuccessResponse{data=UserResponse} "User retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Error(errors.NewValidation("invalid user id", nil))
		return
	}

	resp, err := h.usersClient.GetUser(c.Request.Context(), &userspb.GetUserRequest{
		Id: id,
	})
	if err != nil {
		c.Error(errors.FromGRPCStatus(err))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data: UserResponse{
			ID:        uint(resp.GetId()),
			Name:      resp.GetName(),
			Email:     resp.GetEmail(),
			CreatedAt: resp.GetCreatedAt(),
		},
		TraceID: c.GetString(middleware.TraceIDKey),
	})
}

// =============================================================================
// Orders Handlers
// =============================================================================

// CreateOrder creates a new order
// @Summary Create a new order
// @Description Create a new order for a user
// @Tags orders
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "Order creation request"
// @Success 201 {object} SuccessResponse{data=OrderResponse} "Order created successfully"
// @Failure 400 {object} ErrorResponse "Validation error (including user not found)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewValidation("invalid request body", err.Error()))
		return
	}

	resp, err := h.ordersClient.CreateOrder(c.Request.Context(), &orderspb.CreateOrderRequest{
		UserId: uint64(req.UserID),
		Total:  req.Total,
	})
	if err != nil {
		c.Error(errors.FromGRPCStatus(err))
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Data: OrderResponse{
			ID:        uint(resp.GetId()),
			UserID:    uint(resp.GetUserId()),
			Total:     resp.GetTotal(),
			Status:    resp.GetStatus(),
			CreatedAt: resp.GetCreatedAt(),
		},
		TraceID: c.GetString(middleware.TraceIDKey),
	})
}

// GetOrder retrieves an order by ID
// @Summary Get an order by ID
// @Description Retrieve order details by its ID
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} SuccessResponse{data=OrderResponse} "Order retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid order ID"
// @Failure 404 {object} ErrorResponse "Order not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/orders/{id} [get]
func (h *Handler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Error(errors.NewValidation("invalid order id", nil))
		return
	}

	resp, err := h.ordersClient.GetOrder(c.Request.Context(), &orderspb.GetOrderRequest{
		Id: id,
	})
	if err != nil {
		c.Error(errors.FromGRPCStatus(err))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data: OrderResponse{
			ID:        uint(resp.GetId()),
			UserID:    uint(resp.GetUserId()),
			Total:     resp.GetTotal(),
			Status:    resp.GetStatus(),
			CreatedAt: resp.GetCreatedAt(),
		},
		TraceID: c.GetString(middleware.TraceIDKey),
	})
}

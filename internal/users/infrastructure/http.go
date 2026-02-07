package infrastructure

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-micro/internal/users/application"
	"go-micro/pkg/errors"
	"go-micro/pkg/middleware"
)

// HTTPHandler handles HTTP requests for users
type HTTPHandler struct {
	useCase *application.UserUseCase
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(useCase *application.UserUseCase) *HTTPHandler {
	return &HTTPHandler{useCase: useCase}
}

// RegisterRoutes registers the user routes
func (h *HTTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		users.POST("", h.CreateUser)
		users.GET("/:id", h.GetUser)
	}
}

// CreateUserRequest is the request body for creating a user
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// UserResponse is the response body for user operations
type UserResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// CreateUser handles POST /users
func (h *HTTPHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewValidation("invalid request body", err.Error()))
		return
	}

	output, err := h.useCase.CreateUser(c.Request.Context(), application.CreateUserInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": UserResponse{
			ID:        output.User.ID,
			Name:      output.User.Name,
			Email:     output.User.Email,
			CreatedAt: output.User.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		"trace_id": c.GetString(middleware.TraceIDKey),
	})
}

// GetUser handles GET /users/:id
func (h *HTTPHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Error(errors.NewValidation("invalid user id", nil))
		return
	}

	output, err := h.useCase.GetUser(c.Request.Context(), application.GetUserInput{
		ID: uint(id),
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": UserResponse{
			ID:        output.User.ID,
			Name:      output.User.Name,
			Email:     output.User.Email,
			CreatedAt: output.User.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		"trace_id": c.GetString(middleware.TraceIDKey),
	})
}

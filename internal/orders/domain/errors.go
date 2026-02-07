package domain

import "go-micro/pkg/errors"

// Domain-specific errors
var (
	ErrUserIDRequired = errors.NewValidation("user_id is required", nil)
	ErrInvalidTotal   = errors.NewValidation("total must be greater than 0", nil)
	ErrTotalTooHigh   = errors.NewValidation("total cannot exceed 1,000,000", nil)
	ErrOrderNotFound  = errors.NewNotFound("order", "unknown")
	ErrUserNotFound   = errors.NewNotFound("user", "unknown")
)

// NewOrderNotFound creates a not found error with the order ID
func NewOrderNotFound(id uint) error {
	return errors.NewNotFound("order", id)
}

// NewUserNotFoundError creates a not found error for user validation
func NewUserNotFoundError(userID uint) error {
	return errors.NewValidation("user not found", map[string]interface{}{
		"user_id": userID,
	})
}

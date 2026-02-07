package domain

import "go-micro/pkg/errors"

// Domain-specific errors
var (
	ErrNameRequired  = errors.NewValidation("name is required", nil)
	ErrNameLength    = errors.NewValidation("name must be between 2 and 100 characters", nil)
	ErrEmailRequired = errors.NewValidation("email is required", nil)
	ErrEmailInvalid  = errors.NewValidation("email format is invalid", nil)
	ErrEmailExists   = errors.NewConflict("email already exists")
	ErrUserNotFound  = errors.NewNotFound("user", "unknown")
)

// NewUserNotFound creates a not found error with the user ID
func NewUserNotFound(id uint) error {
	return errors.NewNotFound("user", id)
}

package ports

import (
	"context"

	"go-micro/internal/users/domain"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *domain.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uint) (*domain.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *domain.User) error

	// Delete deletes a user by ID
	Delete(ctx context.Context, id uint) error
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	// PublishUserCreated publishes a user created event
	PublishUserCreated(ctx context.Context, user *domain.User) error
}

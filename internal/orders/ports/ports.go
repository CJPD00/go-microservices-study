package ports

import (
	"context"

	"go-micro/internal/orders/domain"
)

// OrderRepository defines the interface for order persistence
type OrderRepository interface {
	// Create creates a new order
	Create(ctx context.Context, order *domain.Order) error

	// GetByID retrieves an order by ID
	GetByID(ctx context.Context, id uint) (*domain.Order, error)

	// Update updates an existing order
	Update(ctx context.Context, order *domain.Order) error

	// Delete deletes an order by ID
	Delete(ctx context.Context, id uint) error

	// GetByUserID retrieves orders for a user
	GetByUserID(ctx context.Context, userID uint) ([]*domain.Order, error)
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	// PublishOrderCreated publishes an order created event
	PublishOrderCreated(ctx context.Context, order *domain.Order) error
}

// UserClient defines the interface for user service communication
type UserClient interface {
	// GetUser retrieves a user by ID (validates user exists)
	GetUser(ctx context.Context, userID uint) (*UserInfo, error)
}

// UserInfo represents user information from the users service
type UserInfo struct {
	ID    uint
	Name  string
	Email string
}

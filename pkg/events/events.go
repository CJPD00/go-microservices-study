package events

import "time"

// Exchange names
const (
	ExchangeUsers  = "users.events"
	ExchangeOrders = "orders.events"
)

// Routing keys
const (
	RoutingKeyUserCreated  = "user.created"
	RoutingKeyOrderCreated = "order.created"
)

// UserCreatedEvent is published when a user is created
type UserCreatedEvent struct {
	Version   string             `json:"version"`
	EventType string             `json:"event_type"`
	Timestamp time.Time          `json:"timestamp"`
	TraceID   string             `json:"trace_id"`
	Payload   UserCreatedPayload `json:"payload"`
}

// UserCreatedPayload contains user data
type UserCreatedPayload struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// NewUserCreatedEvent creates a new UserCreatedEvent
func NewUserCreatedEvent(id uint, name, email string, createdAt time.Time, traceID string) *UserCreatedEvent {
	return &UserCreatedEvent{
		Version:   "1.0",
		EventType: "user.created",
		Timestamp: time.Now(),
		TraceID:   traceID,
		Payload: UserCreatedPayload{
			ID:        id,
			Name:      name,
			Email:     email,
			CreatedAt: createdAt,
		},
	}
}

// OrderCreatedEvent is published when an order is created
type OrderCreatedEvent struct {
	Version   string              `json:"version"`
	EventType string              `json:"event_type"`
	Timestamp time.Time           `json:"timestamp"`
	TraceID   string              `json:"trace_id"`
	Payload   OrderCreatedPayload `json:"payload"`
}

// OrderCreatedPayload contains order data
type OrderCreatedPayload struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Total     float64   `json:"total"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// NewOrderCreatedEvent creates a new OrderCreatedEvent
func NewOrderCreatedEvent(id, userID uint, total float64, status string, createdAt time.Time, traceID string) *OrderCreatedEvent {
	return &OrderCreatedEvent{
		Version:   "1.0",
		EventType: "order.created",
		Timestamp: time.Now(),
		TraceID:   traceID,
		Payload: OrderCreatedPayload{
			ID:        id,
			UserID:    userID,
			Total:     total,
			Status:    status,
			CreatedAt: createdAt,
		},
	}
}

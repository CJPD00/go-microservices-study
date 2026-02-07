package domain

import (
	"time"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order represents the order domain entity
type Order struct {
	ID        uint
	UserID    uint
	Total     float64
	Status    OrderStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate validates the order entity
func (o *Order) Validate() error {
	if o.UserID == 0 {
		return ErrUserIDRequired
	}
	if o.Total <= 0 {
		return ErrInvalidTotal
	}
	if o.Total > 1000000 {
		return ErrTotalTooHigh
	}
	return nil
}

// NewOrder creates a new order with validation
func NewOrder(userID uint, total float64) (*Order, error) {
	order := &Order{
		UserID:    userID,
		Total:     total,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := order.Validate(); err != nil {
		return nil, err
	}

	return order, nil
}

// Confirm confirms the order
func (o *Order) Confirm() {
	o.Status = OrderStatusConfirmed
	o.UpdatedAt = time.Now()
}

// Cancel cancels the order
func (o *Order) Cancel() {
	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()
}

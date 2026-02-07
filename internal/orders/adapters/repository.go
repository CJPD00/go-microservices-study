package adapters

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"go-micro/internal/orders/domain"
	apperrors "go-micro/pkg/errors"
)

// OrderModel is the GORM model for orders (persistence layer)
type OrderModel struct {
	ID        uint               `gorm:"primaryKey"`
	UserID    uint               `gorm:"index;not null"`
	Total     float64            `gorm:"not null"`
	Status    domain.OrderStatus `gorm:"size:20;not null;default:'pending'"`
	CreatedAt time.Time          `gorm:"autoCreateTime"`
	UpdatedAt time.Time          `gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (OrderModel) TableName() string {
	return "orders"
}

// PostgresOrderRepository implements OrderRepository using PostgreSQL
type PostgresOrderRepository struct {
	db *gorm.DB
}

// NewPostgresOrderRepository creates a new PostgreSQL order repository
func NewPostgresOrderRepository(db *gorm.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

// Migrate runs auto-migration for the order model
func (r *PostgresOrderRepository) Migrate() error {
	return r.db.AutoMigrate(&OrderModel{})
}

// Create creates a new order
func (r *PostgresOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	model := toModel(order)

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return result.Error
	}

	// Update domain entity with generated ID
	order.ID = model.ID
	order.CreatedAt = model.CreatedAt
	order.UpdatedAt = model.UpdatedAt

	return nil
}

// GetByID retrieves an order by ID
func (r *PostgresOrderRepository) GetByID(ctx context.Context, id uint) (*domain.Order, error) {
	var model OrderModel

	result := r.db.WithContext(ctx).First(&model, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.NewOrderNotFound(id)
		}
		return nil, apperrors.NewInternal("failed to get order", result.Error)
	}

	return toDomain(&model), nil
}

// Update updates an existing order
func (r *PostgresOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	model := toModel(order)

	result := r.db.WithContext(ctx).Save(model)
	if result.Error != nil {
		return apperrors.NewInternal("failed to update order", result.Error)
	}

	order.UpdatedAt = model.UpdatedAt
	return nil
}

// Delete deletes an order by ID
func (r *PostgresOrderRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&OrderModel{}, id)
	if result.Error != nil {
		return apperrors.NewInternal("failed to delete order", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.NewOrderNotFound(id)
	}
	return nil
}

// GetByUserID retrieves orders for a user
func (r *PostgresOrderRepository) GetByUserID(ctx context.Context, userID uint) ([]*domain.Order, error) {
	var models []OrderModel

	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models)
	if result.Error != nil {
		return nil, apperrors.NewInternal("failed to get orders by user", result.Error)
	}

	orders := make([]*domain.Order, len(models))
	for i, model := range models {
		orders[i] = toDomain(&model)
	}

	return orders, nil
}

// toModel converts a domain entity to a GORM model
func toModel(order *domain.Order) *OrderModel {
	return &OrderModel{
		ID:        order.ID,
		UserID:    order.UserID,
		Total:     order.Total,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}
}

// toDomain converts a GORM model to a domain entity
func toDomain(model *OrderModel) *domain.Order {
	return &domain.Order{
		ID:        model.ID,
		UserID:    model.UserID,
		Total:     model.Total,
		Status:    model.Status,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

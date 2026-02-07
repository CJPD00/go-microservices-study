package application

import (
	"context"

	"go-micro/internal/orders/domain"
	"go-micro/internal/orders/ports"
	"go-micro/pkg/errors"
	"go-micro/pkg/logger"

	"go.uber.org/zap"
)

// OrderUseCase handles order business logic
type OrderUseCase struct {
	repo       ports.OrderRepository
	publisher  ports.EventPublisher
	userClient ports.UserClient
	log        *logger.Logger
}

// NewOrderUseCase creates a new order use case
func NewOrderUseCase(
	repo ports.OrderRepository,
	publisher ports.EventPublisher,
	userClient ports.UserClient,
	log *logger.Logger,
) *OrderUseCase {
	return &OrderUseCase{
		repo:       repo,
		publisher:  publisher,
		userClient: userClient,
		log:        log,
	}
}

// CreateOrderInput represents the input for creating an order
type CreateOrderInput struct {
	UserID uint
	Total  float64
}

// CreateOrderOutput represents the output of creating an order
type CreateOrderOutput struct {
	Order *domain.Order
}

// CreateOrder creates a new order
func (uc *OrderUseCase) CreateOrder(ctx context.Context, input CreateOrderInput) (*CreateOrderOutput, error) {
	// Validate user exists via gRPC
	if uc.userClient != nil {
		_, err := uc.userClient.GetUser(ctx, input.UserID)
		if err != nil {
			if errors.Is(err, errors.CodeNotFound) {
				return nil, domain.NewUserNotFoundError(input.UserID)
			}
			return nil, errors.Wrap(err, "failed to validate user")
		}
	}

	// Create domain entity with validation
	order, err := domain.NewOrder(input.UserID, input.Total)
	if err != nil {
		return nil, err
	}

	// Create order in repository
	if err := uc.repo.Create(ctx, order); err != nil {
		return nil, errors.NewInternal("failed to create order", err)
	}

	// Publish event (async, don't fail on error)
	if uc.publisher != nil {
		if err := uc.publisher.PublishOrderCreated(ctx, order); err != nil {
			uc.log.WithContext(ctx).Error("failed to publish order created event",
				zap.Error(err),
				zap.Uint("order_id", order.ID),
			)
		}
	}

	uc.log.WithContext(ctx).Info("order created",
		zap.Uint("order_id", order.ID),
		zap.Uint("user_id", order.UserID),
		zap.Float64("total", order.Total),
	)

	return &CreateOrderOutput{Order: order}, nil
}

// GetOrderInput represents the input for getting an order
type GetOrderInput struct {
	ID uint
}

// GetOrderOutput represents the output of getting an order
type GetOrderOutput struct {
	Order *domain.Order
}

// GetOrder retrieves an order by ID
func (uc *OrderUseCase) GetOrder(ctx context.Context, input GetOrderInput) (*GetOrderOutput, error) {
	order, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	return &GetOrderOutput{Order: order}, nil
}

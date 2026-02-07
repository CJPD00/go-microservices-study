package adapters

import (
	"context"

	"go-micro/internal/orders/domain"
	"go-micro/pkg/events"
	"go-micro/pkg/logger"
	"go-micro/pkg/rabbitmq"
)

// RabbitMQPublisher implements EventPublisher using RabbitMQ
type RabbitMQPublisher struct {
	publisher *rabbitmq.Publisher
	log       *logger.Logger
}

// NewRabbitMQPublisher creates a new RabbitMQ event publisher
func NewRabbitMQPublisher(publisher *rabbitmq.Publisher, log *logger.Logger) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		publisher: publisher,
		log:       log,
	}
}

// PublishOrderCreated publishes an order created event
func (p *RabbitMQPublisher) PublishOrderCreated(ctx context.Context, order *domain.Order) error {
	traceID := logger.GetTraceID(ctx)

	event := events.NewOrderCreatedEvent(
		order.ID,
		order.UserID,
		order.Total,
		string(order.Status),
		order.CreatedAt,
		traceID,
	)

	return p.publisher.Publish(ctx, events.RoutingKeyOrderCreated, event)
}

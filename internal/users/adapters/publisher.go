package adapters

import (
	"context"

	"go-micro/internal/users/domain"
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

// PublishUserCreated publishes a user created event
func (p *RabbitMQPublisher) PublishUserCreated(ctx context.Context, user *domain.User) error {
	traceID := logger.GetTraceID(ctx)

	event := events.NewUserCreatedEvent(
		user.ID,
		user.Name,
		user.Email,
		user.CreatedAt,
		traceID,
	)

	return p.publisher.Publish(ctx, events.RoutingKeyUserCreated, event)
}

package adapters

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"go-micro/pkg/events"
	"go-micro/pkg/logger"
	"go-micro/pkg/rabbitmq"
)

// UserCreatedConsumer consumes UserCreated events
type UserCreatedConsumer struct {
	consumer *rabbitmq.Consumer
	log      *logger.Logger
}

// NewUserCreatedConsumer creates a new consumer for UserCreated events
func NewUserCreatedConsumer(conn *rabbitmq.Connection, log *logger.Logger) (*UserCreatedConsumer, error) {
	consumer, err := rabbitmq.NewConsumer(
		conn,
		"orders.user-created", // queue name
		events.ExchangeUsers,  // exchange
		[]string{events.RoutingKeyUserCreated},
		log,
	)
	if err != nil {
		return nil, err
	}

	return &UserCreatedConsumer{
		consumer: consumer,
		log:      log,
	}, nil
}

// Start starts consuming UserCreated events
func (c *UserCreatedConsumer) Start(ctx context.Context) error {
	return c.consumer.Consume(ctx, c.handleMessage)
}

func (c *UserCreatedConsumer) handleMessage(ctx context.Context, body []byte) error {
	var event events.UserCreatedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.log.WithContext(ctx).Error("failed to unmarshal UserCreatedEvent",
			zap.Error(err),
		)
		return err
	}

	// Demo: just log the event
	c.log.WithContext(ctx).Info("received UserCreated event",
		zap.Uint("user_id", event.Payload.ID),
		zap.String("user_name", event.Payload.Name),
		zap.String("user_email", event.Payload.Email),
		zap.String("trace_id", event.TraceID),
	)

	// In a real application, you might:
	// - Cache the user info
	// - Update a local read model
	// - Trigger some business logic

	return nil
}

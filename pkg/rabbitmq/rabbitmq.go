package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"go-micro/pkg/logger"
)

// Connection manages a RabbitMQ connection with reconnect capability
type Connection struct {
	url        string
	conn       *amqp.Connection
	channel    *amqp.Channel
	log        *logger.Logger
	mu         sync.RWMutex
	closeChan  chan struct{}
	reconnects int
}

// NewConnection creates a new RabbitMQ connection
func NewConnection(url string, log *logger.Logger) (*Connection, error) {
	c := &Connection{
		url:       url,
		log:       log,
		closeChan: make(chan struct{}),
	}

	if err := c.connect(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Connection) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	c.conn = conn
	c.channel = ch

	c.log.Info("connected to RabbitMQ")
	return nil
}

// Channel returns the current channel
func (c *Connection) Channel() *amqp.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.channel
}

// Close closes the connection
func (c *Connection) Close() error {
	close(c.closeChan)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Publisher publishes messages to RabbitMQ
type Publisher struct {
	conn     *Connection
	exchange string
	log      *logger.Logger
}

// NewPublisher creates a new publisher
func NewPublisher(conn *Connection, exchange string, log *logger.Logger) (*Publisher, error) {
	// Declare exchange
	err := conn.Channel().ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &Publisher{
		conn:     conn,
		exchange: exchange,
		log:      log,
	}, nil
}

// Publish publishes a message
func (p *Publisher) Publish(ctx context.Context, routingKey string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	traceID := logger.GetTraceID(ctx)

	err = p.conn.Channel().PublishWithContext(
		ctx,
		p.exchange, // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          body,
			DeliveryMode:  amqp.Persistent,
			Timestamp:     time.Now(),
			CorrelationId: traceID,
			Headers: amqp.Table{
				"x-trace-id": traceID,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.log.WithContext(ctx).Debug("message published",
		zap.String("exchange", p.exchange),
		zap.String("routing_key", routingKey),
		zap.String("trace_id", traceID),
	)

	return nil
}

// Consumer consumes messages from RabbitMQ
type Consumer struct {
	conn        *Connection
	queue       string
	exchange    string
	routingKeys []string
	log         *logger.Logger
}

// NewConsumer creates a new consumer
func NewConsumer(conn *Connection, queue, exchange string, routingKeys []string, log *logger.Logger) (*Consumer, error) {
	ch := conn.Channel()

	// Declare queue
	_, err := ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange": exchange + ".dlx",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange for each routing key
	for _, key := range routingKeys {
		err = ch.QueueBind(queue, key, exchange, false, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to bind queue: %w", err)
		}
	}

	return &Consumer{
		conn:        conn,
		queue:       queue,
		exchange:    exchange,
		routingKeys: routingKeys,
		log:         log,
	}, nil
}

// MessageHandler is a function that handles a message
type MessageHandler func(ctx context.Context, body []byte) error

// Consume starts consuming messages
func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	msgs, err := c.conn.Channel().Consume(
		c.queue, // queue
		"",      // consumer
		false,   // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				// Extract trace ID from headers
				traceID := ""
				if tid, ok := msg.Headers["x-trace-id"].(string); ok {
					traceID = tid
				}
				msgCtx := logger.WithTraceIDContext(ctx, traceID)

				c.log.WithContext(msgCtx).Debug("message received",
					zap.String("queue", c.queue),
					zap.String("routing_key", msg.RoutingKey),
					zap.String("trace_id", traceID),
				)

				if err := handler(msgCtx, msg.Body); err != nil {
					c.log.WithContext(msgCtx).Error("failed to handle message",
						zap.Error(err),
						zap.String("queue", c.queue),
					)
					// Retry with delay (basic retry)
					time.Sleep(time.Second)
					msg.Nack(false, true)
				} else {
					msg.Ack(false)
				}
			}
		}
	}()

	c.log.Info("consumer started",
		zap.String("queue", c.queue),
		zap.Strings("routing_keys", c.routingKeys),
	)

	return nil
}

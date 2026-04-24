package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/andev0x/event-driven-order-system/pkg/events"
	"github.com/andev0x/event-driven-order-system/pkg/observability"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	exchangeName = "orders"
	exchangeType = "topic"
	queueName    = "analytics.orders"
	routingKey   = "order.created"

	deadLetterExchangeName = "orders.dlx"
	deadLetterRoutingKey   = "analytics.orders.failed"
	deadLetterQueueName    = "analytics.orders.dlq"
)

// RabbitMQConsumer implements event consumption from RabbitMQ.
type RabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tracer  trace.Tracer
}

// NewRabbitMQConsumer creates a new RabbitMQ consumer.
func NewRabbitMQConsumer(url string) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			slog.Error("Failed to close RabbitMQ connection", "error", closeErr)
		}
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		exchangeName,
		exchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		closeResources(channel, conn)
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare dead-letter exchange
	err = channel.ExchangeDeclare(
		deadLetterExchangeName,
		exchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		closeResources(channel, conn)
		return nil, fmt.Errorf("failed to declare dead-letter exchange: %w", err)
	}

	// Declare dead-letter queue
	dlq, err := channel.QueueDeclare(
		deadLetterQueueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		closeResources(channel, conn)
		return nil, fmt.Errorf("failed to declare dead-letter queue: %w", err)
	}

	// Bind dead-letter queue to dead-letter exchange
	err = channel.QueueBind(
		dlq.Name,
		deadLetterRoutingKey,
		deadLetterExchangeName,
		false,
		nil,
	)
	if err != nil {
		closeResources(channel, conn)
		return nil, fmt.Errorf("failed to bind dead-letter queue: %w", err)
	}

	// Declare queue
	queue, err := channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    deadLetterExchangeName,
			"x-dead-letter-routing-key": deadLetterRoutingKey,
		},
	)
	if err != nil {
		closeResources(channel, conn)
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		queue.Name,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		closeResources(channel, conn)
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	slog.Info("RabbitMQ consumer connected",
		"queue", queueName,
		"exchange", exchangeName,
		"dead_letter_queue", deadLetterQueueName,
	)

	return &RabbitMQConsumer{
		conn:    conn,
		channel: channel,
		tracer:  otel.Tracer("analytics-service/rabbitmq"),
	}, nil
}

// EventHandler is a function that processes an order created event.
type EventHandler func(context.Context, *events.OrderCreated) error

// StartConsuming starts consuming messages from the queue.
func (c *RabbitMQConsumer) StartConsuming(ctx context.Context, handler EventHandler) error {
	// Set QoS
	err := c.channel.Qos(1, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := c.channel.Consume(
		queueName,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	slog.Info("Analytics service started consuming order events", "queue", queueName)

	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("Stopping RabbitMQ consumer", "queue", queueName)
				return
			case msg, ok := <-msgs:
				if !ok {
					slog.Warn("RabbitMQ message channel closed", "queue", queueName)
					return
				}

				carrier := observability.NewAMQPCarrier(msg.Headers)
				messageCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)
				messageCtx, span := c.tracer.Start(messageCtx, "rabbitmq.consume order.created",
					trace.WithSpanKind(trace.SpanKindConsumer),
					trace.WithAttributes(
						attribute.String("messaging.system", "rabbitmq"),
						attribute.String("messaging.destination.name", queueName),
						attribute.String("messaging.destination.kind", "queue"),
						attribute.String("messaging.rabbitmq.routing_key", routingKey),
						attribute.String("messaging.operation", "process"),
					),
				)

				var event events.OrderCreated
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					slog.Error("Failed to unmarshal order event", "queue", queueName, "error", err)
					span.RecordError(err)
					span.End()
					if nackErr := msg.Nack(false, false); nackErr != nil {
						slog.Error("Failed to nack RabbitMQ message", "queue", queueName, "error", nackErr)
					}
					continue
				}
				span.SetAttributes(attribute.String("order.id", event.OrderID))

				slog.Info("Received OrderCreated event",
					"order_id", event.OrderID,
					"customer_id", event.CustomerID,
					"total_amount", event.TotalAmount,
				)

				if err := handler(messageCtx, &event); err != nil {
					slog.Error("Failed to process order event",
						"order_id", event.OrderID,
						"queue", queueName,
						"error", err,
					)
					span.RecordError(err)
					span.End()
					if nackErr := msg.Nack(false, false); nackErr != nil {
						slog.Error("Failed to nack RabbitMQ message", "queue", queueName, "error", nackErr)
					}
					slog.Warn("Message moved to dead-letter queue",
						"dead_letter_queue", deadLetterQueueName,
						"order_id", event.OrderID,
					)
					continue
				}

				if ackErr := msg.Ack(false); ackErr != nil {
					slog.Error("Failed to ack RabbitMQ message", "queue", queueName, "error", ackErr)
					span.RecordError(ackErr)
					span.End()
					continue
				}
				slog.Info("Successfully processed order event", "order_id", event.OrderID)
				span.End()
			}
		}
	}()

	return nil
}

// Close closes the RabbitMQ connection.
func (c *RabbitMQConsumer) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return err
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

// HealthCheck checks if the RabbitMQ connection is alive.
func (c *RabbitMQConsumer) HealthCheck() error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	if c.conn.IsClosed() {
		return fmt.Errorf("connection is closed")
	}
	if c.channel == nil {
		return fmt.Errorf("channel is nil")
	}
	return nil
}

func closeResources(channel *amqp.Channel, conn *amqp.Connection) {
	if channel != nil {
		if err := channel.Close(); err != nil {
			slog.Error("Failed to close RabbitMQ channel", "error", err)
		}
	}
	if conn != nil {
		if err := conn.Close(); err != nil {
			slog.Error("Failed to close RabbitMQ connection", "error", err)
		}
	}
}

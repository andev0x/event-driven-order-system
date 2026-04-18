package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher provides methods for publishing messages to RabbitMQ.
type Publisher struct {
	conn *Connection
}

// NewPublisher creates a new Publisher using an existing connection.
func NewPublisher(conn *Connection) *Publisher {
	return &Publisher{conn: conn}
}

// Publish sends a message to the exchange with the given routing key.
func (p *Publisher) Publish(ctx context.Context, routingKey string, body []byte) error {
	exchangeName := p.conn.config.ExchangeName
	if exchangeName == "" {
		exchangeName = DefaultExchangeName
	}

	err := p.conn.channel.PublishWithContext(
		ctx,
		exchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// PublishJSON marshals the payload to JSON and publishes it.
func (p *Publisher) PublishJSON(ctx context.Context, routingKey string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	return p.Publish(ctx, routingKey, body)
}

// Close closes the underlying connection.
func (p *Publisher) Close() error {
	return p.conn.Close()
}

// HealthCheck verifies the publisher connection is alive.
func (p *Publisher) HealthCheck() error {
	return p.conn.HealthCheck()
}

// Consumer provides methods for consuming messages from RabbitMQ.
type Consumer struct {
	conn      *Connection
	queueName string
}

// NewConsumer creates a new Consumer for the specified queue.
func NewConsumer(conn *Connection, queueName, routingKey string) (*Consumer, error) {
	if err := conn.DeclareQueue(queueName, routingKey); err != nil {
		return nil, err
	}

	// Set QoS
	err := conn.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &Consumer{
		conn:      conn,
		queueName: queueName,
	}, nil
}

// MessageHandler is a function that processes a message.
type MessageHandler func(body []byte) error

// StartConsuming begins consuming messages and calls the handler for each.
func (c *Consumer) StartConsuming(ctx context.Context, handler MessageHandler) error {
	msgs, err := c.conn.channel.Consume(
		c.queueName,
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

	slog.Info("RabbitMQ consumer started", "queue", c.queueName)

	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("Stopping RabbitMQ consumer", "queue", c.queueName)
				return
			case msg, ok := <-msgs:
				if !ok {
					slog.Warn("RabbitMQ message channel closed", "queue", c.queueName)
					return
				}

				if err := handler(msg.Body); err != nil {
					slog.Error("Failed to process RabbitMQ message", "queue", c.queueName, "error", err)
					// Requeue on error
					if nackErr := msg.Nack(false, true); nackErr != nil {
						slog.Error("Failed to nack RabbitMQ message", "queue", c.queueName, "error", nackErr)
					}
					continue
				}

				if ackErr := msg.Ack(false); ackErr != nil {
					slog.Error("Failed to ack RabbitMQ message", "queue", c.queueName, "error", ackErr)
				}
			}
		}
	}()

	return nil
}

// Close closes the underlying connection.
func (c *Consumer) Close() error {
	return c.conn.Close()
}

// HealthCheck verifies the consumer connection is alive.
func (c *Consumer) HealthCheck() error {
	return c.conn.HealthCheck()
}

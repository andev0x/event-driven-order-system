package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/andev0x/event-driven-order-system/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchangeName = "orders"
	exchangeType = "topic"
	queueName    = "notifications.orders"
	routingKey   = "order.created"
)

// RabbitMQConsumer implements event consumption from RabbitMQ.
type RabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQConsumer creates a new RabbitMQ consumer with retry logic.
func NewRabbitMQConsumer(url string, maxRetries int) (*RabbitMQConsumer, error) {
	var conn *amqp.Connection
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %v", i+1, maxRetries, err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
	}

	channel, err := conn.Channel()
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Error closing connection: %v", closeErr)
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

	// Declare queue
	queue, err := channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
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

	log.Printf("RabbitMQ consumer connected, queue '%s' bound to exchange '%s'", queueName, exchangeName)

	return &RabbitMQConsumer{
		conn:    conn,
		channel: channel,
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

	log.Println("Notification worker is now consuming order events...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping consumer...")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("Message channel closed")
					return
				}

				var event events.OrderCreated
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					log.Printf("Error unmarshaling event: %v", err)
					if nackErr := msg.Nack(false, false); nackErr != nil {
						log.Printf("Error nacking message: %v", nackErr)
					}
					continue
				}

				log.Printf("Received OrderCreated event: OrderID=%s, CustomerID=%s",
					event.OrderID, event.CustomerID)

				if err := handler(ctx, &event); err != nil {
					log.Printf("Error processing event: %v", err)
					if nackErr := msg.Nack(false, true); nackErr != nil {
						log.Printf("Error nacking message: %v", nackErr)
					}
					continue
				}

				if ackErr := msg.Ack(false); ackErr != nil {
					log.Printf("Error acking message: %v", ackErr)
				}
				log.Printf("Successfully sent notification for order: %s", event.OrderID)
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
			log.Printf("Error closing channel: %v", err)
		}
	}
	if conn != nil {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}
}

package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	pkgevents "github.com/andev0x/event-driven-order-system/pkg/events"
	"github.com/andev0x/order-service/internal/order"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchangeName = "orders"
	exchangeType = "topic"
	routingKey   = "order.created"
)

// RabbitMQPublisher implements order.EventPublisher using RabbitMQ.
type RabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQPublisher creates a new RabbitMQ publisher.
func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
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
		if closeErr := channel.Close(); closeErr != nil {
			log.Printf("Error closing channel: %v", closeErr)
		}
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Error closing connection: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Printf("RabbitMQ publisher connected and exchange '%s' declared", exchangeName)

	return &RabbitMQPublisher{
		conn:    conn,
		channel: channel,
	}, nil
}

// PublishOrderCreated publishes an order created event.
func (p *RabbitMQPublisher) PublishOrderCreated(ctx context.Context, o *order.Order) error {
	event := pkgevents.NewOrderCreated(
		o.ID,
		o.CustomerID,
		o.ProductID,
		o.Quantity,
		o.TotalAmount,
		o.Status,
		o.CreatedAt,
	)

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.channel.PublishWithContext(
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
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Published OrderCreated event for order: %s", o.ID)
	return nil
}

// Close closes the RabbitMQ connection.
func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			return err
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

// HealthCheck checks if the RabbitMQ connection is alive.
func (p *RabbitMQPublisher) HealthCheck() error {
	if p.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	if p.conn.IsClosed() {
		return fmt.Errorf("connection is closed")
	}
	if p.channel == nil {
		return fmt.Errorf("channel is nil")
	}
	return nil
}

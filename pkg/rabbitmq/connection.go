package rabbitmq

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Default exchange and routing configuration.
const (
	DefaultExchangeName = "orders"
	DefaultExchangeType = "topic"
	RoutingKeyCreated   = "order.created"
	RoutingKeyConfirmed = "order.confirmed"
	RoutingKeyCancelled = "order.cancelled"
)

// Config holds RabbitMQ connection configuration.
type Config struct {
	URL          string
	ExchangeName string
	ExchangeType string
	MaxRetries   int
	RetryDelay   time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		URL:          "amqp://guest:guest@localhost:5672/",
		ExchangeName: DefaultExchangeName,
		ExchangeType: DefaultExchangeType,
		MaxRetries:   10,
		RetryDelay:   5 * time.Second,
	}
}

// Connection wraps an AMQP connection with its channel.
type Connection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  Config
}

// Connect establishes a connection to RabbitMQ with retry logic.
func Connect(cfg Config) (*Connection, error) {
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 10
	}
	retryDelay := cfg.RetryDelay
	if retryDelay <= 0 {
		retryDelay = 5 * time.Second
	}

	var conn *amqp.Connection
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(cfg.URL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
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
	exchangeName := cfg.ExchangeName
	if exchangeName == "" {
		exchangeName = DefaultExchangeName
	}
	exchangeType := cfg.ExchangeType
	if exchangeType == "" {
		exchangeType = DefaultExchangeType
	}

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

	log.Printf("RabbitMQ connected, exchange '%s' declared", exchangeName)

	return &Connection{
		conn:    conn,
		channel: channel,
		config:  cfg,
	}, nil
}

// Channel returns the underlying AMQP channel.
func (c *Connection) Channel() *amqp.Channel {
	return c.channel
}

// Close closes the RabbitMQ connection and channel.
func (c *Connection) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}
	return nil
}

// HealthCheck verifies the RabbitMQ connection is alive.
func (c *Connection) HealthCheck() error {
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

// DeclareQueue declares a queue and binds it to the exchange.
func (c *Connection) DeclareQueue(queueName, routingKey string) error {
	exchangeName := c.config.ExchangeName
	if exchangeName == "" {
		exchangeName = DefaultExchangeName
	}

	queue, err := c.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = c.channel.QueueBind(
		queue.Name,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Printf("Queue '%s' declared and bound to exchange '%s' with routing key '%s'",
		queueName, exchangeName, routingKey)

	return nil
}

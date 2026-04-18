package notification

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/andev0x/event-driven-order-system/pkg/events"
)

// Service handles notification business logic.
type Service struct {
	sender Sender
}

// NewService creates a new notification service.
func NewService(sender Sender) *Service {
	return &Service{sender: sender}
}

// ProcessOrderCreated processes an order created event and sends a notification.
func (s *Service) ProcessOrderCreated(ctx context.Context, event *events.OrderCreated) error {
	subject := fmt.Sprintf("Order Confirmation - %s", event.OrderID)
	body := fmt.Sprintf(
		"Thank you for your order!\n\n"+
			"Order ID: %s\n"+
			"Product: %s\n"+
			"Quantity: %d\n"+
			"Total: $%.2f\n"+
			"Status: %s\n\n"+
			"We will notify you when your order ships.",
		event.OrderID,
		event.ProductID,
		event.Quantity,
		event.TotalAmount,
		event.Status,
	)

	if err := s.sender.Send(ctx, event.CustomerID, subject, body); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	slog.Info("Notification sent",
		"order_id", event.OrderID,
		"customer_id", event.CustomerID,
	)
	return nil
}

// ConsoleSender implements Sender by logging to console (for demo/development).
type ConsoleSender struct {
	simulatedDelay time.Duration
}

// NewConsoleSender creates a new console sender.
func NewConsoleSender() *ConsoleSender {
	return &ConsoleSender{
		simulatedDelay: 500 * time.Millisecond,
	}
}

// NewConsoleSenderWithDelay creates a console sender with custom delay.
func NewConsoleSenderWithDelay(delay time.Duration) *ConsoleSender {
	return &ConsoleSender{
		simulatedDelay: delay,
	}
}

// Send logs the notification to console.
func (s *ConsoleSender) Send(_ context.Context, recipient, subject, _ string) error {
	// Simulate notification delay
	time.Sleep(s.simulatedDelay)

	slog.Info("Console notification dispatched",
		"recipient", recipient,
		"subject", subject,
	)
	return nil
}

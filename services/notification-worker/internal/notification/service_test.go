package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andev0x/event-driven-order-system/pkg/events"
)

// Mock sender for testing
type mockSender struct {
	sendFunc func(ctx context.Context, recipient, subject, body string) error
	calls    []sendCall
}

type sendCall struct {
	recipient string
	subject   string
	body      string
}

func (m *mockSender) Send(ctx context.Context, recipient, subject, body string) error {
	m.calls = append(m.calls, sendCall{recipient, subject, body})
	if m.sendFunc != nil {
		return m.sendFunc(ctx, recipient, subject, body)
	}
	return nil
}

func TestService_ProcessOrderCreated(t *testing.T) {
	t.Run("successfully sends notification", func(t *testing.T) {
		sender := &mockSender{}
		svc := NewService(sender)

		event := &events.OrderCreated{
			OrderID:     "order-123",
			CustomerID:  "cust-456",
			ProductID:   "prod-789",
			Quantity:    2,
			TotalAmount: 99.99,
			Status:      events.OrderStatusPending,
			CreatedAt:   time.Now(),
		}

		err := svc.ProcessOrderCreated(context.Background(), event)
		if err != nil {
			t.Errorf("ProcessOrderCreated() error = %v", err)
		}

		if len(sender.calls) != 1 {
			t.Errorf("Expected 1 call, got %d", len(sender.calls))
		}

		if sender.calls[0].recipient != "cust-456" {
			t.Errorf("recipient = %v, want cust-456", sender.calls[0].recipient)
		}
	})

	t.Run("returns error when sender fails", func(t *testing.T) {
		sender := &mockSender{
			sendFunc: func(_ context.Context, _, _, _ string) error {
				return errors.New("send failed")
			},
		}
		svc := NewService(sender)

		event := &events.OrderCreated{
			OrderID:    "order-123",
			CustomerID: "cust-456",
		}

		err := svc.ProcessOrderCreated(context.Background(), event)
		if err == nil {
			t.Error("ProcessOrderCreated() expected error")
		}
	})
}

func TestConsoleSender(t *testing.T) {
	t.Run("creates sender with default delay", func(t *testing.T) {
		sender := NewConsoleSender()
		if sender.simulatedDelay != 500*time.Millisecond {
			t.Errorf("simulatedDelay = %v, want 500ms", sender.simulatedDelay)
		}
	})

	t.Run("creates sender with custom delay", func(t *testing.T) {
		sender := NewConsoleSenderWithDelay(100 * time.Millisecond)
		if sender.simulatedDelay != 100*time.Millisecond {
			t.Errorf("simulatedDelay = %v, want 100ms", sender.simulatedDelay)
		}
	})
}

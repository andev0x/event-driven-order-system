package notification

import "context"

// Sender defines the interface for sending notifications.
type Sender interface {
	// Send sends a notification to a recipient.
	Send(ctx context.Context, recipient, subject, body string) error
}

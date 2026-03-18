// Package notification contains the core domain logic for notification processing.
//
// This package implements the Notification domain which handles sending
// notifications to customers when order events occur.
//
// Example usage:
//
//	sender := notification.NewConsoleSender()
//	svc := notification.NewService(sender)
//	err := svc.ProcessOrderCreated(ctx, event)
package notification

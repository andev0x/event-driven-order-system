// Package rabbitmq provides shared RabbitMQ connection and messaging utilities.
//
// This package contains common RabbitMQ setup including connection handling,
// exchange/queue declaration, and message publishing/consuming patterns.
//
// Example usage for publishing:
//
//	pub, err := rabbitmq.NewPublisher(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer pub.Close()
//
//	err = pub.Publish(ctx, "order.created", eventBytes)
package rabbitmq

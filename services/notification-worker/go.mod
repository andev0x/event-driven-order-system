module github.com/andev0x/notification-worker

go 1.21

require (
	github.com/andev0x/event-driven-order-system/pkg v0.0.0
	github.com/rabbitmq/amqp091-go v1.10.0
)

replace github.com/andev0x/event-driven-order-system/pkg => ../../pkg

package observability

import amqp "github.com/rabbitmq/amqp091-go"

// AMQPCarrier adapts RabbitMQ headers to OpenTelemetry text maps.
type AMQPCarrier struct {
	headers amqp.Table
}

// NewAMQPCarrier creates a carrier from AMQP headers.
func NewAMQPCarrier(headers amqp.Table) AMQPCarrier {
	if headers == nil {
		headers = amqp.Table{}
	}

	return AMQPCarrier{headers: headers}
}

// Get returns a header value for the given key.
func (c AMQPCarrier) Get(key string) string {
	val, ok := c.headers[key]
	if !ok {
		return ""
	}

	str, ok := val.(string)
	if !ok {
		return ""
	}

	return str
}

// Set stores a key-value pair in message headers.
func (c AMQPCarrier) Set(key, value string) {
	c.headers[key] = value
}

// Keys returns all keys present in the carrier.
func (c AMQPCarrier) Keys() []string {
	keys := make([]string, 0, len(c.headers))
	for key := range c.headers {
		keys = append(keys, key)
	}

	return keys
}

// Headers returns underlying AMQP headers.
func (c AMQPCarrier) Headers() amqp.Table {
	return c.headers
}

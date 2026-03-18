package rabbitmq

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.URL != "amqp://guest:guest@localhost:5672/" {
		t.Errorf("URL = %v, want amqp://guest:guest@localhost:5672/", cfg.URL)
	}
	if cfg.ExchangeName != DefaultExchangeName {
		t.Errorf("ExchangeName = %v, want %v", cfg.ExchangeName, DefaultExchangeName)
	}
	if cfg.ExchangeType != DefaultExchangeType {
		t.Errorf("ExchangeType = %v, want %v", cfg.ExchangeType, DefaultExchangeType)
	}
	if cfg.MaxRetries != 10 {
		t.Errorf("MaxRetries = %v, want 10", cfg.MaxRetries)
	}
	if cfg.RetryDelay != 5*time.Second {
		t.Errorf("RetryDelay = %v, want 5s", cfg.RetryDelay)
	}
}

func TestConnection_HealthCheck_NilConnection(t *testing.T) {
	c := &Connection{
		conn:    nil,
		channel: nil,
	}

	err := c.HealthCheck()
	if err == nil {
		t.Error("HealthCheck() should return error for nil connection")
	}
}

func TestRoutingKeyConstants(t *testing.T) {
	if RoutingKeyCreated != "order.created" {
		t.Errorf("RoutingKeyCreated = %v, want order.created", RoutingKeyCreated)
	}
	if RoutingKeyConfirmed != "order.confirmed" {
		t.Errorf("RoutingKeyConfirmed = %v, want order.confirmed", RoutingKeyConfirmed)
	}
	if RoutingKeyCancelled != "order.cancelled" {
		t.Errorf("RoutingKeyCancelled = %v, want order.cancelled", RoutingKeyCancelled)
	}
}

package redis

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Host != "localhost" {
		t.Errorf("Host = %v, want localhost", cfg.Host)
	}
	if cfg.Port != "6379" {
		t.Errorf("Port = %v, want 6379", cfg.Port)
	}
	if cfg.Password != "" {
		t.Errorf("Password = %v, want empty", cfg.Password)
	}
	if cfg.DB != 0 {
		t.Errorf("DB = %v, want 0", cfg.DB)
	}
	if cfg.PoolSize != 10 {
		t.Errorf("PoolSize = %v, want 10", cfg.PoolSize)
	}
}

func TestHealthCheck_NilClient(t *testing.T) {
	err := HealthCheck(nil)
	if err == nil {
		t.Error("HealthCheck(nil) should return error")
	}
}

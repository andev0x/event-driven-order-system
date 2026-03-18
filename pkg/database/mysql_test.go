package database

import (
	"database/sql"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Host != "localhost" {
		t.Errorf("Host = %v, want localhost", cfg.Host)
	}
	if cfg.Port != "3306" {
		t.Errorf("Port = %v, want 3306", cfg.Port)
	}
	if cfg.MaxOpenConns != 25 {
		t.Errorf("MaxOpenConns = %v, want 25", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns = %v, want 5", cfg.MaxIdleConns)
	}
}

func TestHealthCheck_NilDB(t *testing.T) {
	err := HealthCheck(nil)
	if err == nil {
		t.Error("HealthCheck(nil) should return error")
	}
}

func TestHealthCheck_InvalidDB(t *testing.T) {
	// Create a DB with invalid DSN that will fail ping
	db, err := sql.Open("mysql", "invalid:invalid@tcp(localhost:9999)/nonexistent")
	if err != nil {
		t.Skipf("Could not create test DB: %v", err)
	}
	defer db.Close()

	// HealthCheck should fail because we can't actually connect
	err = HealthCheck(db)
	if err == nil {
		t.Error("HealthCheck should fail with invalid connection")
	}
}

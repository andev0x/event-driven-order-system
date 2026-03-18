package config

import (
	"os"
	"testing"
	"time"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		setValue     string
		defaultValue string
		want         string
	}{
		{
			name:         "returns env value when set",
			key:          "TEST_GET_ENV",
			setValue:     "custom_value",
			defaultValue: "default",
			want:         "custom_value",
		},
		{
			name:         "returns default when not set",
			key:          "TEST_GET_ENV_MISSING",
			setValue:     "",
			defaultValue: "default_value",
			want:         "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
				defer os.Unsetenv(tt.key)
			}

			got := GetEnv(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		setValue     string
		defaultValue int
		want         int
	}{
		{
			name:         "returns int value when set",
			key:          "TEST_GET_ENV_INT",
			setValue:     "42",
			defaultValue: 10,
			want:         42,
		},
		{
			name:         "returns default when not set",
			key:          "TEST_GET_ENV_INT_MISSING",
			setValue:     "",
			defaultValue: 100,
			want:         100,
		},
		{
			name:         "returns default when invalid int",
			key:          "TEST_GET_ENV_INT_INVALID",
			setValue:     "not_a_number",
			defaultValue: 50,
			want:         50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
				defer os.Unsetenv(tt.key)
			}

			got := GetEnvInt(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetEnvInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		setValue     string
		defaultValue bool
		want         bool
	}{
		{
			name:         "returns true when set to true",
			key:          "TEST_GET_ENV_BOOL_TRUE",
			setValue:     "true",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "returns false when set to false",
			key:          "TEST_GET_ENV_BOOL_FALSE",
			setValue:     "false",
			defaultValue: true,
			want:         false,
		},
		{
			name:         "returns default when not set",
			key:          "TEST_GET_ENV_BOOL_MISSING",
			setValue:     "",
			defaultValue: true,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
				defer os.Unsetenv(tt.key)
			}

			got := GetEnvBool(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetEnvBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		setValue     string
		defaultValue time.Duration
		want         time.Duration
	}{
		{
			name:         "returns duration when set",
			key:          "TEST_GET_ENV_DURATION",
			setValue:     "5s",
			defaultValue: time.Second,
			want:         5 * time.Second,
		},
		{
			name:         "returns default when not set",
			key:          "TEST_GET_ENV_DURATION_MISSING",
			setValue:     "",
			defaultValue: 10 * time.Minute,
			want:         10 * time.Minute,
		},
		{
			name:         "returns default when invalid",
			key:          "TEST_GET_ENV_DURATION_INVALID",
			setValue:     "invalid",
			defaultValue: 30 * time.Second,
			want:         30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
				defer os.Unsetenv(tt.key)
			}

			got := GetEnvDuration(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetEnvDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

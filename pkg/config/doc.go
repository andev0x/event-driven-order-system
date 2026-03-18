// Package config provides shared configuration utilities.
//
// This package contains helpers for loading configuration from environment
// variables with default values.
//
// Example usage:
//
//	dbHost := config.GetEnv("DB_HOST", "localhost")
//	port := config.GetEnvInt("PORT", 8080)
package config

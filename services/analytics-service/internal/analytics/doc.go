// Package analytics contains the core domain logic for analytics management.
//
// This package implements the Analytics domain following domain-driven design principles.
// It handles processing order events, storing metrics, and computing analytics summaries.
//
// Example usage:
//
//	svc := analytics.NewService(repo, cache)
//	summary, err := svc.GetSummary(ctx)
package analytics

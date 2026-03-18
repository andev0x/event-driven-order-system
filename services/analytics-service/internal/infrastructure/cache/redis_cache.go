package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/andev0x/analytics-service/internal/analytics"
	"github.com/redis/go-redis/v9"
)

const (
	summaryKey = "analytics:summary"
	summaryTTL = 5 * time.Minute
)

// RedisCache implements analytics.Cache using Redis.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis analytics cache.
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// GetSummary retrieves analytics summary from cache.
func (c *RedisCache) GetSummary(ctx context.Context) (*analytics.Summary, error) {
	data, err := c.client.Get(ctx, summaryKey).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("summary not found in cache")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get summary from cache: %w", err)
	}

	var summary analytics.Summary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal summary: %w", err)
	}

	return &summary, nil
}

// SetSummary stores analytics summary in cache.
func (c *RedisCache) SetSummary(ctx context.Context, summary *analytics.Summary) error {
	data, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	if err := c.client.Set(ctx, summaryKey, data, summaryTTL).Err(); err != nil {
		return fmt.Errorf("failed to set summary in cache: %w", err)
	}

	return nil
}

// InvalidateSummary removes analytics summary from cache.
func (c *RedisCache) InvalidateSummary(ctx context.Context) error {
	if err := c.client.Del(ctx, summaryKey).Err(); err != nil {
		return fmt.Errorf("failed to invalidate summary cache: %w", err)
	}
	return nil
}

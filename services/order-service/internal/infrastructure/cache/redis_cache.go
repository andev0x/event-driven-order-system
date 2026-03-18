package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/andev0x/order-service/internal/order"
	"github.com/redis/go-redis/v9"
)

const (
	orderKeyPrefix = "order:"
	orderTTL       = 15 * time.Minute
)

// RedisCache implements order.Cache using Redis.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis order cache.
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// Get retrieves an order from cache.
func (c *RedisCache) Get(ctx context.Context, id string) (*order.Order, error) {
	key := orderKeyPrefix + id
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("order not found in cache")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order from cache: %w", err)
	}

	var o order.Order
	if err := json.Unmarshal(data, &o); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &o, nil
}

// Set stores an order in cache.
func (c *RedisCache) Set(ctx context.Context, o *order.Order) error {
	key := orderKeyPrefix + o.ID
	data, err := json.Marshal(o)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	if err := c.client.Set(ctx, key, data, orderTTL).Err(); err != nil {
		return fmt.Errorf("failed to set order in cache: %w", err)
	}

	return nil
}

// Delete removes an order from cache.
func (c *RedisCache) Delete(ctx context.Context, id string) error {
	key := orderKeyPrefix + id
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete order from cache: %w", err)
	}
	return nil
}

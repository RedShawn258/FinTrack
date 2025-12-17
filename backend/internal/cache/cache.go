package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// Cache TTL for summary endpoints (5 minutes)
	DefaultTTL = 5 * time.Minute
)

type CacheService struct {
	client  *redis.Client
	logger  *zap.Logger
	enabled bool
}

// NewCacheService creates a new cache service instance
func NewCacheService(addr, password string, db int, enabled bool, logger *zap.Logger) (*CacheService, error) {
	if !enabled {
		logger.Info("Cache is disabled")
		return &CacheService{
			client:  nil,
			logger:  logger,
			enabled: false,
		}, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis cache connected successfully", zap.String("addr", addr), zap.Int("db", db))

	return &CacheService{
		client:  client,
		logger:  logger,
		enabled: true,
	}, nil
}

// Get retrieves a value from cache
func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	if !c.enabled || c.client == nil {
		return false, nil
	}

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		c.logger.Debug("Cache miss", zap.String("key", key))
		return false, nil
	}
	if err != nil {
		c.logger.Warn("Cache get error", zap.String("key", key), zap.Error(err))
		return false, err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		c.logger.Warn("Cache unmarshal error", zap.String("key", key), zap.Error(err))
		return false, err
	}

	c.logger.Debug("Cache hit", zap.String("key", key))
	return true, nil
}

// Set stores a value in cache with TTL
func (c *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !c.enabled || c.client == nil {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		c.logger.Warn("Cache set error", zap.String("key", key), zap.Error(err))
		return err
	}

	c.logger.Debug("Cache set", zap.String("key", key), zap.Duration("ttl", ttl))
	return nil
}

// Delete removes a key from cache
func (c *CacheService) Delete(ctx context.Context, key string) error {
	if !c.enabled || c.client == nil {
		return nil
	}

	if err := c.client.Del(ctx, key).Err(); err != nil {
		c.logger.Warn("Cache delete error", zap.String("key", key), zap.Error(err))
		return err
	}

	c.logger.Debug("Cache deleted", zap.String("key", key))
	return nil
}

// DeletePattern removes all keys matching a pattern
func (c *CacheService) DeletePattern(ctx context.Context, pattern string) error {
	if !c.enabled || c.client == nil {
		return nil
	}

	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	keys := []string{}

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		c.logger.Warn("Cache scan error", zap.String("pattern", pattern), zap.Error(err))
		return err
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			c.logger.Warn("Cache delete pattern error", zap.String("pattern", pattern), zap.Error(err))
			return err
		}
		c.logger.Debug("Cache deleted pattern", zap.String("pattern", pattern), zap.Int("count", len(keys)))
	}

	return nil
}

// Close closes the Redis connection
func (c *CacheService) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// IsEnabled returns whether caching is enabled
func (c *CacheService) IsEnabled() bool {
	return c.enabled
}

// GenerateKey creates a cache key for summary endpoints
func GenerateKey(prefix string, userID uint, params ...string) string {
	key := fmt.Sprintf("%s:%d", prefix, userID)
	for _, param := range params {
		if param != "" {
			key = fmt.Sprintf("%s:%s", key, param)
		}
	}
	return key
}





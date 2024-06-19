package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Get(key string) (string, error) {
	data, err := c.client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("cache miss for key %q", key)
	} else if err != nil {
		return "", fmt.Errorf("failed to get value for key %q: %v", key, err)
	}
	return data, nil
}

func (c *RedisCache) Set(key string, value string, duration time.Duration) error {
	if err := c.client.Set(context.Background(), key, value, duration).Err(); err != nil {
		return fmt.Errorf("failed to set value for key %q: %v", key, err)
	}
	return nil
}

func (c *RedisCache) Delete(key string) error {
	if err := c.client.Del(context.Background(), key).Err(); err != nil {
		return fmt.Errorf("failed to delete value for key %q: %v", key, err)
	}
	return nil
}

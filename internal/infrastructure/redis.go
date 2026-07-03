package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// dfn of my client
type RedisClient struct {
	client *redis.Client
}

// redis connetion handshaker function - kept a retry duration of 5secs
func NewRedisClient(url string) (*RedisClient, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("url couldnt be parsed : %w", err)
	}
	clnt := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := clnt.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("couldnt ping the redis server: %w", err)
	}

	return &RedisClient{client: clnt}, nil
}

func (r *RedisClient) Set(ctx context.Context, key string, value any, exp time.Duration) error {
	return r.client.Set(ctx, key, value, exp).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

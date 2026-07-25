package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type OrderCache struct {
	client *redis.Client
}

func NewOrderCache(client *redis.Client) *OrderCache {
	return &OrderCache{
		client: client,
	}
}

func (o *OrderCache) Set(ctx context.Context, key string, value interface{}) error {
	return o.client.Set(ctx, key, value, 0).Err()
}

func (o *OrderCache) Get(ctx context.Context, key string) (string, error) {
	return o.client.Get(ctx, key).Result()
}

func (o *OrderCache) Delete(ctx context.Context, key string) error {
	return o.client.Del(ctx, key).Err()
}

package outgoing

import (
	"context"
	"encoding/json"
	"fmt"
	"kitchen-service/src/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisAggregateRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisAggregateRepository(redisClient *redis.Client) *RedisAggregateRepository {
	return &RedisAggregateRepository{
		client: redisClient,
		ttl:    24 * time.Hour,
	}
}

func (r *RedisAggregateRepository) SaveAggregate(ctx context.Context, aggregate *domain.KitchenAggregate) error {
	key := r.getKey(aggregate.OrderID)

	data, err := json.Marshal(aggregate)
	if err != nil {
		return fmt.Errorf("failed to marshal aggregate: %w", err)
	}

	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to save aggregate to redis: %w", err)
	}

	return nil
}

func (r *RedisAggregateRepository) GetAggregate(ctx context.Context, orderID string) (*domain.KitchenAggregate, error) {
	key := r.getKey(orderID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get aggregate from redis: %w", err)
	}

	var aggregate domain.KitchenAggregate
	if err := json.Unmarshal([]byte(data), &aggregate); err != nil {
		return nil, fmt.Errorf("failed to unmarshal aggregate: %w", err)
	}

	return &aggregate, nil
}

func (r *RedisAggregateRepository) DeleteAggregate(ctx context.Context, orderID string) error {
	key := r.getKey(orderID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete aggregate from redis: %w", err)
	}

	return nil
}

func (r *RedisAggregateRepository) getKey(orderID string) string {
	return fmt.Sprintf("kitchen:%s", orderID)
}

package outgoing

import (
	"context"
	"delivery-service/src/application/ports"
	"delivery-service/src/domain"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRepository(redisClient *redis.Client) ports.AggregateRepository {
	return &RedisRepository{
		client: redisClient,
		ttl:    24 * time.Hour,
	}
}

func (r *RedisRepository) SaveAggregate(ctx context.Context, aggregate *domain.DeliveryAggregate) error {
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

func (r *RedisRepository) GetAggregate(ctx context.Context, orderID string) (*domain.DeliveryAggregate, error) {
	key := r.getKey(orderID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get aggregate from redis: %w", err)
	}

	var aggregate domain.DeliveryAggregate
	if err := json.Unmarshal([]byte(data), &aggregate); err != nil {
		return nil, fmt.Errorf("failed to unmarshal aggregate: %w", err)
	}

	return &aggregate, nil
}

func (r *RedisRepository) DeleteAggregate(ctx context.Context, orderID string) error {
	key := r.getKey(orderID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete aggregate from redis: %w", err)
	}

	return nil
}

func (r *RedisRepository) GetReadyToDeliverOrders(ctx context.Context) ([]*domain.DeliveryAggregate, error) {
	pattern := "delivery:*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to scan redis keys: %w", err)
	}

	var readyOrders []*domain.DeliveryAggregate

	for _, key := range keys {
		data, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var aggregate domain.DeliveryAggregate
		if err := json.Unmarshal([]byte(data), &aggregate); err != nil {
			continue
		}

		if aggregate.CanStartDelivery() {
			readyOrders = append(readyOrders, &aggregate)
		}
	}

	return readyOrders, nil
}

func (r *RedisRepository) getKey(orderID string) string {
	return fmt.Sprintf("delivery:%s", orderID)
}

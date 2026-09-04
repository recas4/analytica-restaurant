package ports

import (
	"context"
	"kitchen-service/src/domain"
)

type KitchenRepository interface {
	SaveOrder(ctx context.Context, order *domain.KitchenOrder) error
	UpdateOrder(ctx context.Context, order *domain.KitchenOrder) error
	GetOrderByID(ctx context.Context, orderID string) (*domain.KitchenOrder, error)
	GetOrdersByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.KitchenOrder, error)
	GetAllOrders(ctx context.Context) ([]*domain.KitchenOrder, error)
	DeleteOrder(ctx context.Context, orderID string) error
	GetStats(ctx context.Context) (*domain.KitchenStats, error)
}

type AggregateRepository interface {
	SaveAggregate(ctx context.Context, aggregate *domain.KitchenAggregate) error
	GetAggregate(ctx context.Context, orderID string) (*domain.KitchenAggregate, error)
	DeleteAggregate(ctx context.Context, orderID string) error
}

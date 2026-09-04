package ports

import (
	"context"
	"delivery-service/src/domain"
)

type AggregateRepository interface {
	SaveAggregate(ctx context.Context, aggregate *domain.DeliveryAggregate) error
	GetAggregate(ctx context.Context, orderID string) (*domain.DeliveryAggregate, error)
	DeleteAggregate(ctx context.Context, orderID string) error
	GetReadyToDeliverOrders(ctx context.Context) ([]*domain.DeliveryAggregate, error)
}

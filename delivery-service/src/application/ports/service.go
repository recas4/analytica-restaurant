package ports

import (
	"context"
	"delivery-service/src/domain"
)

type DeliveryService interface {
	HandleOrderCreated(ctx context.Context, event *domain.OrderCreatedEvent) error
	HandleKitchenStatusChanged(ctx context.Context, event *domain.KitchenStatusChangedEvent) error
	GetOrderStatus(ctx context.Context, orderID string) (*domain.DeliveryAggregate, error)
	UpdateDeliveryStatus(ctx context.Context, orderID string, status domain.DeliveryStatus, message string) error
}

package ports

import (
	"context"
	"kitchen-service/src/domain"
)

type KitchenService interface {
	ReceiveOrder(ctx context.Context, order *domain.KitchenOrder) error
	StartPreparation(ctx context.Context, orderID string) error
	CompleteOrder(ctx context.Context, orderID string) error
	MarkPickedUpByDriver(ctx context.Context, orderID string) error
	CancelOrder(ctx context.Context, orderID string) error
	GetOrderStatus(ctx context.Context, orderID string) (*domain.KitchenOrder, error)
	GetAllOrders(ctx context.Context) ([]*domain.KitchenOrder, error)
	GetKitchenStats(ctx context.Context) (*domain.KitchenStats, error)
	ProcessOrderQueue(ctx context.Context) error
	// Event Aggregation
	HandleOrderCreated(ctx context.Context, orderID, customerID string, items []domain.OrderItem, estimatedTime int) error
	HandlePaymentConfirmed(ctx context.Context, orderID string) error
}

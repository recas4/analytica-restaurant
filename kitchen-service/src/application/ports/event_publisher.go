package ports

import (
	"context"
	"kitchen-service/src/domain"
)

type EventPublisher interface {
	PublishOrderStatusChanged(ctx context.Context, event *domain.OrderStatusChangedEvent) error
	PublishKitchenNotification(ctx context.Context, event *domain.KitchenNotificationEvent) error
}

package ports

import (
	"context"
	"delivery-service/src/domain"
)

type EventPublisher interface {
	PublishDeliveryEvent(ctx context.Context, event *domain.DeliveryEvent) error
	Close() error
}

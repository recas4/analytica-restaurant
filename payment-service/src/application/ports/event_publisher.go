package ports

import (
	"context"
	"payment-service/src/domain"
)

type EventPublisher interface {
	PublishPaymentSucceeded(ctx context.Context, event *domain.PaymentSucceededEvent) error
	PublishPaymentFailed(ctx context.Context, event *domain.PaymentFailedEvent) error
}

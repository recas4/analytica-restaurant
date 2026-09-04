package ports

import (
	"context"
	"payment-service/src/domain"

	"github.com/google/uuid"
)

type PaymentService interface {
	CreatePayment(ctx context.Context, p domain.Payment) (*domain.Payment, error)
	GetPayment(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	GetAllPayments(ctx context.Context) ([]domain.Payment, error)
	UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status domain.Status) (*domain.Payment, error)
	DeletePayment(ctx context.Context, id uuid.UUID) error
	GetPaymentsByStatus(ctx context.Context, status domain.Status) ([]domain.Payment, error)
	ProcessOrderPayment(ctx context.Context, orderEvent domain.OrderCreatedEvent) (*domain.Payment, error)
}

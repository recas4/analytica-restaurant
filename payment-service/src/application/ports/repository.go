package ports

import (
	"context"
	"payment-service/src/domain"

	"github.com/google/uuid"
)

type PaymentRepository interface {
	Save(ctx context.Context, p *domain.Payment) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	FindAll(ctx context.Context) ([]domain.Payment, error)
	Update(ctx context.Context, p *domain.Payment) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByStatus(ctx context.Context, status domain.Status) ([]domain.Payment, error)
}

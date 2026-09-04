package outgoing

import (
	"context"
	"errors"
	"payment-service/src/application/ports"
	"payment-service/src/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentEntity struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrderID   string    `gorm:"size:64;not null"`
	UserID    string    `gorm:"size:64;not null"`
	Provider  string    `gorm:"size:32;not null"`
	Amount    float64   `gorm:"not null"`
	Currency  string    `gorm:"size:8;not null"`
	Status    string    `gorm:"size:16;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (PaymentEntity) TableName() string {
	return "payments"
}

func (e *PaymentEntity) ToDomain() *domain.Payment {
	return &domain.Payment{
		ID:        e.ID,
		OrderID:   e.OrderID,
		UserID:    e.UserID,
		Provider:  e.Provider,
		Amount:    e.Amount,
		Currency:  e.Currency,
		Status:    domain.Status(e.Status),
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func FromDomain(p *domain.Payment) *PaymentEntity {
	return &PaymentEntity{
		ID:        p.ID,
		OrderID:   p.OrderID,
		UserID:    p.UserID,
		Provider:  p.Provider,
		Amount:    p.Amount,
		Currency:  p.Currency,
		Status:    string(p.Status),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) ports.PaymentRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(ctx context.Context, payment *domain.Payment) error {
	entity := FromDomain(payment)
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return err
	}

	domainModel := entity.ToDomain()
	*payment = *domainModel
	return nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	var entity PaymentEntity
	if err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}
	return entity.ToDomain(), nil
}

func (r *PostgresRepository) FindAll(ctx context.Context) ([]domain.Payment, error) {
	var entities []PaymentEntity
	if err := r.db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Payment, len(entities))
	for index, entity := range entities {
		domainModel := entity.ToDomain()
		result[index] = *domainModel
	}
	return result, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&PaymentEntity{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("payment not found")
	}
	return nil
}

func (r *PostgresRepository) FindByStatus(ctx context.Context, status domain.Status) ([]domain.Payment, error) {
	var entities []PaymentEntity
	if err := r.db.WithContext(ctx).Where("status = ?",
		string(status)).Find(&entities).Error; err != nil {
		return nil, err
	}

	result := make([]domain.Payment, len(entities))
	for i, entity := range entities {
		domainModel := entity.ToDomain()
		result[i] = *domainModel
	}
	return result, nil
}

func (r *PostgresRepository) Update(ctx context.Context, payment *domain.Payment) error {
	entity := FromDomain(payment)
	entity.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Save(entity)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("payment not found")
	}
	domainModel := entity.ToDomain()
	*payment = *domainModel
	return nil
}

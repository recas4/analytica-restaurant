package services

import (
	"context"
	"fmt"
	"log"
	"payment-service/src/application/ports"
	"payment-service/src/domain"
	"time"

	"github.com/google/uuid"
)

type PaymentService struct {
	repo           ports.PaymentRepository
	eventPublisher ports.EventPublisher
}

func NewPaymentService(r ports.PaymentRepository, ep ports.EventPublisher) *PaymentService {
	return &PaymentService{
		repo:           r,
		eventPublisher: ep,
	}
}

func (s *PaymentService) CreatePayment(ctx context.Context, p domain.Payment) (*domain.Payment, error) {
	payment, err := domain.NewPayment(p.OrderID, p.UserID, p.Provider,
		p.Amount, p.Currency)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, payment); err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *PaymentService) GetAllPayments(ctx context.Context) ([]domain.Payment, error) {
	return s.repo.FindAll(ctx)
}

func (s *PaymentService) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status domain.Status) (*domain.Payment, error) {
	payment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	switch status {
	case domain.StatusSuccess:
		if err := payment.MarkSuccess(); err != nil {
			return nil, err
		}
	case domain.StatusFailed:
		if err := payment.MarkFailed(); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid status transition to: %s", status)
	}

	if err := s.repo.Update(ctx, payment); err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *PaymentService) DeletePayment(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *PaymentService) GetPaymentsByStatus(ctx context.Context, status domain.Status) ([]domain.Payment, error) {
	return s.repo.FindByStatus(ctx, status)
}

func (s *PaymentService) ProcessOrderPayment(ctx context.Context, orderEvent domain.OrderCreatedEvent) (*domain.Payment, error) {
	log.Printf("Processing payment for order %s, amount: %.2f %s", orderEvent.OrderID,
		orderEvent.TotalAmount, orderEvent.Currency)

	payment := domain.Payment{
		OrderID:  orderEvent.OrderID,
		UserID:   orderEvent.UserID,
		Provider: orderEvent.PaymentProvider,
		Amount:   orderEvent.TotalAmount,
		Currency: orderEvent.Currency,
	}

	createdPayment, err := s.CreatePayment(ctx, payment)
	if err != nil {
		failedEvent := &domain.PaymentFailedEvent{
			EventType: "payment_failed",
			OrderID:   orderEvent.OrderID,
			Reason:    err.Error(),
			Amount:    orderEvent.TotalAmount,
			Currency:  orderEvent.Currency,
			Timestamp: time.Now(),
		}
		s.eventPublisher.PublishPaymentFailed(ctx, failedEvent)
		return nil, err
	}

	// Simulate payment processing (in real world, this would call external payment provider)

	// Different delays by provider
	var processingTime time.Duration
	switch createdPayment.Provider {
	case "stripe":
		processingTime = 2 * time.Second
	case "paypal":
		processingTime = 4 * time.Second
	case "visa":
		processingTime = 3 * time.Second
	default:
		processingTime = 3 * time.Second
	}

	log.Printf("Processing payment with %s provider, will take %v...", createdPayment.Provider, processingTime)
	time.Sleep(processingTime)
	log.Printf("Payment processing delay completed for %s", createdPayment.Provider)

	// Update payment status to success
	updatedPayment, err := s.UpdatePaymentStatus(ctx, createdPayment.ID, domain.StatusSuccess)
	if err != nil {
		log.Printf("Failed to update payment status to success: %v", err)
		// Still publish failed event
		failedEvent := &domain.PaymentFailedEvent{
			EventType: "payment_failed",
			OrderID:   orderEvent.OrderID,
			Reason:    fmt.Sprintf("failed to update payment status: %v", err),
			Amount:    orderEvent.TotalAmount,
			Currency:  orderEvent.Currency,
			Timestamp: time.Now(),
		}
		s.eventPublisher.PublishPaymentFailed(ctx, failedEvent)
		return createdPayment, err
	}

	// Publish payment succeeded event
	successEvent := &domain.PaymentSucceededEvent{
		EventType: "payment_succeeded",
		OrderID:   orderEvent.OrderID,
		PaymentID: updatedPayment.ID.String(),
		Amount:    updatedPayment.Amount,
		Currency:  updatedPayment.Currency,
		Timestamp: time.Now(),
	}

	if err := s.eventPublisher.PublishPaymentSucceeded(ctx, successEvent); err != nil {
		log.Printf("Failed to publish payment succeeded event: %v", err)
	}

	log.Printf("Payment %s successfully processed and marked as success", updatedPayment.ID.String())
	return updatedPayment, nil
}

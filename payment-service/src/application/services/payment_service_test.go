package services

import (
	"context"
	"errors"
	"payment-service/src/domain"
	"testing"

	"github.com/google/uuid"
)

// Mock Repository
type mockPaymentRepository struct {
	payments    map[uuid.UUID]*domain.Payment
	shouldError bool
}

func newMockPaymentRepository() *mockPaymentRepository {
	return &mockPaymentRepository{
		payments: make(map[uuid.UUID]*domain.Payment),
	}
}

func (m *mockPaymentRepository) Save(ctx context.Context, p *domain.Payment) error {
	if m.shouldError {
		return errors.New("database error")
	}
	m.payments[p.ID] = p
	return nil
}

func (m *mockPaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	if m.shouldError {
		return nil, errors.New("database error")
	}
	payment, exists := m.payments[id]
	if !exists {
		return nil, errors.New("payment not found")
	}
	return payment, nil
}

func (m *mockPaymentRepository) FindAll(ctx context.Context) ([]domain.Payment, error) {
	if m.shouldError {
		return nil, errors.New("database error")
	}
	var payments []domain.Payment
	for _, p := range m.payments {
		payments = append(payments, *p)
	}
	return payments, nil
}

func (m *mockPaymentRepository) Update(ctx context.Context, p *domain.Payment) error {
	if m.shouldError {
		return errors.New("database error")
	}
	m.payments[p.ID] = p
	return nil
}

func (m *mockPaymentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.shouldError {
		return errors.New("database error")
	}
	delete(m.payments, id)
	return nil
}

func (m *mockPaymentRepository) FindByStatus(ctx context.Context, status domain.Status) ([]domain.Payment, error) {
	if m.shouldError {
		return nil, errors.New("database error")
	}
	var payments []domain.Payment
	for _, p := range m.payments {
		if p.Status == status {
			payments = append(payments, *p)
		}
	}
	return payments, nil
}

// Mock Event Publisher
type mockEventPublisher struct {
	shouldError bool
	events      []interface{}
}

func (m *mockEventPublisher) PublishPaymentSucceeded(ctx context.Context, event *domain.PaymentSucceededEvent) error {
	if m.shouldError {
		return errors.New("publish error")
	}
	m.events = append(m.events, event)
	return nil
}

func (m *mockEventPublisher) PublishPaymentFailed(ctx context.Context, event *domain.PaymentFailedEvent) error {
	if m.shouldError {
		return errors.New("publish error")
	}
	m.events = append(m.events, event)
	return nil
}

// Tests

func TestNewPaymentService(t *testing.T) {
	repo := newMockPaymentRepository()
	publisher := &mockEventPublisher{}

	service := NewPaymentService(repo, publisher)

	if service == nil {
		t.Error("NewPaymentService should not return nil")
	}
}

func TestCreatePayment_Success(t *testing.T) {
	repo := newMockPaymentRepository()
	publisher := &mockEventPublisher{}
	service := NewPaymentService(repo, publisher)
	ctx := context.Background()

	payment := domain.Payment{
		OrderID:  "order-123",
		UserID:   "user-456",
		Provider: "stripe",
		Amount:   100.00,
		Currency: "EUR",
	}

	result, err := service.CreatePayment(ctx, payment)

	if err != nil {
		t.Errorf("CreatePayment should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Status != domain.StatusPending {
		t.Errorf("Expected status 'pending', got '%s'", result.Status)
	}

	if len(repo.payments) != 1 {
		t.Errorf("Expected 1 payment in repository, got %d", len(repo.payments))
	}
}

func TestCreatePayment_InvalidAmount(t *testing.T) {
	repo := newMockPaymentRepository()
	publisher := &mockEventPublisher{}
	service := NewPaymentService(repo, publisher)
	ctx := context.Background()

	payment := domain.Payment{
		OrderID:  "order-123",
		UserID:   "user-456",
		Provider: "stripe",
		Amount:   -50.00,
		Currency: "EUR",
	}

	_, err := service.CreatePayment(ctx, payment)

	if err == nil {
		t.Error("CreatePayment with negative amount should fail")
	}
}

func TestGetPayment_Success(t *testing.T) {
	repo := newMockPaymentRepository()
	publisher := &mockEventPublisher{}
	service := NewPaymentService(repo, publisher)
	ctx := context.Background()

	// Create a payment first
	payment := domain.Payment{
		OrderID:  "order-123",
		UserID:   "user-456",
		Provider: "stripe",
		Amount:   75.50,
		Currency: "EUR",
	}
	created, _ := service.CreatePayment(ctx, payment)

	// Get the payment
	result, err := service.GetPayment(ctx, created.ID)

	if err != nil {
		t.Errorf("GetPayment should succeed, got error: %v", err)
	}

	if result.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, result.ID)
	}
}

func TestGetPayment_NotFound(t *testing.T) {
	repo := newMockPaymentRepository()
	publisher := &mockEventPublisher{}
	service := NewPaymentService(repo, publisher)
	ctx := context.Background()

	_, err := service.GetPayment(ctx, uuid.New())

	if err == nil {
		t.Error("GetPayment for non-existent ID should fail")
	}
}

func TestGetAllPayments(t *testing.T) {
	repo := newMockPaymentRepository()
	publisher := &mockEventPublisher{}
	service := NewPaymentService(repo, publisher)
	ctx := context.Background()

	// Create multiple payments
	for i := 0; i < 3; i++ {
		payment := domain.Payment{
			OrderID:  "order-" + string(rune('1'+i)),
			UserID:   "user-456",
			Provider: "stripe",
			Amount:   float64(i+1) * 10.00,
			Currency: "EUR",
		}
		_, _ = service.CreatePayment(ctx, payment)
	}

	payments, err := service.GetAllPayments(ctx)

	if err != nil {
		t.Errorf("GetAllPayments should succeed, got error: %v", err)
	}

	if len(payments) != 3 {
		t.Errorf("Expected 3 payments, got %d", len(payments))
	}
}

func TestUpdatePaymentStatus_ToSuccess(t *testing.T) {
	repo := newMockPaymentRepository()
	publisher := &mockEventPublisher{}
	service := NewPaymentService(repo, publisher)
	ctx := context.Background()

	payment := domain.Payment{
		OrderID:  "order-123",
		UserID:   "user-456",
		Provider: "stripe",
		Amount:   50.00,
		Currency: "EUR",
	}
	created, _ := service.CreatePayment(ctx, payment)

	result, err := service.UpdatePaymentStatus(ctx, created.ID, domain.StatusSuccess)

	if err != nil {
		t.Errorf("UpdatePaymentStatus should succeed, got error: %v", err)
	}

	if result.Status != domain.StatusSuccess {
		t.Errorf("Expected status 'success', got '%s'", result.Status)
	}
}

func TestUpdatePaymentStatus_ToFailed(t *testing.T) {
	repo := newMockPaymentRepository()
	publisher := &mockEventPublisher{}
	service := NewPaymentService(repo, publisher)
	ctx := context.Background()

	payment := domain.Payment{
		OrderID:  "order-123",
		UserID:   "user-456",
		Provider: "stripe",
		Amount:   50.00,
		Currency: "EUR",
	}
	created, _ := service.CreatePayment(ctx, payment)

	result, err := service.UpdatePaymentStatus(ctx, created.ID, domain.StatusFailed)

	if err != nil {
		t.Errorf("UpdatePaymentStatus should succeed, got error: %v", err)
	}

	if result.Status != domain.StatusFailed {
		t.Errorf("Expected status 'failed', got '%s'", result.Status)
	}
}

func TestUpdatePaymentStatus_InvalidTransition(t *testing.T) {
	repo := newMockPaymentRepository()
	publisher := &mockEventPublisher{}
	service := NewPaymentService(repo, publisher)
	ctx := context.Background()

	payment := domain.Payment{
		OrderID:  "order-123",
		UserID:   "user-456",
		Provider: "stripe",
		Amount:   50.00,
		Currency: "EUR",
	}
	created, _ := service.CreatePayment(ctx, payment)

	// First mark as success
	_, _ = service.UpdatePaymentStatus(ctx, created.ID, domain.StatusSuccess)

	// Try to mark as failed (should fail)
	_, err := service.UpdatePaymentStatus(ctx, created.ID, domain.StatusFailed)

	if err == nil {
		t.Error("Updating already successful payment should fail")
	}
}

func TestDeletePayment(t *testing.T) {
	repo := newMockPaymentRepository()
	publisher := &mockEventPublisher{}
	service := NewPaymentService(repo, publisher)
	ctx := context.Background()

	payment := domain.Payment{
		OrderID:  "order-123",
		UserID:   "user-456",
		Provider: "stripe",
		Amount:   50.00,
		Currency: "EUR",
	}
	created, _ := service.CreatePayment(ctx, payment)

	err := service.DeletePayment(ctx, created.ID)

	if err != nil {
		t.Errorf("DeletePayment should succeed, got error: %v", err)
	}

	if len(repo.payments) != 0 {
		t.Errorf("Expected 0 payments after delete, got %d", len(repo.payments))
	}
}

func TestGetPaymentsByStatus(t *testing.T) {
	repo := newMockPaymentRepository()
	publisher := &mockEventPublisher{}
	service := NewPaymentService(repo, publisher)
	ctx := context.Background()

	// Create payments
	for i := 0; i < 3; i++ {
		payment := domain.Payment{
			OrderID:  "order-" + string(rune('1'+i)),
			UserID:   "user-456",
			Provider: "stripe",
			Amount:   float64(i+1) * 10.00,
			Currency: "EUR",
		}
		created, _ := service.CreatePayment(ctx, payment)
		if i == 0 {
			_, _ = service.UpdatePaymentStatus(ctx, created.ID, domain.StatusSuccess)
		}
	}

	// Get pending payments
	pendingPayments, err := service.GetPaymentsByStatus(ctx, domain.StatusPending)

	if err != nil {
		t.Errorf("GetPaymentsByStatus should succeed, got error: %v", err)
	}

	if len(pendingPayments) != 2 {
		t.Errorf("Expected 2 pending payments, got %d", len(pendingPayments))
	}

	// Get successful payments
	successPayments, _ := service.GetPaymentsByStatus(ctx, domain.StatusSuccess)

	if len(successPayments) != 1 {
		t.Errorf("Expected 1 successful payment, got %d", len(successPayments))
	}
}

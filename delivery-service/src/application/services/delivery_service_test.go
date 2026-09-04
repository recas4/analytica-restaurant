package services

import (
	"context"
	"delivery-service/src/domain"
	"errors"
	"testing"
)

// Mock implementations for testing

type mockAggregateRepository struct {
	aggregates  map[string]*domain.DeliveryAggregate
	shouldError bool
}

func newMockAggregateRepository() *mockAggregateRepository {
	return &mockAggregateRepository{
		aggregates: make(map[string]*domain.DeliveryAggregate),
	}
}

func (m *mockAggregateRepository) SaveAggregate(ctx context.Context, aggregate *domain.DeliveryAggregate) error {
	if m.shouldError {
		return errors.New("database error")
	}
	m.aggregates[aggregate.OrderID] = aggregate
	return nil
}

func (m *mockAggregateRepository) GetAggregate(ctx context.Context, orderID string) (*domain.DeliveryAggregate, error) {
	if m.shouldError {
		return nil, errors.New("database error")
	}
	aggregate, exists := m.aggregates[orderID]
	if !exists {
		return nil, nil
	}
	return aggregate, nil
}

func (m *mockAggregateRepository) DeleteAggregate(ctx context.Context, orderID string) error {
	if m.shouldError {
		return errors.New("database error")
	}
	delete(m.aggregates, orderID)
	return nil
}

func (m *mockAggregateRepository) GetReadyToDeliverOrders(ctx context.Context) ([]*domain.DeliveryAggregate, error) {
	if m.shouldError {
		return nil, errors.New("database error")
	}
	var orders []*domain.DeliveryAggregate
	for _, aggregate := range m.aggregates {
		if aggregate.Status == domain.StatusReadyToDeliver {
			orders = append(orders, aggregate)
		}
	}
	return orders, nil
}

type mockEventPublisher struct {
	shouldError    bool
	deliveryEvents []*domain.DeliveryEvent
}

func newMockEventPublisher() *mockEventPublisher {
	return &mockEventPublisher{
		deliveryEvents: []*domain.DeliveryEvent{},
	}
}

func (m *mockEventPublisher) PublishDeliveryEvent(ctx context.Context, event *domain.DeliveryEvent) error {
	if m.shouldError {
		return errors.New("publish error")
	}
	m.deliveryEvents = append(m.deliveryEvents, event)
	return nil
}

func (m *mockEventPublisher) Close() error {
	return nil
}

// Tests

func TestNewDeliveryService(t *testing.T) {
	repo := newMockAggregateRepository()
	publisher := newMockEventPublisher()

	service := NewDeliveryService(repo, publisher)

	if service == nil {
		t.Error("NewDeliveryService should not return nil")
	}
}

func TestHandleOrderCreated_NewAggregate(t *testing.T) {
	repo := newMockAggregateRepository()
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	event := &domain.OrderCreatedEvent{
		OrderID:     "order-123",
		UserID:      "user-456",
		OrderType:   "delivery",
		TotalAmount: 50.00,
		Currency:    "EUR",
		DeliveryInfo: &domain.DeliveryInfo{
			CustomerName:  "John Doe",
			CustomerPhone: "123456789",
			Street:        "Main St",
			HouseNumber:   "123",
			PostalCode:    "12345",
			City:          "Berlin",
		},
	}

	err := service.HandleOrderCreated(ctx, event)

	if err != nil {
		t.Errorf("HandleOrderCreated should succeed, got error: %v", err)
	}

	aggregate := repo.aggregates["order-123"]
	if aggregate == nil {
		t.Fatal("Aggregate should be created")
	}

	if !aggregate.OrderReceived {
		t.Error("OrderReceived should be true")
	}

	if aggregate.UserID != "user-456" {
		t.Errorf("Expected UserID 'user-456', got '%s'", aggregate.UserID)
	}

	if aggregate.OrderType != "delivery" {
		t.Errorf("Expected OrderType 'delivery', got '%s'", aggregate.OrderType)
	}

	if aggregate.Status != domain.StatusWaitingForKitchen {
		t.Errorf("Expected status 'waiting_for_kitchen', got '%s'", aggregate.Status)
	}
}

func TestHandleOrderCreated_PickupOrder(t *testing.T) {
	repo := newMockAggregateRepository()
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	event := &domain.OrderCreatedEvent{
		OrderID:     "order-123",
		UserID:      "user-456",
		OrderType:   "pickup",
		TotalAmount: 50.00,
		Currency:    "EUR",
	}

	err := service.HandleOrderCreated(ctx, event)

	if err != nil {
		t.Errorf("HandleOrderCreated should succeed, got error: %v", err)
	}

	aggregate := repo.aggregates["order-123"]
	if aggregate.OrderType != "pickup" {
		t.Errorf("Expected OrderType 'pickup', got '%s'", aggregate.OrderType)
	}
}

func TestHandleKitchenStatusChanged_ReadyStatus(t *testing.T) {
	repo := newMockAggregateRepository()
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	event := &domain.KitchenStatusChangedEvent{
		EventType: "order_ready",
		OrderID:   "order-123",
		Status:    "ready",
	}

	err := service.HandleKitchenStatusChanged(ctx, event)

	if err != nil {
		t.Errorf("HandleKitchenStatusChanged should succeed, got error: %v", err)
	}

	aggregate := repo.aggregates["order-123"]
	if aggregate == nil {
		t.Fatal("Aggregate should be created")
	}

	if !aggregate.KitchenReady {
		t.Error("KitchenReady should be true")
	}

	if aggregate.Status != domain.StatusWaitingForOrder {
		t.Errorf("Expected status 'waiting_for_order', got '%s'", aggregate.Status)
	}
}

func TestHandleKitchenStatusChanged_IgnoreNonReadyStatus(t *testing.T) {
	repo := newMockAggregateRepository()
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	event := &domain.KitchenStatusChangedEvent{
		EventType: "order_preparing",
		OrderID:   "order-123",
		Status:    "preparing",
	}

	err := service.HandleKitchenStatusChanged(ctx, event)

	if err != nil {
		t.Errorf("HandleKitchenStatusChanged should succeed, got error: %v", err)
	}

	aggregate := repo.aggregates["order-123"]
	if aggregate != nil {
		t.Error("Aggregate should not be created for non-ready status")
	}
}

func TestHandleOrderCreated_ThenKitchenReady_StartsDelivery(t *testing.T) {
	repo := newMockAggregateRepository()
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	orderEvent := &domain.OrderCreatedEvent{
		OrderID:     "order-123",
		UserID:      "user-456",
		OrderType:   "delivery",
		TotalAmount: 50.00,
		Currency:    "EUR",
		DeliveryInfo: &domain.DeliveryInfo{
			CustomerName:  "John Doe",
			CustomerPhone: "123456789",
			Street:        "Main St",
			HouseNumber:   "123",
			PostalCode:    "12345",
			City:          "Berlin",
		},
	}
	service.HandleOrderCreated(ctx, orderEvent)

	kitchenEvent := &domain.KitchenStatusChangedEvent{
		EventType: "order_ready",
		OrderID:   "order-123",
		Status:    "ready",
	}
	err := service.HandleKitchenStatusChanged(ctx, kitchenEvent)

	if err != nil {
		t.Errorf("HandleKitchenStatusChanged should succeed, got error: %v", err)
	}

	aggregate := repo.aggregates["order-123"]
	if !aggregate.DeliveryStarted {
		t.Error("DeliveryStarted should be true")
	}

	if aggregate.Status != domain.StatusReadyToDeliver {
		t.Errorf("Expected status 'ready_to_deliver', got '%s'", aggregate.Status)
	}

	if aggregate.DriverID == nil {
		t.Error("DriverID should be assigned")
	}

	if len(publisher.deliveryEvents) == 0 {
		t.Error("Expected delivery event to be published")
	}

	if publisher.deliveryEvents[0].Status != domain.DeliveryStatusAssigned {
		t.Errorf("Expected status 'assigned', got '%s'", publisher.deliveryEvents[0].Status)
	}
}

func TestHandleOrderCreated_ThenKitchenReady_PickupOrder(t *testing.T) {
	repo := newMockAggregateRepository()
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	orderEvent := &domain.OrderCreatedEvent{
		OrderID:     "order-123",
		UserID:      "user-456",
		OrderType:   "pickup",
		TotalAmount: 50.00,
		Currency:    "EUR",
	}
	service.HandleOrderCreated(ctx, orderEvent)

	kitchenEvent := &domain.KitchenStatusChangedEvent{
		EventType: "order_ready",
		OrderID:   "order-123",
		Status:    "ready",
	}
	err := service.HandleKitchenStatusChanged(ctx, kitchenEvent)

	if err != nil {
		t.Errorf("HandleKitchenStatusChanged should succeed, got error: %v", err)
	}

	aggregate := repo.aggregates["order-123"]
	if aggregate.Status != domain.StatusPickupReady {
		t.Errorf("Expected status 'pickup_ready', got '%s'", aggregate.Status)
	}

	if len(publisher.deliveryEvents) == 0 {
		t.Error("Expected delivery event to be published")
	}

	if publisher.deliveryEvents[0].Status != domain.DeliveryStatusPickupReady {
		t.Errorf("Expected status 'pickup_ready', got '%s'", publisher.deliveryEvents[0].Status)
	}
}

func TestGetOrderStatus_Success(t *testing.T) {
	repo := newMockAggregateRepository()
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	aggregate := &domain.DeliveryAggregate{
		OrderID:       "order-123",
		OrderReceived: true,
		KitchenReady:  true,
		OrderType:     "delivery",
	}
	repo.aggregates["order-123"] = aggregate

	result, err := service.GetOrderStatus(ctx, "order-123")

	if err != nil {
		t.Errorf("GetOrderStatus should succeed, got error: %v", err)
	}

	if result.OrderID != "order-123" {
		t.Errorf("Expected OrderID 'order-123', got '%s'", result.OrderID)
	}
}

func TestGetOrderStatus_NotFound(t *testing.T) {
	repo := newMockAggregateRepository()
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	result, err := service.GetOrderStatus(ctx, "nonexistent-order")

	if err != nil {
		t.Errorf("GetOrderStatus should not error for non-existent order, got error: %v", err)
	}

	if result != nil {
		t.Error("Result should be nil for non-existent order")
	}
}

func TestUpdateDeliveryStatus_Success(t *testing.T) {
	repo := newMockAggregateRepository()
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	aggregate := &domain.DeliveryAggregate{
		OrderID:         "order-123",
		OrderReceived:   true,
		KitchenReady:    true,
		DeliveryStarted: true,
		OrderType:       "delivery",
	}
	repo.aggregates["order-123"] = aggregate

	err := service.UpdateDeliveryStatus(ctx, "order-123", domain.DeliveryStatusInTransit, "Driver picked up")

	if err != nil {
		t.Errorf("UpdateDeliveryStatus should succeed, got error: %v", err)
	}

	if len(publisher.deliveryEvents) == 0 {
		t.Error("Expected delivery event to be published")
	}

	event := publisher.deliveryEvents[0]
	if event.Status != domain.DeliveryStatusInTransit {
		t.Errorf("Expected status 'in_transit', got '%s'", event.Status)
	}

	if event.Message != "Driver picked up" {
		t.Errorf("Expected message 'Driver picked up', got '%s'", event.Message)
	}
}

func TestUpdateDeliveryStatus_OrderNotFound(t *testing.T) {
	repo := newMockAggregateRepository()
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	err := service.UpdateDeliveryStatus(ctx, "nonexistent-order", domain.DeliveryStatusInTransit, "Test")

	if err == nil {
		t.Error("UpdateDeliveryStatus should fail when order not found")
	}
}

func TestUpdateDeliveryStatus_NotDeliveryOrder(t *testing.T) {
	repo := newMockAggregateRepository()
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	aggregate := &domain.DeliveryAggregate{
		OrderID:   "order-123",
		OrderType: "pickup",
	}
	repo.aggregates["order-123"] = aggregate

	err := service.UpdateDeliveryStatus(ctx, "order-123", domain.DeliveryStatusInTransit, "Test")

	if err == nil {
		t.Error("UpdateDeliveryStatus should fail for pickup orders")
	}
}

func TestHandleOrderCreated_WithoutDeliveryInfo(t *testing.T) {
	repo := newMockAggregateRepository()
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	orderEvent := &domain.OrderCreatedEvent{
		OrderID:      "order-123",
		UserID:       "user-456",
		OrderType:    "delivery",
		TotalAmount:  50.00,
		Currency:     "EUR",
		DeliveryInfo: nil,
	}
	service.HandleOrderCreated(ctx, orderEvent)

	kitchenEvent := &domain.KitchenStatusChangedEvent{
		EventType: "order_ready",
		OrderID:   "order-123",
		Status:    "ready",
	}
	err := service.HandleKitchenStatusChanged(ctx, kitchenEvent)

	if err != nil {
		t.Errorf("HandleKitchenStatusChanged should succeed even without DeliveryInfo, got error: %v", err)
	}

	if len(publisher.deliveryEvents) == 0 {
		t.Error("Expected delivery event to be published")
	}
}

func TestHandleOrderCreated_RepositoryError(t *testing.T) {
	repo := newMockAggregateRepository()
	repo.shouldError = true
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	event := &domain.OrderCreatedEvent{
		OrderID:     "order-123",
		UserID:      "user-456",
		OrderType:   "delivery",
		TotalAmount: 50.00,
		Currency:    "EUR",
	}

	err := service.HandleOrderCreated(ctx, event)

	if err == nil {
		t.Error("HandleOrderCreated should fail when repository errors")
	}
}

func TestHandleKitchenStatusChanged_RepositoryError(t *testing.T) {
	repo := newMockAggregateRepository()
	repo.shouldError = true
	publisher := newMockEventPublisher()
	service := NewDeliveryService(repo, publisher)
	ctx := context.Background()

	event := &domain.KitchenStatusChangedEvent{
		EventType: "order_ready",
		OrderID:   "order-123",
		Status:    "ready",
	}

	err := service.HandleKitchenStatusChanged(ctx, event)

	if err == nil {
		t.Error("HandleKitchenStatusChanged should fail when repository errors")
	}
}

package services

import (
	"context"
	"errors"
	"kitchen-service/src/domain"
	"testing"
)

// Mock implementations for testing

type mockKitchenRepository struct {
	orders      map[string]*domain.KitchenOrder
	shouldError bool
}

func newMockKitchenRepository() *mockKitchenRepository {
	return &mockKitchenRepository{
		orders: make(map[string]*domain.KitchenOrder),
	}
}

func (m *mockKitchenRepository) SaveOrder(ctx context.Context, order *domain.KitchenOrder) error {
	if m.shouldError {
		return errors.New("database error")
	}
	m.orders[order.OrderID] = order
	return nil
}

func (m *mockKitchenRepository) UpdateOrder(ctx context.Context, order *domain.KitchenOrder) error {
	if m.shouldError {
		return errors.New("database error")
	}
	m.orders[order.OrderID] = order
	return nil
}

func (m *mockKitchenRepository) GetOrderByID(ctx context.Context, orderID string) (*domain.KitchenOrder, error) {
	if m.shouldError {
		return nil, errors.New("database error")
	}
	order, exists := m.orders[orderID]
	if !exists {
		return nil, errors.New("order not found")
	}
	return order, nil
}

func (m *mockKitchenRepository) GetOrdersByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.KitchenOrder, error) {
	if m.shouldError {
		return nil, errors.New("database error")
	}
	var orders []*domain.KitchenOrder
	for _, order := range m.orders {
		if order.Status == status {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

func (m *mockKitchenRepository) GetAllOrders(ctx context.Context) ([]*domain.KitchenOrder, error) {
	if m.shouldError {
		return nil, errors.New("database error")
	}
	var orders []*domain.KitchenOrder
	for _, order := range m.orders {
		orders = append(orders, order)
	}
	return orders, nil
}

func (m *mockKitchenRepository) DeleteOrder(ctx context.Context, orderID string) error {
	if m.shouldError {
		return errors.New("database error")
	}
	delete(m.orders, orderID)
	return nil
}

func (m *mockKitchenRepository) GetStats(ctx context.Context) (*domain.KitchenStats, error) {
	if m.shouldError {
		return nil, errors.New("database error")
	}
	return &domain.KitchenStats{
		TotalOrders: len(m.orders),
	}, nil
}

type mockAggregateRepository struct {
	aggregates  map[string]*domain.KitchenAggregate
	shouldError bool
}

func newMockAggregateRepository() *mockAggregateRepository {
	return &mockAggregateRepository{
		aggregates: make(map[string]*domain.KitchenAggregate),
	}
}

func (m *mockAggregateRepository) SaveAggregate(ctx context.Context, aggregate *domain.KitchenAggregate) error {
	if m.shouldError {
		return errors.New("database error")
	}
	m.aggregates[aggregate.OrderID] = aggregate
	return nil
}

func (m *mockAggregateRepository) GetAggregate(ctx context.Context, orderID string) (*domain.KitchenAggregate, error) {
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

type mockEventPublisher struct {
	shouldError        bool
	statusEvents       []*domain.OrderStatusChangedEvent
	notificationEvents []*domain.KitchenNotificationEvent
}

func newMockEventPublisher() *mockEventPublisher {
	return &mockEventPublisher{
		statusEvents:       []*domain.OrderStatusChangedEvent{},
		notificationEvents: []*domain.KitchenNotificationEvent{},
	}
}

func (m *mockEventPublisher) PublishOrderStatusChanged(ctx context.Context, event *domain.OrderStatusChangedEvent) error {
	if m.shouldError {
		return errors.New("publish error")
	}
	m.statusEvents = append(m.statusEvents, event)
	return nil
}

func (m *mockEventPublisher) PublishKitchenNotification(ctx context.Context, event *domain.KitchenNotificationEvent) error {
	if m.shouldError {
		return errors.New("publish error")
	}
	m.notificationEvents = append(m.notificationEvents, event)
	return nil
}

// Tests

func TestNewKitchenService(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()

	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)

	if service == nil {
		t.Error("NewKitchenService should not return nil")
	}
}

func TestReceiveOrder_Success(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	order := &domain.KitchenOrder{
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		EstimatedTime: 300,
		Items: []domain.OrderItem{
			{ProductID: "prod-1", ProductName: "Pizza", Quantity: 1, Price: 10.99, PrepTime: 300},
		},
	}

	err := service.ReceiveOrder(ctx, order)

	if err != nil {
		t.Errorf("ReceiveOrder should succeed, got error: %v", err)
	}

	if order.Status != domain.StatusReceived {
		t.Errorf("Expected status 'received', got '%s'", order.Status)
	}

	if len(repo.orders) != 1 {
		t.Errorf("Expected 1 order in repository, got %d", len(repo.orders))
	}

	if len(publisher.statusEvents) == 0 {
		t.Error("Expected status change event to be published")
	}

	if len(publisher.notificationEvents) == 0 {
		t.Error("Expected notification event to be published")
	}
}

func TestReceiveOrder_RepositoryError(t *testing.T) {
	repo := newMockKitchenRepository()
	repo.shouldError = true
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	order := &domain.KitchenOrder{
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		EstimatedTime: 300,
	}

	err := service.ReceiveOrder(ctx, order)

	if err == nil {
		t.Error("ReceiveOrder should fail when repository errors")
	}
}

func TestStartPreparation_Success(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	order := &domain.KitchenOrder{
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		Status:        domain.StatusReceived,
		EstimatedTime: 300,
	}
	repo.orders[order.OrderID] = order

	err := service.StartPreparation(ctx, order.OrderID)

	if err != nil {
		t.Errorf("StartPreparation should succeed, got error: %v", err)
	}

	updatedOrder := repo.orders[order.OrderID]
	if updatedOrder.Status != domain.StatusPreparing {
		t.Errorf("Expected status 'preparing', got '%s'", updatedOrder.Status)
	}

	if updatedOrder.StartedAt == nil {
		t.Error("StartedAt should be set")
	}
}

func TestStartPreparation_InvalidStatus(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	order := &domain.KitchenOrder{
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		Status:        domain.StatusPreparing,
		EstimatedTime: 300,
	}
	repo.orders[order.OrderID] = order

	err := service.StartPreparation(ctx, order.OrderID)

	if err == nil {
		t.Error("StartPreparation should fail when order is not in received status")
	}
}

func TestStartPreparation_OrderNotFound(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	err := service.StartPreparation(ctx, "nonexistent-order")

	if err == nil {
		t.Error("StartPreparation should fail when order not found")
	}
}

func TestCompleteOrder_Success(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	order := &domain.KitchenOrder{
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		Status:        domain.StatusPreparing,
		EstimatedTime: 300,
	}
	repo.orders[order.OrderID] = order

	err := service.CompleteOrder(ctx, order.OrderID)

	if err != nil {
		t.Errorf("CompleteOrder should succeed, got error: %v", err)
	}

	updatedOrder := repo.orders[order.OrderID]
	if updatedOrder.Status != domain.StatusReady {
		t.Errorf("Expected status 'ready', got '%s'", updatedOrder.Status)
	}

	if updatedOrder.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}
}

func TestCompleteOrder_InvalidStatus(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	order := &domain.KitchenOrder{
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		Status:        domain.StatusReceived,
		EstimatedTime: 300,
	}
	repo.orders[order.OrderID] = order

	err := service.CompleteOrder(ctx, order.OrderID)

	if err == nil {
		t.Error("CompleteOrder should fail when order is not being prepared")
	}
}

func TestMarkPickedUpByDriver_Success(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	order := &domain.KitchenOrder{
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		Status:        domain.StatusReady,
		EstimatedTime: 300,
	}
	repo.orders[order.OrderID] = order

	err := service.MarkPickedUpByDriver(ctx, order.OrderID)

	if err != nil {
		t.Errorf("MarkPickedUpByDriver should succeed, got error: %v", err)
	}

	updatedOrder := repo.orders[order.OrderID]
	if updatedOrder.Status != domain.StatusPickedUpByDriver {
		t.Errorf("Expected status 'picked_up_by_driver', got '%s'", updatedOrder.Status)
	}
}

func TestMarkPickedUpByDriver_InvalidStatus(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	order := &domain.KitchenOrder{
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		Status:        domain.StatusPreparing,
		EstimatedTime: 300,
	}
	repo.orders[order.OrderID] = order

	err := service.MarkPickedUpByDriver(ctx, order.OrderID)

	if err == nil {
		t.Error("MarkPickedUpByDriver should fail when order is not ready")
	}
}

func TestCancelOrder_Success(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	order := &domain.KitchenOrder{
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		Status:        domain.StatusReceived,
		EstimatedTime: 300,
	}
	repo.orders[order.OrderID] = order

	err := service.CancelOrder(ctx, order.OrderID)

	if err != nil {
		t.Errorf("CancelOrder should succeed, got error: %v", err)
	}

	updatedOrder := repo.orders[order.OrderID]
	if updatedOrder.Status != domain.StatusCancelled {
		t.Errorf("Expected status 'cancelled', got '%s'", updatedOrder.Status)
	}
}

func TestCancelOrder_AlreadyPickedUp(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	order := &domain.KitchenOrder{
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		Status:        domain.StatusPickedUpByDriver,
		EstimatedTime: 300,
	}
	repo.orders[order.OrderID] = order

	err := service.CancelOrder(ctx, order.OrderID)

	if err == nil {
		t.Error("CancelOrder should fail when order is already picked up")
	}
}

func TestGetOrderStatus_Success(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	order := &domain.KitchenOrder{
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		Status:        domain.StatusPreparing,
		EstimatedTime: 300,
	}
	repo.orders[order.OrderID] = order

	result, err := service.GetOrderStatus(ctx, order.OrderID)

	if err != nil {
		t.Errorf("GetOrderStatus should succeed, got error: %v", err)
	}

	if result.OrderID != order.OrderID {
		t.Errorf("Expected OrderID %s, got %s", order.OrderID, result.OrderID)
	}
}

func TestGetOrderStatus_NotFound(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	_, err := service.GetOrderStatus(ctx, "nonexistent-order")

	if err == nil {
		t.Error("GetOrderStatus should fail when order not found")
	}
}

func TestGetAllOrders(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		order := &domain.KitchenOrder{
			OrderID:       "order-" + string(rune('1'+i)),
			CustomerID:    "customer-456",
			Status:        domain.StatusReceived,
			EstimatedTime: 300,
		}
		repo.orders[order.OrderID] = order
	}

	orders, err := service.GetAllOrders(ctx)

	if err != nil {
		t.Errorf("GetAllOrders should succeed, got error: %v", err)
	}

	if len(orders) != 3 {
		t.Errorf("Expected 3 orders, got %d", len(orders))
	}
}

func TestGetKitchenStats(t *testing.T) {
	repo := newMockKitchenRepository()
	publisher := newMockEventPublisher()
	service := NewKitchenService(repo, newMockAggregateRepository(), publisher)
	ctx := context.Background()

	order := &domain.KitchenOrder{
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		Status:        domain.StatusReceived,
		EstimatedTime: 300,
	}
	repo.orders[order.OrderID] = order

	stats, err := service.GetKitchenStats(ctx)

	if err != nil {
		t.Errorf("GetKitchenStats should succeed, got error: %v", err)
	}

	if stats == nil {
		t.Error("Stats should not be nil")
	}

	if stats.TotalOrders != 1 {
		t.Errorf("Expected 1 total order, got %d", stats.TotalOrders)
	}
}

package services

import (
	"context"
	"fmt"
	"kitchen-service/src/application/ports"
	"kitchen-service/src/domain"
	"log"
	"math/rand"
	"time"
)

type KitchenService struct {
	repo          ports.KitchenRepository
	aggregateRepo ports.AggregateRepository
	publisher     ports.EventPublisher
}

func NewKitchenService(repo ports.KitchenRepository, aggregateRepo ports.AggregateRepository, publisher ports.EventPublisher) *KitchenService {
	return &KitchenService{
		repo:          repo,
		aggregateRepo: aggregateRepo,
		publisher:     publisher,
	}
}

func (ks *KitchenService) ReceiveOrder(ctx context.Context, order *domain.KitchenOrder) error {
	order.Status = domain.StatusReceived
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	if err := ks.repo.SaveOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to save received order: %w", err)
	}

	event := &domain.OrderStatusChangedEvent{
		EventType:     "order_received_in_kitchen",
		OrderID:       order.OrderID,
		Status:        domain.StatusReceived,
		EstimatedTime: order.EstimatedTime,
		Timestamp:     time.Now(),
	}

	if err := ks.publisher.PublishOrderStatusChanged(ctx, event); err != nil {
		log.Printf("Failed to publish order received event: %v", err)
	}

	notification := &domain.KitchenNotificationEvent{
		EventType: "kitchen_notification",
		OrderID:   order.OrderID,
		Status:    domain.StatusReceived,
		Message:   fmt.Sprintf("Neue Bestellung erhalten! Geschätzte Zubereitungszeit: %d Sekunden", order.EstimatedTime),
		Timestamp: time.Now(),
	}

	if err := ks.publisher.PublishKitchenNotification(ctx, notification); err != nil {
		log.Printf("Failed to publish kitchen notification: %v", err)
	}

	log.Printf("Kitchen received order %s with %d items", order.OrderID, len(order.Items))
	return nil
}

func (ks *KitchenService) StartPreparation(ctx context.Context, orderID string) error {
	order, err := ks.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status != domain.StatusReceived {
		return fmt.Errorf("order %s is not in received status, current: %s", orderID, order.Status)
	}

	now := time.Now()
	order.Status = domain.StatusPreparing
	order.StartedAt = &now
	order.UpdatedAt = now

	if err := ks.repo.UpdateOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	event := &domain.OrderStatusChangedEvent{
		EventType:     "order_preparation_started",
		OrderID:       order.OrderID,
		Status:        domain.StatusPreparing,
		EstimatedTime: order.EstimatedTime,
		Timestamp:     time.Now(),
	}

	if err := ks.publisher.PublishOrderStatusChanged(ctx, event); err != nil {
		log.Printf("Failed to publish preparation started event: %v", err)
	}

	notification := &domain.KitchenNotificationEvent{
		EventType: "kitchen_notification",
		OrderID:   order.OrderID,
		Status:    domain.StatusPreparing,
		Message:   "Die Zubereitung hat begonnen! Der Koch ist am Werk...",
		Timestamp: time.Now(),
	}

	if err := ks.publisher.PublishKitchenNotification(ctx, notification); err != nil {
		log.Printf("Failed to publish kitchen notification: %v", err)
	}

	go ks.simulatePreparation(ctx, order)

	log.Printf("Started preparing order %s", orderID)
	return nil
}

func (ks *KitchenService) simulatePreparation(ctx context.Context, order *domain.KitchenOrder) {
	estimatedSeconds := order.EstimatedTime
	actualSeconds := ks.calculateActualPrepTime(estimatedSeconds)

	log.Printf("Simulating preparation for order %s: estimated %d sec, actual %d sec",
		order.OrderID, estimatedSeconds, actualSeconds)

	preparationSteps := []string{
		"Zutaten werden vorbereitet...",
		"Kochen hat begonnen...",
		"Chef fügt Gewürze hinzu...",
		"Fast fertig...",
	}

	stepDuration := time.Duration(actualSeconds) * time.Second / time.Duration(len(preparationSteps))

	for i, step := range preparationSteps {
		select {
		case <-ctx.Done():
			return
		case <-time.After(stepDuration):
			notification := &domain.KitchenNotificationEvent{
				EventType: "kitchen_notification",
				OrderID:   order.OrderID,
				Status:    domain.StatusPreparing,
				Message:   fmt.Sprintf("%s (Schritt %d/%d)", step, i+1, len(preparationSteps)),
				Timestamp: time.Now(),
			}

			if err := ks.publisher.PublishKitchenNotification(ctx, notification); err != nil {
				log.Printf("Failed to publish preparation step notification: %v", err)
			}
		}
	}

	if err := ks.CompleteOrder(ctx, order.OrderID); err != nil {
		log.Printf("Failed to complete order automatically: %v", err)
	}
}

func (ks *KitchenService) calculateActualPrepTime(prepTimeSeconds int) int {
	variance := 0.2
	minTime := float64(prepTimeSeconds) * (1.0 - variance)
	maxTime := float64(prepTimeSeconds) * (1.0 + variance)

	actualTime := minTime + rand.Float64()*(maxTime-minTime)

	if actualTime < 5 {
		actualTime = 5 // minimum 5 seconds
	}

	return int(actualTime)
}

func (ks *KitchenService) CompleteOrder(ctx context.Context, orderID string) error {
	order, err := ks.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status != domain.StatusPreparing {
		return fmt.Errorf("order %s is not being prepared, current: %s", orderID, order.Status)
	}

	now := time.Now()
	order.Status = domain.StatusReady
	order.CompletedAt = &now
	order.UpdatedAt = now

	if err := ks.repo.UpdateOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	event := &domain.OrderStatusChangedEvent{
		EventType: "order_ready",
		OrderID:   order.OrderID,
		Status:    domain.StatusReady,
		Timestamp: time.Now(),
	}

	if err := ks.publisher.PublishOrderStatusChanged(ctx, event); err != nil {
		log.Printf("Failed to publish order ready event: %v", err)
	}

	notification := &domain.KitchenNotificationEvent{
		EventType: "kitchen_notification",
		OrderID:   order.OrderID,
		Status:    domain.StatusReady,
		Message:   "Bestellung ist fertig! Bereit zur Abholung.",
		Timestamp: time.Now(),
	}

	if err := ks.publisher.PublishKitchenNotification(ctx, notification); err != nil {
		log.Printf("Failed to publish kitchen notification: %v", err)
	}

	log.Printf("Order %s is ready!", orderID)
	return nil
}

func (ks *KitchenService) MarkPickedUpByDriver(ctx context.Context, orderID string) error {
	order, err := ks.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status != domain.StatusReady {
		return fmt.Errorf("order %s is not ready for pickup, current: %s", orderID, order.Status)
	}

	order.Status = domain.StatusPickedUpByDriver
	order.UpdatedAt = time.Now()

	if err := ks.repo.UpdateOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	event := &domain.OrderStatusChangedEvent{
		EventType: "order_picked_up_by_driver",
		OrderID:   order.OrderID,
		Status:    domain.StatusPickedUpByDriver,
		Timestamp: time.Now(),
	}

	if err := ks.publisher.PublishOrderStatusChanged(ctx, event); err != nil {
		log.Printf("Failed to publish order pickup event: %v", err)
	}

	notification := &domain.KitchenNotificationEvent{
		EventType: "kitchen_notification",
		OrderID:   order.OrderID,
		Status:    domain.StatusPickedUpByDriver,
		Message:   "Bestellung wurde vom Fahrer abgeholt! Auf dem Weg zum Kunden.",
		Timestamp: time.Now(),
	}

	if err := ks.publisher.PublishKitchenNotification(ctx, notification); err != nil {
		log.Printf("Failed to publish kitchen notification: %v", err)
	}

	log.Printf("Order %s picked up by driver!", orderID)
	return nil
}

func (ks *KitchenService) CancelOrder(ctx context.Context, orderID string) error {
	order, err := ks.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status == domain.StatusPickedUpByDriver {
		return fmt.Errorf("cannot cancel order %s - already picked up by driver", orderID)
	}

	order.Status = domain.StatusCancelled
	order.UpdatedAt = time.Now()

	if err := ks.repo.UpdateOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	event := &domain.OrderStatusChangedEvent{
		EventType: "order_cancelled_in_kitchen",
		OrderID:   order.OrderID,
		Status:    domain.StatusCancelled,
		Timestamp: time.Now(),
	}

	if err := ks.publisher.PublishOrderStatusChanged(ctx, event); err != nil {
		log.Printf("Failed to publish order cancelled event: %v", err)
	}

	notification := &domain.KitchenNotificationEvent{
		EventType: "kitchen_notification",
		OrderID:   order.OrderID,
		Status:    domain.StatusCancelled,
		Message:   "Bestellung wurde storniert",
		Timestamp: time.Now(),
	}

	if err := ks.publisher.PublishKitchenNotification(ctx, notification); err != nil {
		log.Printf("Failed to publish kitchen notification: %v", err)
	}

	log.Printf("Order %s cancelled", orderID)
	return nil
}

func (ks *KitchenService) GetOrderStatus(ctx context.Context, orderID string) (*domain.KitchenOrder, error) {
	return ks.repo.GetOrderByID(ctx, orderID)
}

func (ks *KitchenService) GetAllOrders(ctx context.Context) ([]*domain.KitchenOrder, error) {
	return ks.repo.GetAllOrders(ctx)
}

func (ks *KitchenService) GetKitchenStats(ctx context.Context) (*domain.KitchenStats, error) {
	return ks.repo.GetStats(ctx)
}

func (ks *KitchenService) ProcessOrderQueue(ctx context.Context) error {
	receivedOrders, err := ks.repo.GetOrdersByStatus(ctx, domain.StatusReceived)
	if err != nil {
		return fmt.Errorf("failed to get received orders: %w", err)
	}

	for _, order := range receivedOrders {
		if err := ks.StartPreparation(ctx, order.OrderID); err != nil {
			log.Printf("Failed to start preparation for order %s: %v", order.OrderID, err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * 2):
		}
	}

	return nil
}

// Event Aggregation Methods

func (ks *KitchenService) HandleOrderCreated(ctx context.Context, orderID, customerID string, items []domain.OrderItem, estimatedTime int) error {
	log.Printf("Processing OrderCreated event for order %s", orderID)

	aggregate, err := ks.aggregateRepo.GetAggregate(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get aggregate: %w", err)
	}

	if aggregate == nil {
		aggregate = &domain.KitchenAggregate{
			OrderID:   orderID,
			CreatedAt: time.Now(),
		}
	}

	aggregate.OrderReceived = true
	aggregate.CustomerID = customerID
	aggregate.Items = items
	aggregate.EstimatedTime = estimatedTime
	aggregate.UpdatedAt = time.Now()

	if err := ks.aggregateRepo.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	log.Printf("Order %s received, waiting for payment confirmation", orderID)
	return ks.checkAndStartPreparing(ctx, aggregate)
}

func (ks *KitchenService) HandlePaymentConfirmed(ctx context.Context, orderID string) error {
	log.Printf("Processing PaymentConfirmed event for order %s", orderID)

	aggregate, err := ks.aggregateRepo.GetAggregate(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get aggregate: %w", err)
	}

	if aggregate == nil {
		aggregate = &domain.KitchenAggregate{
			OrderID:   orderID,
			CreatedAt: time.Now(),
		}
	}

	aggregate.PaymentConfirmed = true
	aggregate.UpdatedAt = time.Now()

	if err := ks.aggregateRepo.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	log.Printf("Payment confirmed for order %s", orderID)
	return ks.checkAndStartPreparing(ctx, aggregate)
}

func (ks *KitchenService) checkAndStartPreparing(ctx context.Context, aggregate *domain.KitchenAggregate) error {
	if !aggregate.CanStartPreparing() {
		log.Printf("Order %s not ready yet - order_received: %v, payment_confirmed: %v",
			aggregate.OrderID, aggregate.OrderReceived, aggregate.PaymentConfirmed)
		return nil
	}

	log.Printf("Both events received for order %s - starting preparation!", aggregate.OrderID)

	kitchenOrder := &domain.KitchenOrder{
		OrderID:       aggregate.OrderID,
		CustomerID:    aggregate.CustomerID,
		Items:         aggregate.Items,
		Status:        domain.StatusReceived,
		EstimatedTime: aggregate.EstimatedTime,
		CreatedAt:     aggregate.CreatedAt,
		UpdatedAt:     time.Now(),
	}

	return ks.ReceiveOrder(ctx, kitchenOrder)
}

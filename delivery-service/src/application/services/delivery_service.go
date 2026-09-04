package services

import (
	"context"
	"delivery-service/src/application/ports"
	"delivery-service/src/domain"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type DeliveryService struct {
	repo      ports.AggregateRepository
	publisher ports.EventPublisher
}

func NewDeliveryService(repo ports.AggregateRepository, publisher ports.EventPublisher) ports.DeliveryService {
	return &DeliveryService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *DeliveryService) HandleOrderCreated(ctx context.Context, event *domain.OrderCreatedEvent) error {
	log.Printf("Processing OrderCreated event for order %s, type: %s", event.OrderID, event.OrderType)

	aggregate, err := s.repo.GetAggregate(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get aggregate: %w", err)
	}

	if aggregate == nil {
		aggregate = &domain.DeliveryAggregate{
			OrderID:   event.OrderID,
			Status:    domain.StatusWaitingForKitchen,
			CreatedAt: time.Now(),
		}
	}

	aggregate.OrderReceived = true
	aggregate.UserID = event.UserID
	aggregate.OrderType = event.OrderType
	aggregate.TotalAmount = event.TotalAmount
	aggregate.Currency = event.Currency
	aggregate.DeliveryInfo = event.DeliveryInfo
	aggregate.UpdatedAt = time.Now()
	aggregate.UpdateStatus()

	if err := s.repo.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	log.Printf("Order %s marked as received, waiting for kitchen", event.OrderID)
	return s.checkAndTriggerActions(ctx, aggregate)
}

func (s *DeliveryService) HandleKitchenStatusChanged(ctx context.Context, event *domain.KitchenStatusChangedEvent) error {
	log.Printf("Processing KitchenStatusChanged event for order %s, status: %s", event.OrderID, event.Status)

	if event.Status != "ready" {
		log.Printf("Ignoring kitchen status %s for order %s", event.Status, event.OrderID)
		return nil
	}

	aggregate, err := s.repo.GetAggregate(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get aggregate: %w", err)
	}

	if aggregate == nil {
		aggregate = &domain.DeliveryAggregate{
			OrderID:   event.OrderID,
			Status:    domain.StatusWaitingForOrder,
			CreatedAt: time.Now(),
		}
	}

	aggregate.KitchenReady = true
	aggregate.UpdatedAt = time.Now()
	aggregate.UpdateStatus()

	if err := s.repo.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	log.Printf("Order %s marked as kitchen ready", event.OrderID)
	return s.checkAndTriggerActions(ctx, aggregate)
}

func (s *DeliveryService) checkAndTriggerActions(ctx context.Context, aggregate *domain.DeliveryAggregate) error {
	switch {
	case aggregate.CanStartDelivery():
		return s.startDelivery(ctx, aggregate)
	case aggregate.IsPickupReady():
		return s.notifyPickupReady(ctx, aggregate)
	default:
		log.Printf("Order %s not ready yet - waiting for more events", aggregate.OrderID)
		return nil
	}
}

func (s *DeliveryService) startDelivery(ctx context.Context, aggregate *domain.DeliveryAggregate) error {
	log.Printf("Starting delivery for order %s", aggregate.OrderID)

	aggregate.DeliveryStarted = true
	aggregate.Status = domain.StatusReadyToDeliver
	aggregate.UpdatedAt = time.Now()

	driverID := fmt.Sprintf("driver-%03d", 1+rand.Intn(5))
	aggregate.DriverID = &driverID

	if err := s.repo.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate after delivery start: %w", err)
	}

	// Use DeliveryInfo directly (may be nil for orders without delivery info)
	if aggregate.DeliveryInfo == nil {
		log.Printf("Warning: No delivery info for order %s", aggregate.OrderID)
	}

	event := &domain.DeliveryEvent{
		EventType:    "delivery_event",
		OrderID:      aggregate.OrderID,
		Status:       domain.DeliveryStatusAssigned,
		Message:      "Fahrer wurde zugewiesen",
		DriverID:     &driverID,
		DeliveryInfo: aggregate.DeliveryInfo,
		Timestamp:    time.Now(),
	}

	if err := s.publisher.PublishDeliveryEvent(ctx, event); err != nil {
		log.Printf("Failed to publish delivery event: %v", err)
		return err
	}

	log.Printf("Delivery assigned for order %s to driver %s", aggregate.OrderID, driverID)

	go s.simulateDriverUpdates(context.Background(), aggregate.OrderID)

	return nil
}

func (s *DeliveryService) notifyPickupReady(ctx context.Context, aggregate *domain.DeliveryAggregate) error {
	log.Printf("Order %s ready for pickup", aggregate.OrderID)

	aggregate.Status = domain.StatusPickupReady
	aggregate.UpdatedAt = time.Now()

	if err := s.repo.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate after pickup ready: %w", err)
	}

	event := &domain.DeliveryEvent{
		EventType: "delivery_event",
		OrderID:   aggregate.OrderID,
		Status:    domain.DeliveryStatusPickupReady,
		Message:   "Bestellung ist abholbereit",
		Timestamp: time.Now(),
	}

	if err := s.publisher.PublishDeliveryEvent(ctx, event); err != nil {
		log.Printf("Failed to publish pickup ready event: %v", err)
		return err
	}

	log.Printf("Pickup ready notification sent for order %s", aggregate.OrderID)
	return nil
}

func (s *DeliveryService) GetOrderStatus(ctx context.Context, orderID string) (*domain.DeliveryAggregate, error) {
	return s.repo.GetAggregate(ctx, orderID)
}

func (s *DeliveryService) UpdateDeliveryStatus(ctx context.Context, orderID string, status domain.DeliveryStatus, message string) error {
	log.Printf("Updating delivery status for order %s to %s", orderID, status)

	aggregate, err := s.repo.GetAggregate(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get aggregate: %w", err)
	}

	if aggregate == nil {
		return fmt.Errorf("order %s not found", orderID)
	}

	if aggregate.OrderType != "delivery" {
		return fmt.Errorf("order %s is not a delivery order", orderID)
	}

	// Map DeliveryStatus to AggregateStatus
	switch status {
	case domain.DeliveryStatusInTransit:
		aggregate.Status = domain.StatusInTransit
	case domain.DeliveryStatusDelivered:
		aggregate.Status = domain.StatusDelivered
	}
	aggregate.UpdatedAt = time.Now()

	if err := s.repo.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	event := &domain.DeliveryEvent{
		EventType: "delivery_event",
		OrderID:   orderID,
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
	}

	if err := s.publisher.PublishDeliveryEvent(ctx, event); err != nil {
		log.Printf("Failed to publish delivery status event: %v", err)
		return err
	}

	log.Printf("Delivery status updated for order %s: %s - %s", orderID, status, message)
	return nil
}

func (s *DeliveryService) simulateDriverUpdates(ctx context.Context, orderID string) {
	log.Printf("Starting driver simulation for order %s", orderID)

	// Wait before driver starts transit
	delay1 := time.Duration(10+rand.Intn(10)) * time.Second
	log.Printf("Driver will start transit for order %s in %v", orderID, delay1)
	time.Sleep(delay1)

	err := s.UpdateDeliveryStatus(ctx, orderID, domain.DeliveryStatusInTransit, "Fahrer ist unterwegs zum Kunden")
	if err != nil {
		log.Printf("Failed to update transit status for order %s: %v", orderID, err)
		return
	}

	// Wait before delivery
	delay2 := time.Duration(15+rand.Intn(15)) * time.Second
	log.Printf("Driver will deliver order %s in %v", orderID, delay2)
	time.Sleep(delay2)

	err = s.UpdateDeliveryStatus(ctx, orderID, domain.DeliveryStatusDelivered, "Bestellung wurde erfolgreich zugestellt")
	if err != nil {
		log.Printf("Failed to update delivered status for order %s: %v", orderID, err)
		return
	}

	log.Printf("Driver simulation completed for order %s", orderID)
}

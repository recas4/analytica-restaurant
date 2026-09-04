package domain

import "testing"

func TestAggregateStatus_Constants(t *testing.T) {
	tests := []struct {
		name   string
		status AggregateStatus
		want   string
	}{
		{"StatusWaitingForOrder", StatusWaitingForOrder, "waiting_for_order"},
		{"StatusWaitingForKitchen", StatusWaitingForKitchen, "waiting_for_kitchen"},
		{"StatusReadyToDeliver", StatusReadyToDeliver, "ready_to_deliver"},
		{"StatusDelivering", StatusDelivering, "delivering"},
		{"StatusInTransit", StatusInTransit, "in_transit"},
		{"StatusDelivered", StatusDelivered, "delivered"},
		{"StatusPickupReady", StatusPickupReady, "pickup_ready"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("Expected %s, got %s", tt.want, string(tt.status))
			}
		})
	}
}

func TestDeliveryStatus_Constants(t *testing.T) {
	tests := []struct {
		name   string
		status DeliveryStatus
		want   string
	}{
		{"DeliveryStatusInTransit", DeliveryStatusInTransit, "in_transit"},
		{"DeliveryStatusDelivered", DeliveryStatusDelivered, "delivered"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("Expected %s, got %s", tt.want, string(tt.status))
			}
		})
	}
}

func TestCanStartDelivery_True(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:         "order-123",
		OrderReceived:   true,
		KitchenReady:    true,
		DeliveryStarted: false,
		OrderType:       "delivery",
	}

	if !aggregate.CanStartDelivery() {
		t.Error("CanStartDelivery should return true when all conditions are met")
	}
}

func TestCanStartDelivery_False_NotReceived(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:         "order-123",
		OrderReceived:   false,
		KitchenReady:    true,
		DeliveryStarted: false,
		OrderType:       "delivery",
	}

	if aggregate.CanStartDelivery() {
		t.Error("CanStartDelivery should return false when order not received")
	}
}

func TestCanStartDelivery_False_KitchenNotReady(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:         "order-123",
		OrderReceived:   true,
		KitchenReady:    false,
		DeliveryStarted: false,
		OrderType:       "delivery",
	}

	if aggregate.CanStartDelivery() {
		t.Error("CanStartDelivery should return false when kitchen not ready")
	}
}

func TestCanStartDelivery_False_AlreadyStarted(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:         "order-123",
		OrderReceived:   true,
		KitchenReady:    true,
		DeliveryStarted: true,
		OrderType:       "delivery",
	}

	if aggregate.CanStartDelivery() {
		t.Error("CanStartDelivery should return false when delivery already started")
	}
}

func TestCanStartDelivery_False_PickupOrder(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:         "order-123",
		OrderReceived:   true,
		KitchenReady:    true,
		DeliveryStarted: false,
		OrderType:       "pickup",
	}

	if aggregate.CanStartDelivery() {
		t.Error("CanStartDelivery should return false for pickup orders")
	}
}

func TestIsPickupReady_True(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:       "order-123",
		OrderReceived: true,
		KitchenReady:  true,
		OrderType:     "pickup",
	}

	if !aggregate.IsPickupReady() {
		t.Error("IsPickupReady should return true when all conditions are met")
	}
}

func TestIsPickupReady_False_DeliveryOrder(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:       "order-123",
		OrderReceived: true,
		KitchenReady:  true,
		OrderType:     "delivery",
	}

	if aggregate.IsPickupReady() {
		t.Error("IsPickupReady should return false for delivery orders")
	}
}

func TestIsPickupReady_False_NotReady(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:       "order-123",
		OrderReceived: false,
		KitchenReady:  true,
		OrderType:     "pickup",
	}

	if aggregate.IsPickupReady() {
		t.Error("IsPickupReady should return false when order not received")
	}
}

func TestUpdateStatus_PickupReady(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:       "order-123",
		OrderReceived: true,
		KitchenReady:  true,
		OrderType:     "pickup",
	}

	aggregate.UpdateStatus()

	if aggregate.Status != StatusPickupReady {
		t.Errorf("Expected status 'pickup_ready', got '%s'", aggregate.Status)
	}
}

func TestUpdateStatus_ReadyToDeliver(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:         "order-123",
		OrderReceived:   true,
		KitchenReady:    true,
		DeliveryStarted: false,
		OrderType:       "delivery",
	}

	aggregate.UpdateStatus()

	if aggregate.Status != StatusReadyToDeliver {
		t.Errorf("Expected status 'ready_to_deliver', got '%s'", aggregate.Status)
	}
}

func TestUpdateStatus_Delivering(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:         "order-123",
		OrderReceived:   true,
		KitchenReady:    true,
		DeliveryStarted: true,
		OrderType:       "delivery",
	}

	aggregate.UpdateStatus()

	if aggregate.Status != StatusDelivering {
		t.Errorf("Expected status 'delivering', got '%s'", aggregate.Status)
	}
}

func TestUpdateStatus_WaitingForKitchen(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:         "order-123",
		OrderReceived:   true,
		KitchenReady:    false,
		DeliveryStarted: false,
		OrderType:       "delivery",
	}

	aggregate.UpdateStatus()

	if aggregate.Status != StatusWaitingForKitchen {
		t.Errorf("Expected status 'waiting_for_kitchen', got '%s'", aggregate.Status)
	}
}

func TestUpdateStatus_WaitingForOrder(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:         "order-123",
		OrderReceived:   false,
		KitchenReady:    true,
		DeliveryStarted: false,
		OrderType:       "delivery",
	}

	aggregate.UpdateStatus()

	if aggregate.Status != StatusWaitingForOrder {
		t.Errorf("Expected status 'waiting_for_order', got '%s'", aggregate.Status)
	}
}

func TestUpdateStatus_Default(t *testing.T) {
	aggregate := &DeliveryAggregate{
		OrderID:         "order-123",
		OrderReceived:   false,
		KitchenReady:    false,
		DeliveryStarted: false,
		OrderType:       "delivery",
	}

	aggregate.UpdateStatus()

	if aggregate.Status != StatusWaitingForOrder {
		t.Errorf("Expected status 'waiting_for_order', got '%s'", aggregate.Status)
	}
}

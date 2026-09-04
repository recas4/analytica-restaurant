package domain

import "testing"

func TestOrderStatus_Constants(t *testing.T) {
	tests := []struct {
		name   string
		status OrderStatus
		want   string
	}{
		{"StatusReceived", StatusReceived, "received"},
		{"StatusPreparing", StatusPreparing, "preparing"},
		{"StatusReady", StatusReady, "ready"},
		{"StatusPickedUpByDriver", StatusPickedUpByDriver, "picked_up_by_driver"},
		{"StatusCancelled", StatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("Expected %s, got %s", tt.want, string(tt.status))
			}
		})
	}
}

func TestKitchenOrder_Fields(t *testing.T) {
	order := KitchenOrder{
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		Status:        StatusReceived,
		EstimatedTime: 300,
	}

	if order.OrderID != "order-123" {
		t.Errorf("Expected OrderID 'order-123', got '%s'", order.OrderID)
	}

	if order.CustomerID != "customer-456" {
		t.Errorf("Expected CustomerID 'customer-456', got '%s'", order.CustomerID)
	}

	if order.Status != StatusReceived {
		t.Errorf("Expected Status 'received', got '%s'", order.Status)
	}

	if order.EstimatedTime != 300 {
		t.Errorf("Expected EstimatedTime 300, got %d", order.EstimatedTime)
	}
}

func TestOrderItem_Fields(t *testing.T) {
	item := OrderItem{
		ProductID:   "product-123",
		ProductName: "Pizza Margherita",
		Quantity:    2,
		Price:       12.99,
		PrepTime:    600,
	}

	if item.ProductID != "product-123" {
		t.Errorf("Expected ProductID 'product-123', got '%s'", item.ProductID)
	}

	if item.ProductName != "Pizza Margherita" {
		t.Errorf("Expected ProductName 'Pizza Margherita', got '%s'", item.ProductName)
	}

	if item.Quantity != 2 {
		t.Errorf("Expected Quantity 2, got %d", item.Quantity)
	}

	if item.Price != 12.99 {
		t.Errorf("Expected Price 12.99, got %f", item.Price)
	}

	if item.PrepTime != 600 {
		t.Errorf("Expected PrepTime 600, got %d", item.PrepTime)
	}
}

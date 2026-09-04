package domain

import "time"

// Incoming Events

type OrderCreatedEvent struct {
	OrderID         string        `json:"order_id"`
	UserID          string        `json:"user_id"`
	TotalAmount     float64       `json:"total_amount"`
	Currency        string        `json:"currency"`
	PaymentProvider string        `json:"payment_provider"`
	Items           []OrderItem   `json:"items"`
	CreatedAt       time.Time     `json:"created_at"`
	OrderType       string        `json:"order_type"`
	DeliveryInfo    *DeliveryInfo `json:"delivery_info,omitempty"`
}

type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	PrepTime    int     `json:"prep_time"`
}

type KitchenStatusChangedEvent struct {
	EventType     string    `json:"event_type"`
	OrderID       string    `json:"order_id"`
	Status        string    `json:"status"`
	EstimatedTime int       `json:"estimated_time,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// Outgoing Event (unified)

type DeliveryEvent struct {
	EventType    string         `json:"event_type"`
	OrderID      string         `json:"order_id"`
	Status       DeliveryStatus `json:"status"`
	Message      string         `json:"message,omitempty"`
	DriverID     *string        `json:"driver_id,omitempty"`
	DeliveryInfo *DeliveryInfo  `json:"delivery_info,omitempty"`
	Timestamp    time.Time      `json:"timestamp"`
}

type DeliveryStatus string

const (
	DeliveryStatusAssigned    DeliveryStatus = "assigned"
	DeliveryStatusInTransit   DeliveryStatus = "in_transit"
	DeliveryStatusDelivered   DeliveryStatus = "delivered"
	DeliveryStatusPickupReady DeliveryStatus = "pickup_ready"
)

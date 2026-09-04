package models

import (
	"time"
)

// Order repraesentiert eine Bestellung im System.
// Enthaelt alle relevanten Informationen fuer die Zahlungsabwicklung.
type Order struct {
	ID              string        `json:"id"`
	UserID          string        `json:"user_id"`
	TotalAmount     float64       `json:"total_amount"`
	Currency        string        `json:"currency"`
	PaymentProvider string        `json:"payment_provider"`
	Items           []string      `json:"items"`
	Status          string        `json:"status"`
	OrderType       string        `json:"order_type"`
	DeliveryInfo    *DeliveryInfo `json:"delivery_info,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
}

// OrderCreatedEvent repraesentiert das Event das bei Bestellerstellung veroeffentlicht wird.
// Wird ueber Kafka an andere Services gesendet.
type OrderCreatedEvent struct {
	OrderID         string        `json:"order_id"`
	UserID          string        `json:"user_id"`
	TotalAmount     float64       `json:"total_amount"`
	Currency        string        `json:"currency"`
	PaymentProvider string        `json:"payment_provider"`
	Items           []string      `json:"items"`
	OrderType       string        `json:"order_type"`
	DeliveryInfo    *DeliveryInfo `json:"delivery_info,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
}

// NewOrder erstellt eine neue Bestellung mit den angegebenen Parametern.
// Nutzt die uebergebene Order-ID statt eine neue zu generieren.
func NewOrder(orderID, userID string, items []string, totalAmount float64, paymentProvider, orderType string, deliveryInfo *DeliveryInfo) *Order {
	return &Order{
		ID:              orderID,
		UserID:          userID,
		TotalAmount:     totalAmount,
		Currency:        "EUR",
		PaymentProvider: paymentProvider,
		Items:           items,
		Status:          "created",
		OrderType:       orderType,
		DeliveryInfo:    deliveryInfo,
		CreatedAt:       time.Now(),
	}
}

// ToEvent konvertiert die Bestellung in ein Event fuer die Nachrichtenverarbeitung.
func (o *Order) ToEvent() *OrderCreatedEvent {
	return &OrderCreatedEvent{
		OrderID:         o.ID,
		UserID:          o.UserID,
		TotalAmount:     o.TotalAmount,
		Currency:        o.Currency,
		PaymentProvider: o.PaymentProvider,
		Items:           o.Items,
		OrderType:       o.OrderType,
		DeliveryInfo:    o.DeliveryInfo,
		CreatedAt:       o.CreatedAt,
	}
}

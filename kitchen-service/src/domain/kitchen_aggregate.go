package domain

import "time"

type KitchenAggregate struct {
	OrderID          string      `json:"order_id"`
	OrderReceived    bool        `json:"order_received"`
	PaymentConfirmed bool        `json:"payment_confirmed"`
	CustomerID       string      `json:"customer_id"`
	Items            []OrderItem `json:"items"`
	EstimatedTime    int         `json:"estimated_time"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

func (a *KitchenAggregate) CanStartPreparing() bool {
	return a.OrderReceived && a.PaymentConfirmed
}

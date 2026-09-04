package models

import (
	"fmt"

	"github.com/google/uuid"
)

// OrderItem repraesentiert einen einzelnen Artikel in einer Bestellung.
type OrderItem struct {
	ProductID  string  `json:"product_id"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
}

// DeliveryInfo enthaelt die Lieferadresse.
type DeliveryInfo struct {
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
	Street        string `json:"street"`
	HouseNumber   string `json:"house_number"`
	PostalCode    string `json:"postal_code"`
	City          string `json:"city"`
	Floor         string `json:"floor,omitempty"`
	Instructions  string `json:"instructions,omitempty"`
}

// CheckoutRequest repraesentiert eine eingehende Checkout-Anfrage.
// Wird von Kafka konsumiert und in eine Bestellung umgewandelt.
type CheckoutRequest struct {
	EventType    string        `json:"event_type"`
	OrderID      string        `json:"order_id"`
	UserID       string        `json:"user_id"`
	Items        []OrderItem   `json:"items"`
	ProductIDs   []string      `json:"product_ids"`
	TotalAmount  float64       `json:"total_amount"`
	Currency     string        `json:"currency"`
	Status       string        `json:"status"`
	Timestamp    string        `json:"timestamp"`
	OrderType    string        `json:"order_type"`
	DeliveryInfo *DeliveryInfo `json:"delivery_info,omitempty"`
}

// Validate prueft ob die Checkout-Anfrage alle erforderlichen Felder enthaelt.
// Gibt einen Fehler zurueck wenn die Validierung fehlschlaegt.
func (c *CheckoutRequest) Validate() error {
	if len(c.Items) == 0 {
		return fmt.Errorf("Artikel duerfen nicht leer sein")
	}
	if c.TotalAmount <= 0 {
		return fmt.Errorf("Gesamtbetrag muss positiv sein")
	}
	if c.UserID == "" {
		return fmt.Errorf("Benutzer-ID ist erforderlich")
	}
	return nil
}

// ToOrder konvertiert die Checkout-Anfrage in eine Bestellung.
// Setzt Standardwerte fuer optionale Felder.
func (c *CheckoutRequest) ToOrder() *Order {
	userID := c.UserID
	if userID == "" {
		userID = uuid.New().String()
	}

	orderID := c.OrderID
	if orderID == "" {
		orderID = uuid.New().String()
	}

	// Artikel-IDs aus den Bestellpositionen extrahieren
	productIDs := make([]string, len(c.Items))
	for i, item := range c.Items {
		productIDs[i] = item.ProductID
	}

	// Standard-Zahlungsanbieter
	paymentProvider := "stripe"

	// Order-Type (default: delivery)
	orderType := c.OrderType
	if orderType == "" {
		orderType = "delivery"
	}

	return NewOrder(orderID, userID, productIDs, c.TotalAmount, paymentProvider, orderType, c.DeliveryInfo)
}

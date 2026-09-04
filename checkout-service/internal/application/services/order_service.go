package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"checkout-service/internal/application/ports"
	"checkout-service/internal/domain/models"
)

// OrderServiceImpl implementiert das OrderService-Interface.
// Koordiniert die Checkout-Verarbeitung und Event-Veroeffentlichung.
type OrderServiceImpl struct {
	eventPublisher ports.EventPublisher
}

// NewOrderService erstellt eine neue OrderService-Instanz.
// Benoetigt einen EventPublisher fuer die Nachrichtenveroeffentlichung.
func NewOrderService(publisher ports.EventPublisher) ports.OrderService {
	return &OrderServiceImpl{
		eventPublisher: publisher,
	}
}

// ProcessCheckout verarbeitet eine eingehende Checkout-Anfrage.
// Validiert die Daten, erstellt eine Bestellung und veroeffentlicht das Event.
func (s *OrderServiceImpl) ProcessCheckout(ctx context.Context, checkoutData []byte) error {
	log.Printf("Verarbeite Checkout-Anfrage: %s", string(checkoutData))

	var checkoutReq models.CheckoutRequest
	if err := json.Unmarshal(checkoutData, &checkoutReq); err != nil {
		log.Printf("Checkout-Daten konnten nicht geparst werden: %v", err)
		return fmt.Errorf("Ungueltige Checkout-Daten: %w", err)
	}

	if err := checkoutReq.Validate(); err != nil {
		log.Printf("Validierung fehlgeschlagen: %v", err)
		return err
	}

	order := checkoutReq.ToOrder()
	event := order.ToEvent()

	if err := s.eventPublisher.PublishOrderCreated(ctx, event); err != nil {
		log.Printf("Event-Veroeffentlichung fehlgeschlagen: %v", err)
		return err
	}

	log.Printf("Bestellung %s erfolgreich verarbeitet - Artikel: %v, Gesamt: %.2f EUR",
		order.ID, order.Items, order.TotalAmount)
	return nil
}

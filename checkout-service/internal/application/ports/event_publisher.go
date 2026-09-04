package ports

import (
	"checkout-service/internal/domain/models"
	"context"
)

// EventPublisher definiert das Interface fuer die Veroeffentlichung von Events.
// Abstrahiert den Messaging-Layer von der Geschaeftslogik.
type EventPublisher interface {
	// PublishOrderCreated veroeffentlicht ein Bestellerstellungs-Event.
	// Gibt einen Fehler zurueck wenn die Veroeffentlichung fehlschlaegt.
	PublishOrderCreated(ctx context.Context, event *models.OrderCreatedEvent) error
}

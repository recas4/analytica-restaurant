package ports

import "context"

// EventPublisher definiert die Schnittstelle fuer Event-Veroeffentlichung.
// Wird von der Kafka-Implementierung umgesetzt.
type EventPublisher interface {
	// PublishMessage sendet eine Nachricht an das Message-System.
	PublishMessage(ctx context.Context, message interface{}) error

	// Close schliesst die Verbindung zum Message-System.
	Close() error
}

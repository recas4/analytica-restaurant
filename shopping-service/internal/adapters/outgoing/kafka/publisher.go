package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"shopping-service/internal/application/ports"
)

// EventPublisher implementiert das EventPublisher-Interface mit Kafka.
type EventPublisher struct {
	writer *kafka.Writer
}

// NewEventPublisher erstellt einen neuen Kafka Event-Publisher.
func NewEventPublisher(broker, topic string) ports.EventPublisher {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})

	log.Printf("Kafka Publisher initialisiert - Broker: %s, Topic: %s", broker, topic)
	return &EventPublisher{writer: writer}
}

// PublishMessage sendet eine Nachricht an Kafka.
func (p *EventPublisher) PublishMessage(ctx context.Context, message interface{}) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Fehler beim Serialisieren der Nachricht: %v", err)
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("shopping-event"),
		Value: msgBytes,
	})

	if err != nil {
		log.Printf("Kafka Sendefehler: %v", err)
		return err
	}

	log.Printf("Nachricht an Kafka gesendet: %s", string(msgBytes))
	return nil
}

// Close schliesst die Kafka-Verbindung.
func (p *EventPublisher) Close() error {
	return p.writer.Close()
}

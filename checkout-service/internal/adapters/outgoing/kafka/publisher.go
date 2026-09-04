package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"checkout-service/internal/application/ports"
	"checkout-service/internal/domain/models"

	"github.com/segmentio/kafka-go"
)

// KafkaEventPublisher implementiert EventPublisher mit Kafka als Message-Broker.
// Veroeffentlicht Bestellungs-Events an das checkout-events Topic.
type KafkaEventPublisher struct {
	writer *kafka.Writer
}

// NewEventPublisher erstellt einen neuen Kafka-Event-Publisher.
// Konfiguriert den Writer fuer das checkout-events Topic.
func NewEventPublisher() ports.EventPublisher {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	
	return &KafkaEventPublisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(kafkaBrokers),
			Topic:    "checkout-events",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

// PublishOrderCreated veroeffentlicht ein OrderCreatedEvent an Kafka.
// Serialisiert das Event und sendet es mit der Bestell-ID als Schluessel.
func (p *KafkaEventPublisher) PublishOrderCreated(ctx context.Context, event *models.OrderCreatedEvent) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Key:   []byte(event.OrderID),
		Value: eventData,
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		log.Printf("OrderCreated-Event konnte nicht veroeffentlicht werden: %v", err)
		return err
	}

	log.Printf("OrderCreated-Event fuer Bestellung %s veroeffentlicht", event.OrderID)
	return nil
}

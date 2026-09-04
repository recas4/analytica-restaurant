package outgoing

import (
	"context"
	"encoding/json"
	"log"
	"payment-service/src/application/ports"
	"payment-service/src/domain"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher() ports.EventPublisher {
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP("kafka:9092"),
			Topic:    "payment-events",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaPublisher) PublishPaymentSucceeded(ctx context.Context, event *domain.PaymentSucceededEvent) error {
	return p.publishEvent(ctx, "PaymentSucceeded", event.OrderID, event)
}

func (p *KafkaPublisher) PublishPaymentFailed(ctx context.Context, event *domain.PaymentFailedEvent) error {
	return p.publishEvent(ctx, "PaymentFailed", event.OrderID, event)
}

func (p *KafkaPublisher) publishEvent(ctx context.Context, eventType, orderID string, event interface{}) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Key:   []byte(orderID),
		Value: eventData,
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		log.Printf("Failed to publish %s event: %v", eventType, err)
		return err
	}

	log.Printf("Published %s event for order %s", eventType, orderID)
	return nil
}

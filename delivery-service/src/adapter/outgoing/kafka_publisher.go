package outgoing

import (
	"context"
	"delivery-service/src/application/ports"
	"delivery-service/src/domain"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string) ports.EventPublisher {
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    "delivery-events",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaPublisher) PublishDeliveryEvent(ctx context.Context, event *domain.DeliveryEvent) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Key:   []byte(event.OrderID),
		Value: eventData,
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		log.Printf("Failed to publish event: %v", err)
		return err
	}

	log.Printf("Published delivery event: %s", string(eventData))
	return nil
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}

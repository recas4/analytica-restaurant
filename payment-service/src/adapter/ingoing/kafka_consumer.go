package ingoing

import (
	"context"
	"encoding/json"
	"log"
	"payment-service/src/application/ports"
	"payment-service/src/domain"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer() *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{"kafka:9092"},
			Topic:   "checkout-events",
			GroupID: "payment-service-group",
		}),
	}
}

func (c *KafkaConsumer) StartConsuming(ctx context.Context, paymentService ports.PaymentService) error {
	log.Println("Starting to consume from 'checkout-events' topic")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping order event consumer")
			return c.reader.Close()
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Failed to read order event: %v", err)
				continue
			}

			var orderEvent domain.OrderCreatedEvent
			if err := json.Unmarshal(message.Value, &orderEvent); err != nil {
				log.Printf("Failed to unmarshal order event: %v", err)
				continue
			}

			payment, err := paymentService.ProcessOrderPayment(ctx, orderEvent)
			if err != nil {
				log.Printf("Failed to process payment for order %s: %v", orderEvent.OrderID, err)
				continue
			}

			log.Printf("Payment %s processed for order %s", payment.ID.String(), orderEvent.OrderID)
		}
	}
}

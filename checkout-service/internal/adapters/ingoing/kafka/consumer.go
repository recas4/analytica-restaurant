package kafka

import (
	"context"
	"log"
	"os"

	"checkout-service/internal/application/ports"

	"github.com/segmentio/kafka-go"
)

// CheckoutConsumer konsumiert Checkout-Nachrichten von Kafka.
// Implementiert den eingehenden Adapter fuer Message-Consuming.
type CheckoutConsumer struct {
	reader *kafka.Reader
}

// NewCheckoutConsumer erstellt einen neuen Kafka-Consumer.
// Konfiguriert die Verbindung zum Kafka-Broker und das Topic.
func NewCheckoutConsumer() *CheckoutConsumer {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	
	return &CheckoutConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{kafkaBrokers},
			Topic:   "checkout",
			GroupID: "checkout-group",
		}),
	}
}

// StartConsuming startet die Nachrichtenverarbeitung.
// Laeuft bis der Context abgebrochen wird.
func (c *CheckoutConsumer) StartConsuming(ctx context.Context, orderService ports.OrderService) error {
	log.Println("Starte Konsumierung vom 'checkout' Topic")

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer wird gestoppt")
			return c.reader.Close()
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Nachricht konnte nicht gelesen werden: %v", err)
				continue
			}

			if err := orderService.ProcessCheckout(ctx, message.Value); err != nil {
				log.Printf("Checkout-Verarbeitung fehlgeschlagen: %v", err)
				continue
			}
		}
	}
}

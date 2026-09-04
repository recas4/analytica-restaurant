package ingoing

import (
	"context"
	"delivery-service/src/application/ports"
	"delivery-service/src/domain"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	checkoutReader  *kafka.Reader
	kitchenReader   *kafka.Reader
	deliveryService ports.DeliveryService
}

func NewKafkaConsumer(brokers []string, deliveryService ports.DeliveryService) *KafkaConsumer {
	return &KafkaConsumer{
		checkoutReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			Topic:       "checkout-events",
			GroupID:     "delivery-service-checkout",
			StartOffset: kafka.LastOffset,
		}),
		kitchenReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			Topic:       "kitchen-events",
			GroupID:     "delivery-service-kitchen",
			StartOffset: kafka.LastOffset,
		}),
		deliveryService: deliveryService,
	}
}

func (c *KafkaConsumer) StartConsuming(ctx context.Context) error {
	go func() {
		if err := c.consumeCheckoutEvents(ctx); err != nil {
			log.Printf("Checkout events consumer error: %v", err)
		}
	}()

	go func() {
		if err := c.consumeKitchenEvents(ctx); err != nil {
			log.Printf("Kitchen events consumer error: %v", err)
		}
	}()

	return nil
}

func (c *KafkaConsumer) consumeCheckoutEvents(ctx context.Context) error {
	log.Println("Starting to consume checkout events...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.checkoutReader.FetchMessage(ctx)
			if err != nil {
				log.Printf("Failed to fetch checkout message: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if err := c.handleCheckoutMessage(ctx, msg); err != nil {
				log.Printf("Failed to handle checkout message: %v", err)
			}

			if err := c.checkoutReader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Failed to commit checkout message: %v", err)
			}
		}
	}
}

func (c *KafkaConsumer) consumeKitchenEvents(ctx context.Context) error {
	log.Println("Starting to consume kitchen events...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.kitchenReader.FetchMessage(ctx)
			if err != nil {
				log.Printf("Failed to fetch kitchen message: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if err := c.handleKitchenMessage(ctx, msg); err != nil {
				log.Printf("Failed to handle kitchen message: %v", err)
			}

			if err := c.kitchenReader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Failed to commit kitchen message: %v", err)
			}
		}
	}
}

func (c *KafkaConsumer) handleCheckoutMessage(ctx context.Context, msg kafka.Message) error {
	log.Printf("Received checkout event: %s", string(msg.Value))

	// Parse as raw JSON first to handle flexible items format
	var rawEvent map[string]interface{}
	if err := json.Unmarshal(msg.Value, &rawEvent); err != nil {
		return err
	}

	// Build the event with flexible items parsing
	event := domain.OrderCreatedEvent{
		OrderID:         getString(rawEvent, "order_id"),
		UserID:          getString(rawEvent, "user_id"),
		TotalAmount:     getFloat(rawEvent, "total_amount"),
		Currency:        getString(rawEvent, "currency"),
		PaymentProvider: getString(rawEvent, "payment_provider"),
		OrderType:       getString(rawEvent, "order_type"),
	}

	// Parse delivery_info if present
	if rawDeliveryInfo, ok := rawEvent["delivery_info"].(map[string]interface{}); ok {
		event.DeliveryInfo = &domain.DeliveryInfo{
			CustomerName:  getString(rawDeliveryInfo, "customer_name"),
			CustomerPhone: getString(rawDeliveryInfo, "customer_phone"),
			Street:        getString(rawDeliveryInfo, "street"),
			HouseNumber:   getString(rawDeliveryInfo, "house_number"),
			PostalCode:    getString(rawDeliveryInfo, "postal_code"),
			City:          getString(rawDeliveryInfo, "city"),
			Floor:         getString(rawDeliveryInfo, "floor"),
			Instructions:  getString(rawDeliveryInfo, "instructions"),
		}
	}

	// Handle items - can be []string or []OrderItem
	if rawItems, ok := rawEvent["items"].([]interface{}); ok {
		for _, item := range rawItems {
			switch v := item.(type) {
			case string:
				event.Items = append(event.Items, domain.OrderItem{
					ProductID:   v,
					ProductName: "Product",
					Quantity:    1,
				})
			case map[string]interface{}:
				event.Items = append(event.Items, domain.OrderItem{
					ProductID:   getString(v, "product_id"),
					ProductName: getString(v, "product_name"),
					Quantity:    getInt(v, "quantity"),
					Price:       getFloat(v, "price"),
				})
			}
		}
	}

	// For now, treat all orders as delivery if no order_type specified
	if event.OrderType == "" {
		event.OrderType = "delivery"
	}

	log.Printf("Processing order %s (type: %s) with %d items", event.OrderID, event.OrderType, len(event.Items))
	return c.deliveryService.HandleOrderCreated(ctx, &event)
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

func (c *KafkaConsumer) handleKitchenMessage(ctx context.Context, msg kafka.Message) error {
	log.Printf("Received kitchen event: %s", string(msg.Value))

	var statusEvent domain.KitchenStatusChangedEvent
	if err := json.Unmarshal(msg.Value, &statusEvent); err != nil {
		log.Printf("Failed to parse as status event, trying other formats: %v", err)
		return nil
	}

	return c.deliveryService.HandleKitchenStatusChanged(ctx, &statusEvent)
}

func (c *KafkaConsumer) Close() error {
	if err := c.checkoutReader.Close(); err != nil {
		log.Printf("Error closing checkout reader: %v", err)
	}
	if err := c.kitchenReader.Close(); err != nil {
		log.Printf("Error closing kitchen reader: %v", err)
	}
	return nil
}

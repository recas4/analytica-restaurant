package ingoing

import (
	"context"
	"encoding/json"
	"fmt"
	"kitchen-service/src/application/ports"
	"kitchen-service/src/domain"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	consumer       sarama.ConsumerGroup
	kitchenService ports.KitchenService
	topics         []string
}

func NewKafkaConsumer(brokers []string, groupID string, topics []string, kitchenService ports.KitchenService) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &KafkaConsumer{
		consumer:       consumer,
		kitchenService: kitchenService,
		topics:         topics,
	}, nil
}

func (kc *KafkaConsumer) Start(ctx context.Context) error {
	handler := &ConsumerGroupHandler{
		kitchenService: kc.kitchenService,
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Kitchen consumer context cancelled")
			return ctx.Err()
		default:
			if err := kc.consumer.Consume(ctx, kc.topics, handler); err != nil {
				log.Printf("Error consuming messages: %v", err)
				time.Sleep(time.Second * 5)
			}
		}
	}
}

func (kc *KafkaConsumer) Close() error {
	return kc.consumer.Close()
}

type ConsumerGroupHandler struct {
	kitchenService ports.KitchenService
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Println("Kitchen consumer group setup")
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Kitchen consumer group cleanup")
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			log.Printf("Received message from topic %s: %s", message.Topic, string(message.Value))

			if err := h.processMessage(message); err != nil {
				log.Printf("Error processing message: %v", err)
			} else {
				session.MarkMessage(message, "")
			}

		case <-session.Context().Done():
			return nil
		}
	}
}

func (h *ConsumerGroupHandler) processMessage(message *sarama.ConsumerMessage) error {
	ctx := context.Background()

	switch message.Topic {
	case "checkout-events":
		return h.handleCheckoutEvent(ctx, message.Value)
	case "payment-events":
		return h.handlePaymentEvent(ctx, message.Value)
	case "kitchen-events":
		return h.handleKitchenEvent(ctx, message.Value)
	default:
		log.Printf("Unknown topic: %s", message.Topic)
		return nil
	}
}

func (h *ConsumerGroupHandler) handleCheckoutEvent(ctx context.Context, data []byte) error {
	// First try to parse as raw JSON to handle flexible items format
	var rawEvent map[string]interface{}
	if err := json.Unmarshal(data, &rawEvent); err != nil {
		log.Printf("Failed to parse checkout event as raw JSON: %v", err)
		return err
	}

	orderID, _ := rawEvent["order_id"].(string)
	userID, _ := rawEvent["user_id"].(string)

	// Handle items - can be []string or []OrderItem
	var items []domain.OrderItem
	if rawItems, ok := rawEvent["items"].([]interface{}); ok {
		for _, item := range rawItems {
			switch v := item.(type) {
			case string:
				// Items as string array (product IDs only)
				items = append(items, domain.OrderItem{
					ProductID:   v,
					ProductName: "Product",
					Quantity:    1,
					PrepTime:    30,
				})
			case map[string]interface{}:
				// Items as object array
				orderItem := domain.OrderItem{
					ProductID:   getString(v, "product_id"),
					ProductName: getString(v, "product_name"),
					Quantity:    getInt(v, "quantity"),
					Price:       getFloat(v, "price"),
					PrepTime:    getInt(v, "prep_time"),
				}
				if orderItem.Quantity == 0 {
					orderItem.Quantity = 1
				}
				if orderItem.PrepTime == 0 {
					orderItem.PrepTime = 30
				}
				items = append(items, orderItem)
			}
		}
	}

	log.Printf("Processing checkout event for order %s with %d items", orderID, len(items))

	estimatedTime := calculateEstimatedTime(items)

	// Use event aggregation - wait for payment confirmation
	if err := h.kitchenService.HandleOrderCreated(ctx, orderID, userID, items, estimatedTime); err != nil {
		log.Printf("Failed to handle order created: %v", err)
		return err
	}

	log.Printf("Order %s registered, waiting for payment", orderID)
	return nil
}

// Helper functions for flexible JSON parsing
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

func calculateEstimatedTime(items []domain.OrderItem) int {
	totalTime := 0
	for _, item := range items {
		if item.PrepTime > 0 {
			totalTime += item.PrepTime * item.Quantity
		} else {
			totalTime += 30 * item.Quantity // default 30 seconds per item
		}
	}

	if totalTime < 10 {
		totalTime = 10 // minimum 10 seconds
	}
	if totalTime > 300 {
		totalTime = 300 // maximum 5 minutes
	}

	return totalTime
}

func (h *ConsumerGroupHandler) handlePaymentEvent(ctx context.Context, data []byte) error {
	var event struct {
		EventType string `json:"event_type"`
		OrderID   string `json:"order_id"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Failed to parse payment event: %v", err)
		return err
	}

	// Only handle payment_succeeded events
	if event.EventType != "payment_succeeded" {
		log.Printf("Ignoring payment event type: %s", event.EventType)
		return nil
	}

	log.Printf("Processing payment_succeeded event for order %s", event.OrderID)

	if err := h.kitchenService.HandlePaymentConfirmed(ctx, event.OrderID); err != nil {
		log.Printf("Failed to handle payment confirmed: %v", err)
		return err
	}

	log.Printf("Payment confirmed for order %s", event.OrderID)
	return nil
}

func (h *ConsumerGroupHandler) handleKitchenEvent(ctx context.Context, data []byte) error {
	// Parse event to determine type
	var baseEvent struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(data, &baseEvent); err != nil {
		log.Printf("Failed to parse event type: %v", err)
		return err
	}

	// Only handle order confirmations - ignore our own status events!
	switch baseEvent.EventType {
	case "order_confirmed":
		return h.handleOrderConfirmed(ctx, data)
	case "order_received_in_kitchen", "order_preparation_started", "order_ready",
		"order_picked_up_by_driver", "order_cancelled_in_kitchen", "kitchen_notification":
		// Ignore our own status events to prevent infinite loop
		log.Printf("Ignoring own event type: %s", baseEvent.EventType)
		return nil
	default:
		log.Printf("Unknown kitchen event type: %s", baseEvent.EventType)
		return nil
	}
}

func (h *ConsumerGroupHandler) handleOrderConfirmed(ctx context.Context, data []byte) error {
	var orderEvent domain.OrderReceivedEvent
	if err := json.Unmarshal(data, &orderEvent); err != nil {
		return fmt.Errorf("failed to unmarshal order event: %w", err)
	}

	// Use user_id as customer_id if customer_id is not set
	customerID := orderEvent.CustomerID
	if customerID == "" {
		customerID = orderEvent.UserID
	}

	kitchenOrder := &domain.KitchenOrder{
		OrderID:       orderEvent.OrderID,
		CustomerID:    customerID,
		Items:         orderEvent.Items,
		Status:        domain.StatusReceived,
		EstimatedTime: calculateEstimatedTime(orderEvent.Items),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.kitchenService.ReceiveOrder(ctx, kitchenOrder); err != nil {
		return fmt.Errorf("failed to receive order in kitchen: %w", err)
	}

	log.Printf("Order %s received in kitchen", orderEvent.OrderID)
	return nil
}

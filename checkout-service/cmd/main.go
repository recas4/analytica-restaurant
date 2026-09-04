package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	ingoingKafka "checkout-service/internal/adapters/ingoing/kafka"
	outgoingKafka "checkout-service/internal/adapters/outgoing/kafka"
	"checkout-service/internal/application/services"
)

func main() {
	log.Println("Checkout-Service wird gestartet")

	// Adapter initialisieren (ausgehend)
	publisher := outgoingKafka.NewEventPublisher()

	// Adapter initialisieren (eingehend)
	consumer := ingoingKafka.NewCheckoutConsumer()

	// Anwendungs-Service initialisieren
	orderService := services.NewOrderService(publisher)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Consumer in separater Goroutine starten
	go func() {
		if err := consumer.StartConsuming(ctx, orderService); err != nil {
			log.Fatal("Consumer fehlgeschlagen:", err)
		}
	}()

	// Graceful Shutdown einrichten
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Checkout-Service wird heruntergefahren")
	cancel()
}

package main

import (
	"context"
	"delivery-service/src/adapter/ingoing"
	"delivery-service/src/adapter/outgoing"
	"delivery-service/src/application/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	redisClient "github.com/redis/go-redis/v9"
)

func main() {
	log.Println("Starting Delivery Service with Event Aggregation")

	redisAddr := getEnv("REDIS_ADDR", "redis:6379")
	rdb := redisClient.NewClient(&redisClient.Options{
		Addr: redisAddr,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	aggregateRepo := outgoing.NewRedisRepository(rdb)
	eventPublisher := outgoing.NewKafkaPublisher([]string{"kafka:9092"})
	deliveryService := services.NewDeliveryService(aggregateRepo, eventPublisher)

	eventConsumer := ingoing.NewKafkaConsumer([]string{"kafka:9092"}, deliveryService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := eventConsumer.StartConsuming(ctx); err != nil {
		log.Fatalf("Failed to start consuming events: %v", err)
	}
	log.Println("Started consuming events from checkout-events and kitchen-events")

	// HTTP Server
	router := mux.NewRouter()
	httpHandler := ingoing.NewHTTPHandler(deliveryService)
	httpHandler.RegisterRoutes(router)

	port := getEnv("PORT", "8085")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("HTTP Server listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Delivery Service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)

	cancel()
	if err := eventConsumer.Close(); err != nil {
		log.Printf("Error closing event consumer: %v", err)
	}
	if err := eventPublisher.Close(); err != nil {
		log.Printf("Error closing event publisher: %v", err)
	}
	if err := rdb.Close(); err != nil {
		log.Printf("Error closing Redis client: %v", err)
	}

	log.Println("Delivery Service stopped gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

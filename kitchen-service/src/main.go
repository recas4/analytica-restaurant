package main

import (
	"context"
	"kitchen-service/src/adapter/ingoing"
	"kitchen-service/src/adapter/outgoing"
	"kitchen-service/src/application/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	log.Println("Starting Kitchen ...")

	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	redisAddr := getEnv("REDIS_ADDR", "redis:6379")
	kafkaBrokers := []string{getEnv("KAFKA_BROKER", "localhost:9092")}
	port := getEnv("PORT", "8084")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// MongoDB connection
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer mongoClient.Disconnect(ctx)

	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}
	log.Println("Connected to MongoDB")

	// Redis connection for event aggregation
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()
	log.Println("Connected to Redis")

	database := mongoClient.Database("kitchen_db")
	kitchenRepo := outgoing.NewMongoRepository(database)
	aggregateRepo := outgoing.NewRedisAggregateRepository(redisClient)

	publisher, err := outgoing.NewKafkaPublisher(kafkaBrokers)
	if err != nil {
		log.Fatal("Failed to create Kafka publisher:", err)
	}
	log.Println("Kafka Publisher created")

	kitchenService := services.NewKitchenService(kitchenRepo, aggregateRepo, publisher)

	consumer, err := ingoing.NewKafkaConsumer(
		kafkaBrokers,
		"kitchen-service-group",
		[]string{"checkout-events", "payment-events", "kitchen-events"},
		kitchenService,
	)
	if err != nil {
		log.Fatal("Failed to create Kafka consumer:", err)
	}
	defer consumer.Close()
	log.Println("Kafka Consumer created")

	go func() {
		log.Println("🎧 Starting Kafka consumer...")
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Kafka consumer error: %v", err)
		}
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := kitchenService.ProcessOrderQueue(ctx); err != nil {
					log.Printf("Error processing order queue: %v", err)
				}
			}
		}
	}()

	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Kitchen Service is healthy!"))
	}).Methods("GET")

	kitchenHandler := ingoing.NewHTTPHandler(kitchenService)
	kitchenHandler.RegisterRoutes(router)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("Kitchen Service running on port %s", port)
		log.Printf("Dashboard available at: http://localhost:%s/kitchen/dashboard", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down Kitchen Service...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Kitchen Service stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

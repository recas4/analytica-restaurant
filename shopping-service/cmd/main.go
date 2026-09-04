package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"shopping-service/internal/adapters/ingoing/rest"
	"shopping-service/internal/adapters/outgoing/kafka"
	"shopping-service/internal/adapters/outgoing/mongodb"
	"shopping-service/internal/application/services"
	"shopping-service/internal/config"
)

func main() {
	// Konfiguration laden
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// MongoDB Verbindung herstellen
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal("MongoDB Verbindung fehlgeschlagen:", err)
	}

	// Verbindung pruefen
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB Ping fehlgeschlagen:", err)
	}

	log.Println("MongoDB Verbindung erfolgreich")
	db := client.Database("shopping")

	// Repositories erstellen (Outgoing Adapters)
	productRepo := mongodb.NewProductRepository(db)
	cartRepo := mongodb.NewCartRepository(db)

	// Services erstellen (Application Layer)
	productService := services.NewProductService(productRepo)
	cartService := services.NewCartService(cartRepo)

	// Kafka Publisher erstellen (Outgoing Adapter)
	eventPublisher := kafka.NewEventPublisher(cfg.KafkaBkr, cfg.KafkaTopic)

	// Handler erstellen (Ingoing Adapters)
	productHandler := rest.NewProductHandler(productService, eventPublisher)
	cartHandler := rest.NewCartHandler(cartService, productService, eventPublisher)

	// HTTP-Server starten
	r := gin.Default()
	rest.SetupRoutes(r, productHandler, cartHandler)

	log.Printf("Shopping-Service laeuft auf Port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}

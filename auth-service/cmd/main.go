package main

import (
	"auth-service/internal/adapters/ingoing/rest"
	"auth-service/internal/adapters/outgoing/bcrypt"
	"auth-service/internal/adapters/outgoing/jwt"
	"auth-service/internal/adapters/outgoing/mongodb"
	"auth-service/internal/application/services"
	"auth-service/internal/config"
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Konfiguration aus Umgebungsvariablen laden
	cfg := config.Load()

	// MongoDB-Verbindung initialisieren
	client, err := mongo.NewClient(options.Client().ApplyURI(cfg.MongoDB.URI))
	if err != nil {
		log.Fatalf("MongoDB-Client konnte nicht erstellt werden: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Verbindung zu MongoDB fehlgeschlagen: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Fehler beim Trennen von MongoDB: %v", err)
		}
	}()

	// Datenbankverbindung verifizieren
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB-Ping fehlgeschlagen: %v", err)
	}

	db := client.Database(cfg.MongoDB.Database)

	// Adapter initialisieren (ausgehend)
	userRepo := mongodb.NewMongoUserRepository(db, cfg.MongoDB.Collection)
	passwordHasher := bcrypt.NewBcryptPasswordHasher()
	tokenGenerator := jwt.NewJWTTokenGenerator(cfg.JWT.Secret, cfg.JWT.ExpirySeconds)

	// Anwendungs-Service initialisieren
	authService := services.NewAuthService(userRepo, passwordHasher, tokenGenerator)

	// HTTP-Adapter initialisieren (eingehend)
	handler := rest.NewAuthHandler(authService)

	// HTTP-Server konfigurieren und starten
	router := gin.Default()
	rest.SetupRoutes(router, handler, tokenGenerator)

	log.Printf("Auth-Service startet auf Port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Server-Start fehlgeschlagen: %v", err)
	}
}

package config

import (
	"os"
	"strconv"
)

// Config haelt alle Konfigurationswerte fuer den Auth-Service.
type Config struct {
	Server  ServerConfig
	MongoDB MongoConfig
	JWT     JWTConfig
}

// ServerConfig enthaelt HTTP-Server-Einstellungen.
type ServerConfig struct {
	Port string
}

// MongoConfig enthaelt MongoDB-Verbindungseinstellungen.
type MongoConfig struct {
	URI        string
	Database   string
	Collection string
}

// JWTConfig enthaelt JWT-Token-Einstellungen.
type JWTConfig struct {
	Secret        string
	ExpirySeconds int
}

// Load liest die Konfiguration aus Umgebungsvariablen.
// Faellt auf Standardwerte zurueck wenn Variablen nicht gesetzt sind.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8081"),
		},
		MongoDB: MongoConfig{
			URI:        getEnv("MONGO_URI", "mongodb://localhost:27017"),
			Database:   getEnv("MONGO_DB", "auth_db"),
			Collection: getEnv("MONGO_COLLECTION", "users"),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "default"),
			ExpirySeconds: getEnvAsInt("JWT_EXPIRY", 3600),
		},
	}
}

// getEnv ruft eine Umgebungsvariable ab oder gibt einen Standardwert zurueck.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt ruft eine Umgebungsvariable als Integer ab.
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

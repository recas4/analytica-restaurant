package config

import "os"

// Config enthaelt alle Konfigurationsparameter fuer den Shopping-Service.
// Werte werden aus Umgebungsvariablen geladen mit sinnvollen Standardwerten.
type Config struct {
	MongoURI   string
	KafkaBkr   string
	KafkaTopic string
	JWTSecret  string
	ServerPort string
}

// Load liest die Konfiguration aus den Umgebungsvariablen.
// Falls eine Variable nicht gesetzt ist, wird der Standardwert verwendet.
func Load() *Config {
	return &Config{
		MongoURI:   getEnv("MONGO_URI", "mongodb://shopuser:shoppass@mongo:27017/shopping"),
		KafkaBkr:   getEnv("KAFKA_BROKER", "kafka:9092"),
		KafkaTopic: getEnv("KAFKA_TOPIC", "checkout"),
		JWTSecret:  getEnv("JWT_SECRET", ""),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

// getEnv liest eine Umgebungsvariable oder gibt den Standardwert zurueck.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

package config

import (
	"os"
)

// Config haelt alle Konfigurationswerte fuer den Checkout-Service.
type Config struct {
	Kafka KafkaConfig
}

// KafkaConfig enthaelt Kafka-Verbindungseinstellungen.
type KafkaConfig struct {
	Brokers       string
	CheckoutTopic string
	OrderTopic    string
	GroupID       string
}

// Load liest die Konfiguration aus Umgebungsvariablen.
// Faellt auf Standardwerte zurueck wenn Variablen nicht gesetzt sind.
func Load() *Config {
	return &Config{
		Kafka: KafkaConfig{
			Brokers:       getEnv("KAFKA_BROKERS", "kafka:9092"),
			CheckoutTopic: getEnv("KAFKA_CHECKOUT_TOPIC", "checkout"),
			OrderTopic:    getEnv("KAFKA_ORDER_TOPIC", "checkout-events"),
			GroupID:       getEnv("KAFKA_GROUP_ID", "checkout-group"),
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

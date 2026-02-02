package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL string
	Port        string
	GinMode     string // debug, release, test

	// Kafka
	KafkaBrokers  []string
	KafkaTopic    string
	KafkaGroupID  string

	// Kafka SASL (for Aiven)
	KafkaSASLEnabled  bool
	KafkaSASLUsername string
	KafkaSASLPassword string
	KafkaCAPath       string // Path to CA certificate for TLS
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        getEnvOrDefault("PORT", "8080"),
		GinMode:     getEnvOrDefault("GIN_MODE", "debug"),

		// Kafka
		KafkaBrokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
		KafkaTopic:   getEnvOrDefault("KAFKA_TOPIC", "lineage.events"),
		KafkaGroupID: getEnvOrDefault("KAFKA_GROUP_ID", "lineage-consumer-group"),

		// Kafka SASL
		KafkaSASLEnabled:  getEnvOrDefault("KAFKA_SASL_ENABLED", "false") == "true",
		KafkaSASLUsername: os.Getenv("KAFKA_SASL_USERNAME"),
		KafkaSASLPassword: os.Getenv("KAFKA_SASL_PASSWORD"),
		KafkaCAPath:       os.Getenv("KAFKA_CA_PATH"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if len(cfg.KafkaBrokers) == 0 || cfg.KafkaBrokers[0] == "" {
		cfg.KafkaBrokers = []string{"localhost:9092"}
	}

	// Validate SASL config if enabled
	if cfg.KafkaSASLEnabled {
		if cfg.KafkaSASLUsername == "" || cfg.KafkaSASLPassword == "" {
			return nil, fmt.Errorf("KAFKA_SASL_USERNAME and KAFKA_SASL_PASSWORD are required when KAFKA_SASL_ENABLED=true")
		}
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvIntOrDefault(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

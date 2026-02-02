package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL  string
	KafkaBrokers []string
	KafkaTopic   string
	Port         string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		KafkaBrokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
		KafkaTopic:   getEnvOrDefault("KAFKA_TOPIC", "lineage.events"),
		Port:         getEnvOrDefault("PORT", "8080"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if len(cfg.KafkaBrokers) == 0 || cfg.KafkaBrokers[0] == "" {
		cfg.KafkaBrokers = []string{"localhost:9092"}
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

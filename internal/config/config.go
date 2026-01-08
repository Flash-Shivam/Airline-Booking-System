package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	App      AppConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers       []string
	TopicBookings string
	TopicPayments string
	GroupID       string
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	CacheTTL          time.Duration
	LockTTL           time.Duration
	MaxCacheEntries   int
	TopSearchesPercent float64
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "airline_booking"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		Kafka: KafkaConfig{
			Brokers:       []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			TopicBookings: getEnv("KAFKA_TOPIC_BOOKINGS", "flight-bookings"),
			TopicPayments: getEnv("KAFKA_TOPIC_PAYMENTS", "payment-events"),
			GroupID:       getEnv("KAFKA_GROUP_ID", "booking-service"),
		},
		App: AppConfig{
			CacheTTL:          getDurationEnv("CACHE_TTL", time.Hour),
			LockTTL:           getDurationEnv("LOCK_TTL", 5*time.Minute),
			MaxCacheEntries:   getIntEnv("MAX_CACHE_ENTRIES", 1000),
			TopSearchesPercent: getFloatEnv("TOP_SEARCHES_PERCENT", 0.4),
		},
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv gets an integer environment variable with a default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getFloatEnv gets a float environment variable with a default value
func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// getDurationEnv gets a duration environment variable with a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

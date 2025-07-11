package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Webhook  WebhookConfig  `json:"webhook"`
	Security SecurityConfig `json:"security"`
	Logger   LoggerConfig   `json:"logger"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port    string `json:"port"`
	GinMode string `json:"gin_mode"`
	Host    string `json:"host"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Port     string `json:"port"`
	SSLMode  string `json:"ssl_mode"`
}

// WebhookConfig holds webhook configuration
type WebhookConfig struct {
	BaseURL        string `json:"base_url"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	MaxRetries     int    `json:"max_retries"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret                 string `json:"jwt_secret"`
	TimestampToleranceMinutes int    `json:"timestamp_tolerance_minutes"`
	HMACKeyLength             int    `json:"hmac_key_length"`
	JWTTokenExpiration        int    `json:"jwt_token_expiration"` // hours
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      string `json:"level"`
	Encoding   string `json:"encoding"`
	OutputPath string `json:"output_path"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "release"),
			Host:    getEnv("HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "loki_suite"),
			Port:     getEnv("DB_PORT", "5432"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Webhook: WebhookConfig{
			BaseURL:        getEnv("WEBHOOK_BASE_URL", "https://loki-suite.shavix.com"),
			TimeoutSeconds: getEnvAsInt("WEBHOOK_TIMEOUT_SECONDS", 30),
			MaxRetries:     getEnvAsInt("WEBHOOK_MAX_RETRIES", 3),
		},
		Security: SecurityConfig{
			JWTSecret:                 getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			TimestampToleranceMinutes: getEnvAsInt("TIMESTAMP_TOLERANCE_MINUTES", 5),
			HMACKeyLength:             getEnvAsInt("HMAC_KEY_LENGTH", 32),
			JWTTokenExpiration:        getEnvAsInt("JWT_TOKEN_EXPIRATION", 24), // 24 hours
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Encoding:   getEnv("LOG_ENCODING", "json"),
			OutputPath: getEnv("LOG_OUTPUT_PATH", "stdout"),
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Add validation logic here if needed
	return nil
}

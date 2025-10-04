package storage

import (
	"fmt"
	"os"
)

// Config holds configuration for storage initialization
type Config struct {
	ConnString string // PostgreSQL connection string
}

// NewStorage creates a new PostgreSQL storage instance
// This function provides dependency injection point for storage configuration
func NewStorage(config Config) (Storage, error) {
	return NewPostgresStorage(config.ConnString)
}

// NewStorageFromEnv creates storage from environment variables
// This is useful for containerized deployments
func NewStorageFromEnv() (Storage, error) {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		// Build connection string from individual env vars
		host := getEnvOrDefault("POSTGRES_HOST", "localhost")
		port := getEnvOrDefault("POSTGRES_PORT", "5432")
		user := getEnvOrDefault("POSTGRES_USER", "postgres")
		password := os.Getenv("POSTGRES_PASSWORD")
		dbname := getEnvOrDefault("POSTGRES_DB", "logs")
		sslmode := getEnvOrDefault("POSTGRES_SSLMODE", "disable")

		connString = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode,
		)
	}

	return NewStorage(Config{
		ConnString: connString,
	})
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

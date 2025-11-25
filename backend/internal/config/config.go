package config

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/aliancn/swiftlog/backend/internal/database"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/redis/go-redis/v9"
)

// GetEnv retrieves an environment variable with a fallback default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvDuration retrieves a duration environment variable with a fallback default value
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		duration, err := time.ParseDuration(value)
		if err != nil {
			log.Printf("Warning: Invalid duration for %s, using default: %v", key, defaultValue)
			return defaultValue
		}
		return duration
	}
	return defaultValue
}

// BuildDatabaseURL constructs a properly formatted PostgreSQL connection URL from
// individual components, handling special characters in the password through URL encoding.
// If DATABASE_URL is already set, it returns that value unchanged.
func BuildDatabaseURL() string {
	// If DATABASE_URL is already set, use it directly
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return dbURL
	}

	// Build from individual components
	user := GetEnv("POSTGRES_USER", "swiftlog")
	password := GetEnv("POSTGRES_PASSWORD", "changeme")
	host := GetEnv("POSTGRES_HOST", "postgres")
	port := GetEnv("POSTGRES_PORT", "5432")
	dbName := GetEnv("POSTGRES_DB", "swiftlog")
	sslMode := GetEnv("POSTGRES_SSLMODE", "disable")

	// URL-encode the password to handle special characters
	encodedPassword := url.QueryEscape(password)

	// Build the connection string
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, encodedPassword, host, port, dbName, sslMode)

	return dbURL
}

// InitDatabase initializes a database connection with standard connection pool settings
func InitDatabase(ctx context.Context, dbURL string) (*database.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &database.DB{DB: db}, nil
}

// InitRedis initializes a Redis client connection
func InitRedis(ctx context.Context, redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}

// GetEncryptionKey retrieves or generates the 32-byte AES-256 encryption key
// The serviceName is used as a seed for development mode key generation
func GetEncryptionKey(serviceName string) ([]byte, error) {
	environment := GetEnv("ENVIRONMENT", "development")
	keyStr := os.Getenv("ENCRYPTION_KEY")

	if keyStr != "" {
		// Decode base64-encoded key from environment
		key, err := base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid ENCRYPTION_KEY format (must be base64): %w", err)
		}
		if len(key) != 32 {
			return nil, fmt.Errorf("ENCRYPTION_KEY must be 32 bytes (256 bits), got %d bytes", len(key))
		}
		log.Println("Using encryption key from ENCRYPTION_KEY environment variable")
		return key, nil
	}

	// In production, encryption key is required
	if environment == "production" {
		return nil, fmt.Errorf("ENCRYPTION_KEY environment variable must be set in production")
	}

	// In development, use a deterministic key derived from service name (for convenience)
	hash := sha256.Sum256([]byte("swiftlog-" + serviceName + "-encryption-key"))
	log.Printf("WARNING: Using deterministic encryption key in development mode for %s", serviceName)
	log.Println("WARNING: Set ENCRYPTION_KEY environment variable in production!")
	log.Printf("To generate a secure key: openssl rand -base64 32")
	return hash[:], nil
}

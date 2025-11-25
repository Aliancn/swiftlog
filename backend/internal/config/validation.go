package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return "configuration validation failed:\n  - " + strings.Join(msgs, "\n  - ")
}

// ConfigValidator validates environment configuration
type ConfigValidator struct {
	errors ValidationErrors
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		errors: make(ValidationErrors, 0),
	}
}

// RequireEnv validates that an environment variable is set and not empty
func (v *ConfigValidator) RequireEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   name,
			Message: "required but not set",
		})
	}
	return value
}

// RequireEnvInProduction validates that an environment variable is set in production
func (v *ConfigValidator) RequireEnvInProduction(name, environment string) string {
	value := os.Getenv(name)
	if environment == "production" && value == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   name,
			Message: "required in production but not set",
		})
	}
	return value
}

// ValidateURL validates that a value is a valid URL
func (v *ConfigValidator) ValidateURL(name, value string) {
	if value == "" {
		return
	}
	if _, err := url.Parse(value); err != nil {
		v.errors = append(v.errors, ValidationError{
			Field:   name,
			Message: fmt.Sprintf("invalid URL: %v", err),
		})
	}
}

// ValidatePort validates that a value is a valid port number
func (v *ConfigValidator) ValidatePort(name, value string) {
	if value == "" {
		return
	}
	port, err := strconv.Atoi(value)
	if err != nil {
		v.errors = append(v.errors, ValidationError{
			Field:   name,
			Message: "must be a valid port number",
		})
		return
	}
	if port < 1 || port > 65535 {
		v.errors = append(v.errors, ValidationError{
			Field:   name,
			Message: "port must be between 1 and 65535",
		})
	}
}

// ValidateEnum validates that a value is one of the allowed values
func (v *ConfigValidator) ValidateEnum(name, value string, allowed []string) {
	if value == "" {
		return
	}
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	v.errors = append(v.errors, ValidationError{
		Field:   name,
		Message: fmt.Sprintf("must be one of: %s (got: %s)", strings.Join(allowed, ", "), value),
	})
}

// ValidateMinLength validates that a value meets minimum length requirement
func (v *ConfigValidator) ValidateMinLength(name, value string, minLength int) {
	if value == "" {
		return
	}
	if len(value) < minLength {
		v.errors = append(v.errors, ValidationError{
			Field:   name,
			Message: fmt.Sprintf("must be at least %d characters (got: %d)", minLength, len(value)),
		})
	}
}

// Errors returns all validation errors
func (v *ConfigValidator) Errors() error {
	if len(v.errors) == 0 {
		return nil
	}
	return v.errors
}

// ValidateConfig validates all required environment variables
func ValidateConfig() error {
	v := NewConfigValidator()

	environment := GetEnv("ENVIRONMENT", "development")

	// Database configuration - build from components if needed
	databaseURL := BuildDatabaseURL()
	if databaseURL == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   "DATABASE_URL",
			Message: "required but not set (or unable to build from POSTGRES_* variables)",
		})
	} else {
		v.ValidateURL("DATABASE_URL", databaseURL)
	}

	// Loki configuration
	lokiURL := GetEnv("LOKI_URL", "http://localhost:3100")
	v.ValidateURL("LOKI_URL", lokiURL)

	// Redis configuration
	redisURL := GetEnv("REDIS_URL", "redis://localhost:6379")
	v.ValidateURL("REDIS_URL", redisURL)

	// Port configuration
	apiPort := GetEnv("API_PORT", "8080")
	v.ValidatePort("API_PORT", apiPort)

	wsPort := GetEnv("WS_PORT", "8081")
	v.ValidatePort("WS_PORT", wsPort)

	grpcPort := GetEnv("GRPC_PORT", "50051")
	v.ValidatePort("GRPC_PORT", grpcPort)

	// Security configuration
	jwtSecret := GetEnv("JWT_SECRET", "")
	if environment == "production" {
		if jwtSecret == "" {
			v.errors = append(v.errors, ValidationError{
				Field:   "JWT_SECRET",
				Message: "required in production. Generate with: openssl rand -base64 32",
			})
		} else {
			v.ValidateMinLength("JWT_SECRET", jwtSecret, 32)
		}
	}

	// Encryption key (required in production)
	encryptionKey := GetEnv("ENCRYPTION_KEY", "")
	if environment == "production" && encryptionKey == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   "ENCRYPTION_KEY",
			Message: "required in production. Generate with: openssl rand -base64 32",
		})
	}

	// Environment validation
	v.ValidateEnum("ENVIRONMENT", environment, []string{"development", "production", "staging", "test"})

	// Log level validation
	logLevel := GetEnv("LOG_LEVEL", "info")
	v.ValidateEnum("LOG_LEVEL", logLevel, []string{"debug", "info", "warn", "error", "fatal"})

	// CORS configuration warning
	corsOrigins := GetEnv("CORS_ORIGINS", "")
	if environment == "production" && corsOrigins == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   "CORS_ORIGINS",
			Message: "should be set in production to restrict access",
		})
	}

	return v.Errors()
}

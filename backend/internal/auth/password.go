package auth

import (
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

const (
	// MinPasswordLength is the minimum required password length
	MinPasswordLength = 8
	// BcryptCost is the cost factor for bcrypt hashing
	BcryptCost = 12
)

var (
	// Password strength regex patterns
	hasUpperCase   = regexp.MustCompile(`[A-Z]`)
	hasLowerCase   = regexp.MustCompile(`[a-z]`)
	hasNumber      = regexp.MustCompile(`[0-9]`)
	hasSpecialChar = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`)
)

// ValidatePasswordStrength checks if password meets strength requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}

	if !hasUpperCase.MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !hasLowerCase.MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !hasNumber.MatchString(password) {
		return fmt.Errorf("password must contain at least one number")
	}

	if !hasSpecialChar.MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	// Validate password strength
	if err := ValidatePasswordStrength(password); err != nil {
		return "", err
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(bytes), nil
}

// VerifyPassword checks if the provided password matches the hash
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

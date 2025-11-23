package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptionService handles encryption and decryption of sensitive data
type EncryptionService struct {
	key []byte
}

// NewEncryptionService creates a new encryption service
// The key must be 32 bytes for AES-256
func NewEncryptionService(key []byte) (*EncryptionService, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be exactly 32 bytes for AES-256, got %d bytes", len(key))
	}
	return &EncryptionService{key: key}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
// Returns base64-encoded ciphertext
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Create cipher block
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-256-GCM
func (s *EncryptionService) Decrypt(ciphertextBase64 string) (string, error) {
	if ciphertextBase64 == "" {
		return "", nil
	}

	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

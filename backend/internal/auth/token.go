package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/aliancn/swiftlog/backend/internal/models"
	"github.com/google/uuid"
)

const (
	// TokenByteLength is the length of the raw token in bytes (32 bytes = 256 bits)
	TokenByteLength = 32
	// TokenStringLength is the length of the base64-encoded token (44 characters)
	TokenStringLength = 44
)

// TokenService handles API token operations
type TokenService struct {
	db *sql.DB
}

// NewTokenService creates a new token service
func NewTokenService(db *sql.DB) *TokenService {
	return &TokenService{db: db}
}

// GenerateToken generates a new random API token
func GenerateToken() (string, error) {
	bytes := make([]byte, TokenByteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashToken creates a SHA-256 hash of the token for storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// CreateToken creates a new API token for a user
func (s *TokenService) CreateToken(ctx context.Context, userID uuid.UUID, name string) (string, *models.APIToken, error) {
	// Generate raw token
	rawToken, err := GenerateToken()
	if err != nil {
		return "", nil, err
	}

	// Hash the token for storage
	tokenHash := HashToken(rawToken)

	// Store the token
	token := &models.APIToken{}
	query := `
		INSERT INTO api_tokens (user_id, token_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, name, created_at
	`
	err = s.db.QueryRowContext(ctx, query, userID, tokenHash, name).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.Name,
		&token.CreatedAt,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create token: %w", err)
	}

	return rawToken, token, nil
}

// ValidateToken validates an API token and returns the associated user ID
func (s *TokenService) ValidateToken(ctx context.Context, rawToken string) (uuid.UUID, error) {
	tokenHash := HashToken(rawToken)

	var userID uuid.UUID
	query := `
		SELECT user_id
		FROM api_tokens
		WHERE token_hash = $1
	`
	err := s.db.QueryRowContext(ctx, query, tokenHash).Scan(&userID)
	if err == sql.ErrNoRows {
		return uuid.Nil, fmt.Errorf("invalid token")
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to validate token: %w", err)
	}

	return userID, nil
}

// GetTokenByID retrieves a token by ID
func (s *TokenService) GetTokenByID(ctx context.Context, tokenID uuid.UUID) (*models.APIToken, error) {
	token := &models.APIToken{}
	query := `
		SELECT id, user_id, name, created_at
		FROM api_tokens
		WHERE id = $1
	`
	err := s.db.QueryRowContext(ctx, query, tokenID).Scan(
		&token.ID,
		&token.UserID,
		&token.Name,
		&token.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return token, nil
}

// RevokeToken deletes an API token
func (s *TokenService) RevokeToken(ctx context.Context, tokenID uuid.UUID) error {
	query := `DELETE FROM api_tokens WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

// ListTokensByUserID retrieves all tokens for a user (without the token hash)
func (s *TokenService) ListTokensByUserID(ctx context.Context, userID uuid.UUID) ([]*models.APIToken, error) {
	query := `
		SELECT id, user_id, name, created_at
		FROM api_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*models.APIToken
	for rows.Next() {
		token := &models.APIToken{}
		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.Name,
			&token.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

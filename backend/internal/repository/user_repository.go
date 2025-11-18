package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aliancn/swiftlog/backend/internal/models"
	"github.com/google/uuid"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user with password
func (r *UserRepository) Create(ctx context.Context, username, passwordHash string, isAdmin bool) (*models.User, error) {
	user := &models.User{}
	query := `
		INSERT INTO users (username, password_hash, is_admin)
		VALUES ($1, $2, $3)
		RETURNING id, username, is_admin, created_at
	`
	err := r.db.QueryRowContext(ctx, query, username, passwordHash, isAdmin).Scan(
		&user.ID,
		&user.Username,
		&user.IsAdmin,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	user.PasswordHash = passwordHash
	return user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, password_hash, is_admin, created_at FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, password_hash, is_admin, created_at FROM users WHERE username = $1`
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetOrCreateUser gets an existing user or creates if not exists (for backward compatibility)
func (r *UserRepository) GetOrCreateUser(ctx context.Context, username, passwordHash string) (*models.User, error) {
	// Try to get existing user
	user, err := r.GetByUsername(ctx, username)
	if err == nil {
		return user, nil
	}

	// Create new user if not found
	return r.Create(ctx, username, passwordHash, false)
}

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// ListAll retrieves all users
func (r *UserRepository) ListAll(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, username, is_admin, created_at
		FROM users
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.IsAdmin,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

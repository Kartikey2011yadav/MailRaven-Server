package sqlite

import (
	"context"
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// UserRepository implements ports.UserRepository using SQLite
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new SQLite user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user with hashed password
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, password_hash, created_at, last_login_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.Email, user.PasswordHash, user.CreatedAt.Unix(), user.LastLoginAt.Unix(),
	)
	if err != nil {
		// Check for unique constraint violation
		if err.Error() == "UNIQUE constraint failed: users.email" {
			return ports.ErrAlreadyExists
		}
		return ports.ErrStorageFailure
	}

	return nil
}

// FindByEmail retrieves user by email address
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT email, password_hash, created_at, last_login_at
		FROM users
		WHERE email = ?
	`

	user := &domain.User{}
	var createdAtUnix, lastLoginAtUnix int64

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.Email, &user.PasswordHash, &createdAtUnix, &lastLoginAtUnix,
	)

	if err == sql.ErrNoRows {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, ports.ErrStorageFailure
	}

	user.CreatedAt = time.Unix(createdAtUnix, 0)
	user.LastLoginAt = time.Unix(lastLoginAtUnix, 0)

	return user, nil
}

// Authenticate verifies email/password and returns user
func (r *UserRepository) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := r.FindByEmail(ctx, email)
	if err != nil {
		if err == ports.ErrNotFound {
			return nil, ports.ErrInvalidCredentials
		}
		return nil, err
	}

	// Compare hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ports.ErrInvalidCredentials
	}

	return user, nil
}

// UpdateLastLogin updates the LastLoginAt timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, email string) error {
	query := `UPDATE users SET last_login_at = ? WHERE email = ?`

	result, err := r.db.ExecContext(ctx, query, time.Now().Unix(), email)
	if err != nil {
		return ports.ErrStorageFailure
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return ports.ErrStorageFailure
	}
	if rowsAffected == 0 {
		return ports.ErrNotFound
	}

	return nil
}

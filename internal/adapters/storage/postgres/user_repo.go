package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// UserRepository implements ports.UserRepository using PostgreSQL
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user with hashed password
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.Role == "" {
		user.Role = domain.RoleUser
	}
	query := `
		INSERT INTO users (email, password_hash, role, created_at, last_login_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.Email, user.PasswordHash, user.Role, user.CreatedAt, user.LastLoginAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return ports.ErrAlreadyExists
		}
		return ports.ErrStorageFailure
	}

	return nil
}

// FindByEmail retrieves user by email address
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT email, password_hash, role, created_at, last_login_at
		FROM users
		WHERE email = $1
	`

	user := &domain.User{}
	var role sql.NullString
	// Postgres driver handles time.Time scanning automatically for TIMESTAMP columns

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.Email, &user.PasswordHash, &role, &user.CreatedAt, &user.LastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, ports.ErrStorageFailure
	}

	if role.Valid {
		user.Role = domain.Role(role.String)
	} else {
		user.Role = domain.RoleUser
	}

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

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

func (r *UserRepository) UpdateLastLogin(ctx context.Context, email string) error {
	query := `UPDATE users SET last_login_at = $1 WHERE email = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), email)
	if err != nil {
		return ports.ErrStorageFailure
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	query := `
		SELECT email, role, created_at, last_login_at 
		FROM users 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, ports.ErrStorageFailure
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		var role sql.NullString
		if err := rows.Scan(&u.Email, &role, &u.CreatedAt, &u.LastLoginAt); err != nil {
			return nil, ports.ErrStorageFailure
		}
		if role.Valid {
			u.Role = domain.Role(role.String)
		} else {
			u.Role = domain.RoleUser
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) Delete(ctx context.Context, email string) error {
	query := `DELETE FROM users WHERE email = $1`
	res, err := r.db.ExecContext(ctx, query, email)
	if err != nil {
		return ports.ErrStorageFailure
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return ports.ErrStorageFailure
	}
	if rows == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, email, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE email = $2`
	res, err := r.db.ExecContext(ctx, query, passwordHash, email)
	if err != nil {
		return ports.ErrStorageFailure
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return ports.ErrStorageFailure
	}
	if rows == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, email string, role domain.Role) error {
	query := `UPDATE users SET role = $1 WHERE email = $2`
	res, err := r.db.ExecContext(ctx, query, role, email)
	if err != nil {
		return ports.ErrStorageFailure
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return ports.ErrStorageFailure
	}
	if rows == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (r *UserRepository) Count(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total count
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		return nil, ports.ErrStorageFailure
	}
	stats["total"] = total

	// Count by role? Interface just asks for map. Usually implemented as such.
	rows, err := r.db.QueryContext(ctx, "SELECT role, COUNT(*) FROM users GROUP BY role")
	if err != nil {
		return nil, ports.ErrStorageFailure
	}
	defer rows.Close()

	for rows.Next() {
		var role string
		var count int64
		if err := rows.Scan(&role, &count); err == nil {
			stats[role] = count
		}
	}

	return stats, nil
}

// UpdateQuota sets the max storage in bytes for a user (0 = unlimited)
func (r *UserRepository) UpdateQuota(ctx context.Context, email string, bytes int64) error {
	return ports.ErrStorageFailure // Not implemented
}

// IncrementStorageUsed updates storage usage by delta (can be negative)
func (r *UserRepository) IncrementStorageUsed(ctx context.Context, email string, delta int64) error {
	return ports.ErrStorageFailure // Not implemented
}

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
	if user.Role == "" {
		user.Role = domain.RoleUser
	}
	query := `
		INSERT INTO users (email, password_hash, role, created_at, last_login_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.Email, user.PasswordHash, user.Role, user.CreatedAt.Unix(), user.LastLoginAt.Unix(),
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
		SELECT email, password_hash, role, created_at, last_login_at
		FROM users
		WHERE email = ?
	`

	user := &domain.User{}
	var createdAtUnix, lastLoginAtUnix int64
	var role sql.NullString

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.Email, &user.PasswordHash, &role, &createdAtUnix, &lastLoginAtUnix,
	)

	if err == sql.ErrNoRows {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, ports.ErrStorageFailure
	}

	user.CreatedAt = time.Unix(createdAtUnix, 0)
	user.LastLoginAt = time.Unix(lastLoginAtUnix, 0)
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

// List users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	query := `
		SELECT email, role, created_at, last_login_at
		FROM users
		ORDER BY email ASC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, ports.ErrStorageFailure
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		var createdAt, lastLoginAt int64
		var role sql.NullString

		if err := rows.Scan(&user.Email, &role, &createdAt, &lastLoginAt); err != nil {
			continue
		}

		user.CreatedAt = time.Unix(createdAt, 0)
		user.LastLoginAt = time.Unix(lastLoginAt, 0)
		if role.Valid {
			user.Role = domain.Role(role.String)
		} else {
			user.Role = domain.RoleUser
		}

		users = append(users, &user)
	}

	return users, nil
}

// Delete removes a user
func (r *UserRepository) Delete(ctx context.Context, email string) error {
	query := "DELETE FROM users WHERE email = ?"
	result, err := r.db.ExecContext(ctx, query, email)
	if err != nil {
		return ports.ErrStorageFailure
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ports.ErrNotFound
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, email, passwordHash string) error {
	query := "UPDATE users SET password_hash = ? WHERE email = ?"
	result, err := r.db.ExecContext(ctx, query, passwordHash, email)
	if err != nil {
		return ports.ErrStorageFailure
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ports.ErrNotFound
	}

	return nil
}

// UpdateRole changes a user's role
func (r *UserRepository) UpdateRole(ctx context.Context, email string, role domain.Role) error {
	query := "UPDATE users SET role = ? WHERE email = ?"
	result, err := r.db.ExecContext(ctx, query, string(role), email)
	if err != nil {
		return ports.ErrStorageFailure
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ports.ErrNotFound
	}

	return nil
}

// Count returns user statistics
func (r *UserRepository) Count(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		return nil, err
	}
	stats["total"] = total

	// Active (non-suspended, simpler for now just assume all exist)
	// For this quick implementation we'll just query role='ADMIN' for admins
	var adminCount int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE role = ?", domain.RoleAdmin).Scan(&adminCount); err != nil {
		return nil, err
	}
	stats["admin"] = adminCount

	// Active (users who have logged in at least once, or just total for now)
	stats["active"] = total // Simplification

	return stats, nil
}



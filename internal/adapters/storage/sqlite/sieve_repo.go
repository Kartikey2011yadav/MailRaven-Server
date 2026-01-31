package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain/sieve"
	"github.com/google/uuid"
)

type SqliteScriptRepository struct {
	db *sql.DB
}

func NewSqliteScriptRepository(db *sql.DB) *SqliteScriptRepository {
	return &SqliteScriptRepository{db: db}
}

func (r *SqliteScriptRepository) Save(ctx context.Context, script *sieve.SieveScript) error {
	if script.ID == "" {
		script.ID = uuid.New().String()
	}
	now := time.Now()
	script.UpdatedAt = now
	if script.CreatedAt.IsZero() {
		script.CreatedAt = now
	}

	query := `
		INSERT INTO sieve_scripts (id, user_id, name, content, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, name) DO UPDATE SET
			content = excluded.content,
			updated_at = excluded.updated_at
	`
	_, err := r.db.ExecContext(ctx, query,
		script.ID,
		script.UserID,
		script.Name,
		script.Content,
		script.IsActive,
		script.CreatedAt,
		script.UpdatedAt,
	)
	return err
}

func (r *SqliteScriptRepository) Get(ctx context.Context, userID, name string) (*sieve.SieveScript, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, user_id, name, content, is_active, created_at, updated_at FROM sieve_scripts WHERE user_id = ? AND name = ?", userID, name)
	var s sieve.SieveScript
	err := row.Scan(&s.ID, &s.UserID, &s.Name, &s.Content, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Return nil if found, or wrapped err? Usually nil for repo Get is idiomatic if not found.
		// Actually typical repo pattern returns err not found. Let's return nil, nil to indicate "not found" nicely or wrapped error.
		// For now returning nil, nil.
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SqliteScriptRepository) GetActive(ctx context.Context, userID string) (*sieve.SieveScript, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, user_id, name, content, is_active, created_at, updated_at FROM sieve_scripts WHERE user_id = ? AND is_active = 1", userID)
	var s sieve.SieveScript
	err := row.Scan(&s.ID, &s.UserID, &s.Name, &s.Content, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SqliteScriptRepository) List(ctx context.Context, userID string) ([]sieve.SieveScript, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, user_id, name, content, is_active, created_at, updated_at FROM sieve_scripts WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scripts []sieve.SieveScript
	for rows.Next() {
		var s sieve.SieveScript
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Content, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		scripts = append(scripts, s)
	}
	return scripts, nil
}

func (r *SqliteScriptRepository) SetActive(ctx context.Context, userID, name string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		//nolint:errcheck // Rollback is best-effort
		_ = tx.Rollback()
	}()

	// Deactivate all
	_, err = tx.ExecContext(ctx, "UPDATE sieve_scripts SET is_active = 0 WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	if name != "" {
		// Activate specific one
		res, err := tx.ExecContext(ctx, "UPDATE sieve_scripts SET is_active = 1 WHERE user_id = ? AND name = ?", userID, name)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			// Script not found with that name
			return errors.New("script not found")
		}
	}

	return tx.Commit()
}

func (r *SqliteScriptRepository) Delete(ctx context.Context, userID, name string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sieve_scripts WHERE user_id = ? AND name = ?", userID, name)
	return err
}

package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"modernc.org/sqlite" // For error code checks
)

type DomainRepository struct {
	db *sql.DB
}

func NewDomainRepository(db *sql.DB) *DomainRepository {
	return &DomainRepository{db: db}
}

func (r *DomainRepository) Create(ctx context.Context, d *domain.Domain) error {
	query := `
		INSERT INTO domains (name, created_at, updated_at, active, dkim_selector, dkim_private_key, dkim_public_key)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		d.Name,
		d.CreatedAt.Unix(),
		d.UpdatedAt.Unix(),
		d.Active,
		d.DKIMSelector,
		d.DKIMPrivateKey,
		d.DKIMPublicKey,
	)

	if err != nil {
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code() == 1555 { // SQLITE_CONSTRAINT_PRIMARYKEY
			return ports.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *DomainRepository) Get(ctx context.Context, name string) (*domain.Domain, error) {
	query := `
		SELECT name, created_at, updated_at, active, dkim_selector, dkim_private_key, dkim_public_key
		FROM domains WHERE name = ?
	`
	row := r.db.QueryRowContext(ctx, query, name)

	var d domain.Domain
	var createdAt, updatedAt int64
	var privateKey, selector, publicKey sql.NullString

	err := row.Scan(
		&d.Name,
		&createdAt,
		&updatedAt,
		&d.Active,
		&selector,
		&privateKey,
		&publicKey,
	)

	if err == sql.ErrNoRows {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	d.CreatedAt = time.Unix(createdAt, 0).UTC()
	d.UpdatedAt = time.Unix(updatedAt, 0).UTC()
	if selector.Valid {
		d.DKIMSelector = selector.String
	}
	if privateKey.Valid {
		d.DKIMPrivateKey = privateKey.String
	}
	if publicKey.Valid {
		d.DKIMPublicKey = publicKey.String
	}

	return &d, nil
}

func (r *DomainRepository) List(ctx context.Context, limit, offset int) ([]*domain.Domain, error) {
	query := `
		SELECT name, created_at, updated_at, active, dkim_selector, dkim_public_key
		FROM domains
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*domain.Domain
	for rows.Next() {
		var d domain.Domain
		var createdAt, updatedAt int64
		var selector, publicKey sql.NullString

		if err := rows.Scan(
			&d.Name,
			&createdAt,
			&updatedAt,
			&d.Active,
			&selector,
			&publicKey,
		); err != nil {
			return nil, err
		}

		d.CreatedAt = time.Unix(createdAt, 0).UTC()
		d.UpdatedAt = time.Unix(updatedAt, 0).UTC()
		if selector.Valid {
			d.DKIMSelector = selector.String
		}
		if publicKey.Valid {
			d.DKIMPublicKey = publicKey.String
		}

		domains = append(domains, &d)
	}
	return domains, nil
}

func (r *DomainRepository) Delete(ctx context.Context, name string) error {
	query := "DELETE FROM domains WHERE name = ?"
	res, err := r.db.ExecContext(ctx, query, name)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (r *DomainRepository) Exists(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM domains WHERE name = ?)"
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	return exists, err
}

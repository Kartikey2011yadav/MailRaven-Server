package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
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
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		d.Name,
		d.CreatedAt,
		d.UpdatedAt,
		d.Active,
		d.DKIMSelector,
		d.DKIMPrivateKey,
		d.DKIMPublicKey,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return ports.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *DomainRepository) Get(ctx context.Context, name string) (*domain.Domain, error) {
	query := `
		SELECT name, created_at, updated_at, active, dkim_selector, dkim_private_key, dkim_public_key
		FROM domains WHERE name = $1
	`
	row := r.db.QueryRowContext(ctx, query, name)

	var d domain.Domain
	var privateKey, selector, publicKey sql.NullString

	err := row.Scan(
		&d.Name,
		&d.CreatedAt,
		&d.UpdatedAt,
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
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*domain.Domain
	for rows.Next() {
		var d domain.Domain
		var selector, publicKey sql.NullString

		err := rows.Scan(
			&d.Name,
			&d.CreatedAt,
			&d.UpdatedAt,
			&d.Active,
			&selector,
			&publicKey,
		)
		if err != nil {
			return nil, err
		}

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
	query := `DELETE FROM domains WHERE name = $1`
	result, err := r.db.ExecContext(ctx, query, name)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (r *DomainRepository) Exists(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM domains WHERE name = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

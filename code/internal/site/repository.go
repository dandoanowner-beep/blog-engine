package site

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Get(ctx context.Context, key string) (string, error) {
	var content string
	err := r.db.QueryRow(ctx,
		`SELECT content FROM site_content WHERE key=$1`, key,
	).Scan(&content)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return content, err
}

func (r *PostgresRepository) Upsert(ctx context.Context, key, content string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO site_content (key, content, updated_at) VALUES ($1, $2, NOW())
	     ON CONFLICT (key) DO UPDATE SET content=EXCLUDED.content, updated_at=NOW()`,
		key, content,
	)
	return err
}

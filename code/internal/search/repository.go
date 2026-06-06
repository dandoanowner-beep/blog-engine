package search

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) SearchBlogs(ctx context.Context, q string, viewerID uuid.UUID, page int) ([]*BlogResult, int, error) {
	return []*BlogResult{}, 0, nil
}

func (r *PostgresRepository) SearchUsers(ctx context.Context, q string, page int) ([]*UserResult, int, error) {
	return []*UserResult{}, 0, nil
}

func (r *PostgresRepository) SearchTags(ctx context.Context, q string, page int) ([]*TagResult, int, error) {
	return []*TagResult{}, 0, nil
}

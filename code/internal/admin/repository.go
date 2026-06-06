package admin

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) ListUsers(ctx context.Context, page int, role string) ([]*UserRow, int, error) {
	return []*UserRow{}, 0, nil
}

func (r *PostgresRepository) ChangeUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	return nil
}

func (r *PostgresRepository) ListReports(ctx context.Context, status string, page int) ([]*ReportRow, int, error) {
	return []*ReportRow{}, 0, nil
}

func (r *PostgresRepository) ResolveReport(ctx context.Context, reportID uuid.UUID, action string, resolverID uuid.UUID) error {
	return nil
}

func (r *PostgresRepository) DeleteContent(ctx context.Context, reportID uuid.UUID) error {
	return nil
}

func (r *PostgresRepository) GetStats(ctx context.Context) (*Stats, error) {
	return &Stats{}, nil
}

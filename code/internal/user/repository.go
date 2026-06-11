package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetByUsername(ctx context.Context, username string) (*Profile, error) {
	return nil, ErrNotFound
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Profile, error) {
	return nil, ErrNotFound
}

func (r *PostgresRepository) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*Profile, error) {
	return nil, nil
}

func (r *PostgresRepository) UsernameExists(ctx context.Context, username string, excludeID uuid.UUID) (bool, error) {
	return false, nil
}

func (r *PostgresRepository) IsFriend(ctx context.Context, viewerID, profileID uuid.UUID) (bool, error) {
	return false, nil
}

func (r *PostgresRepository) UpdateLanguagePreference(ctx context.Context, userID uuid.UUID, lang string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET language_preference = $1 WHERE id = $2`,
		lang, userID,
	)
	return err
}

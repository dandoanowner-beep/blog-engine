package portfolio

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

const projectColumns = `id, title, description, tech_stack, repo_url, demo_url, thumbnail_url, sort_order, created_at, updated_at`

func scanProject(row pgx.Row) (*Project, error) {
	p := &Project{}
	err := row.Scan(&p.ID, &p.Title, &p.Description, &p.TechStack, &p.RepoURL,
		&p.DemoURL, &p.ThumbnailURL, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *PostgresRepository) Create(ctx context.Context, p *Project) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO projects (id, title, description, tech_stack, repo_url, demo_url, thumbnail_url, sort_order, created_at, updated_at)
	     VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		p.ID, p.Title, p.Description, p.TechStack, p.RepoURL, p.DemoURL,
		p.ThumbnailURL, p.SortOrder, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	return scanProject(r.db.QueryRow(ctx,
		`SELECT `+projectColumns+` FROM projects WHERE id=$1`, id))
}

func (r *PostgresRepository) Update(ctx context.Context, p *Project) error {
	_, err := r.db.Exec(ctx,
		`UPDATE projects SET title=$2, description=$3, tech_stack=$4, repo_url=$5,
	            demo_url=$6, thumbnail_url=$7, sort_order=$8, updated_at=$9
	     WHERE id=$1`,
		p.ID, p.Title, p.Description, p.TechStack, p.RepoURL,
		p.DemoURL, p.ThumbnailURL, p.SortOrder, p.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM projects WHERE id=$1`, id)
	return err
}

func (r *PostgresRepository) List(ctx context.Context) ([]*Project, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+projectColumns+` FROM projects ORDER BY sort_order ASC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []*Project
	for rows.Next() {
		p := &Project{}
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.TechStack, &p.RepoURL,
			&p.DemoURL, &p.ThumbnailURL, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

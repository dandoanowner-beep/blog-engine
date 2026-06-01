package blog

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, b *Blog) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO blogs (id,author_id,title,content,excerpt,thumbnail_url,privacy,status,feed_score,read_time_min,published_at,created_at,updated_at)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		b.ID, b.AuthorID, b.Title, b.Content, b.Excerpt, b.ThumbnailURL,
		b.Privacy, b.Status, b.FeedScore, b.ReadTimeMin, b.PublishedAt, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Blog, error) {
	b := &Blog{}
	err := r.db.QueryRow(ctx,
		`SELECT id,author_id,title,content,excerpt,thumbnail_url,privacy,status,like_count,dislike_count,comment_count,feed_score,read_time_min,published_at,created_at,updated_at
         FROM blogs WHERE id=$1`, id,
	).Scan(&b.ID, &b.AuthorID, &b.Title, &b.Content, &b.Excerpt, &b.ThumbnailURL,
		&b.Privacy, &b.Status, &b.LikeCount, &b.DislikeCount, &b.CommentCount,
		&b.FeedScore, &b.ReadTimeMin, &b.PublishedAt, &b.CreatedAt, &b.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return b, err
}

func (r *PostgresRepository) Update(ctx context.Context, b *Blog) error {
	_, err := r.db.Exec(ctx,
		`UPDATE blogs SET title=$2,content=$3,thumbnail_url=$4,privacy=$5,status=$6,updated_at=$7 WHERE id=$1`,
		b.ID, b.Title, b.Content, b.ThumbnailURL, b.Privacy, b.Status, time.Now(),
	)
	return err
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM blogs WHERE id=$1`, id)
	return err
}

func (r *PostgresRepository) IsBlocked(ctx context.Context, viewerID, authorID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM blocks WHERE (blocker_id=$1 AND blocked_id=$2) OR (blocker_id=$2 AND blocked_id=$1))`,
		viewerID, authorID,
	).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) AreFriends(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	var exists bool
	a, b := orderUUIDs(userA, userB)
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM friends WHERE user_id_1=$1 AND user_id_2=$2)`, a, b,
	).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) UpsertTags(ctx context.Context, names []string) ([]Tag, error) {
	tags := make([]Tag, 0, len(names))
	for _, name := range names {
		slug := slugify(name)
		var t Tag
		err := r.db.QueryRow(ctx,
			`INSERT INTO tags (id, name, slug) VALUES ($1,$2,$3)
             ON CONFLICT (slug) DO UPDATE SET name=EXCLUDED.name
             RETURNING id, name, slug`,
			uuid.New(), name, slug,
		).Scan(&t.ID, &t.Name, &t.Slug)
		if err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (r *PostgresRepository) UpsertCategories(ctx context.Context, ids []uuid.UUID) error {
	return nil // categories already exist; blog_categories join handled separately
}

func (r *PostgresRepository) UpdateFeedScore(ctx context.Context, blogID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE blogs SET feed_score = (like_count*3 + comment_count*2 + GREATEST(0, 100 - EXTRACT(EPOCH FROM NOW()-published_at)/3600*2)) WHERE id=$1`,
		blogID,
	)
	return err
}

func orderUUIDs(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if a.String() < b.String() {
		return a, b
	}
	return b, a
}

func slugify(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result = append(result, c+32)
		} else if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			result = append(result, c)
		} else if c == ' ' || c == '-' || c == '_' {
			result = append(result, '-')
		}
	}
	return string(result)
}

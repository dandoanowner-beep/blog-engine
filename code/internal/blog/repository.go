package blog

import (
	"context"
	"encoding/json"
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
		`INSERT INTO blogs (id,author_id,title,content,excerpt,thumbnail_url,privacy,status,feed_score,read_time_min,translation_status,published_at,created_at,updated_at)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		b.ID, b.AuthorID, b.Title, b.Content, b.Excerpt, b.ThumbnailURL,
		b.Privacy, b.Status, b.FeedScore, b.ReadTimeMin, b.TranslationStatus,
		b.PublishedAt, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Blog, error) {
	b := &Blog{}
	err := r.db.QueryRow(ctx,
		`SELECT id,author_id,title,content,excerpt,thumbnail_url,privacy,status,
		        like_count,dislike_count,comment_count,feed_score,read_time_min,
		        COALESCE(title_en,''),COALESCE(body_en,''),translation_status,
		        published_at,created_at,updated_at
         FROM blogs WHERE id=$1`, id,
	).Scan(&b.ID, &b.AuthorID, &b.Title, &b.Content, &b.Excerpt, &b.ThumbnailURL,
		&b.Privacy, &b.Status, &b.LikeCount, &b.DislikeCount, &b.CommentCount,
		&b.FeedScore, &b.ReadTimeMin,
		&b.TitleEn, &b.BodyEn, &b.TranslationStatus,
		&b.PublishedAt, &b.CreatedAt, &b.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return b, err
}

func (r *PostgresRepository) Update(ctx context.Context, b *Blog) error {
	_, err := r.db.Exec(ctx,
		`UPDATE blogs SET title=$2,content=$3,excerpt=$4,thumbnail_url=$5,privacy=$6,status=$7,
		                  read_time_min=$8,translation_status=$9,updated_at=$10
		 WHERE id=$1`,
		b.ID, b.Title, b.Content, b.Excerpt, b.ThumbnailURL, b.Privacy, b.Status,
		b.ReadTimeMin, b.TranslationStatus, time.Now(),
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

// UpsertCategories creates categories by name if missing (CR-002: name-based,
// like tags — the owner types category names in the editor).
func (r *PostgresRepository) UpsertCategories(ctx context.Context, names []string) ([]Category, error) {
	cats := make([]Category, 0, len(names))
	for _, name := range names {
		slug := slugify(name)
		var c Category
		err := r.db.QueryRow(ctx,
			`INSERT INTO categories (id, name, slug) VALUES ($1,$2,$3)
	         ON CONFLICT (slug) DO UPDATE SET name=EXCLUDED.name
	         RETURNING id, name, slug`,
			uuid.New(), name, slug,
		).Scan(&c.ID, &c.Name, &c.Slug)
		if err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

// SetBlogCategories replaces a blog's category associations.
func (r *PostgresRepository) SetBlogCategories(ctx context.Context, blogID uuid.UUID, categoryIDs []uuid.UUID) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM blog_categories WHERE blog_id=$1`, blogID); err != nil {
		return err
	}
	for _, catID := range categoryIDs {
		if _, err := r.db.Exec(ctx,
			`INSERT INTO blog_categories (blog_id, category_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			blogID, catID); err != nil {
			return err
		}
	}
	return nil
}

// ListCategories returns all categories with published-public article counts.
func (r *PostgresRepository) ListCategories(ctx context.Context) ([]CategoryWithCount, error) {
	rows, err := r.db.Query(ctx,
		`SELECT c.id, c.name, c.slug,
		        COUNT(b.id) FILTER (WHERE b.status='published' AND b.privacy='public')
	     FROM categories c
	     LEFT JOIN blog_categories bc ON bc.category_id = c.id
	     LEFT JOIN blogs b ON b.id = bc.blog_id
	     GROUP BY c.id, c.name, c.slug
	     ORDER BY c.name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []CategoryWithCount
	for rows.Next() {
		var c CategoryWithCount
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.BlogCount); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *PostgresRepository) UpdateTranslation(ctx context.Context, blogID uuid.UUID, titleEN, bodyEN, status string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE blogs SET title_en=$2, body_en=$3, translation_status=$4 WHERE id=$1`,
		blogID, nullableStr(titleEN), nullableStr(bodyEN), status,
	)
	return err
}

func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// GetArticlesFeed returns one page of published public blogs with denormalized
// author info and tags for feed cards (CR-001 / BUG-007 wire-up). An empty
// category matches everything; a slug filters via blog_categories (CR-002).
func (r *PostgresRepository) GetArticlesFeed(ctx context.Context, page, perPage int, category string) ([]*Blog, int, error) {
	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT b.id, b.author_id, u.username, COALESCE(u.avatar_url,''),
		        b.title, COALESCE(b.excerpt,''), COALESCE(b.thumbnail_url,''),
		        b.privacy, b.like_count, b.dislike_count, b.comment_count,
		        b.read_time_min, b.published_at,
		        COALESCE(b.title_en,''), b.translation_status,
		        COALESCE(
		          (SELECT json_agg(json_build_object('id', t.id, 'name', t.name, 'slug', t.slug))
		           FROM blog_tags bt JOIN tags t ON t.id = bt.tag_id
		           WHERE bt.blog_id = b.id),
		          '[]'::json
		        )
	     FROM blogs b JOIN users u ON u.id = b.author_id
	     WHERE b.status='published' AND b.privacy='public'
	       AND ($3 = '' OR EXISTS (
	            SELECT 1 FROM blog_categories bc JOIN categories c ON c.id = bc.category_id
	            WHERE bc.blog_id = b.id AND c.slug = $3))
	     ORDER BY b.feed_score DESC, b.published_at DESC
	     LIMIT $1 OFFSET $2`, perPage, offset, category,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var blogs []*Blog
	for rows.Next() {
		b := &Blog{}
		var tagsJSON []byte
		if err := rows.Scan(&b.ID, &b.AuthorID, &b.AuthorUsername, &b.AuthorAvatarURL,
			&b.Title, &b.Excerpt, &b.ThumbnailURL,
			&b.Privacy, &b.LikeCount, &b.DislikeCount, &b.CommentCount,
			&b.ReadTimeMin, &b.PublishedAt,
			&b.TitleEn, &b.TranslationStatus,
			&tagsJSON); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(tagsJSON, &b.Tags); err != nil {
			return nil, 0, err
		}
		blogs = append(blogs, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM blogs b
	     WHERE b.status='published' AND b.privacy='public'
	       AND ($1 = '' OR EXISTS (
	            SELECT 1 FROM blog_categories bc JOIN categories c ON c.id = bc.category_id
	            WHERE bc.blog_id = b.id AND c.slug = $1))`, category,
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	return blogs, total, nil
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

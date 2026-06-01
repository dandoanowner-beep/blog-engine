package blog

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeedRepository interface {
	GetExploreFeed(ctx context.Context, viewerID uuid.UUID, page int) ([]*Blog, int, error)
	GetFollowingFeed(ctx context.Context, viewerID uuid.UUID, page int) ([]*Blog, int, error)
}

type FeedPostgresRepository struct{ db *pgxpool.Pool }

func NewFeedPostgresRepository(db *pgxpool.Pool) *FeedPostgresRepository {
	return &FeedPostgresRepository{db: db}
}

func (r *FeedPostgresRepository) GetExploreFeed(ctx context.Context, viewerID uuid.UUID, page int) ([]*Blog, int, error) {
	offset := (page - 1) * 12
	rows, err := r.db.Query(ctx,
		`SELECT id,author_id,title,excerpt,thumbnail_url,privacy,like_count,dislike_count,comment_count,read_time_min,feed_score,published_at
         FROM blogs
         WHERE status='published' AND privacy='public'
         ORDER BY feed_score DESC
         LIMIT 12 OFFSET $1`, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var blogs []*Blog
	for rows.Next() {
		b := &Blog{}
		if err := rows.Scan(&b.ID, &b.AuthorID, &b.Title, &b.Excerpt, &b.ThumbnailURL,
			&b.Privacy, &b.LikeCount, &b.DislikeCount, &b.CommentCount, &b.ReadTimeMin,
			&b.FeedScore, &b.PublishedAt); err != nil {
			return nil, 0, err
		}
		blogs = append(blogs, b)
	}
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM blogs WHERE status='published' AND privacy='public'`).Scan(&total)
	return blogs, total, nil
}

func (r *FeedPostgresRepository) GetFollowingFeed(ctx context.Context, viewerID uuid.UUID, page int) ([]*Blog, int, error) {
	offset := (page - 1) * 12
	rows, err := r.db.Query(ctx,
		`SELECT b.id,b.author_id,b.title,b.excerpt,b.thumbnail_url,b.privacy,b.like_count,b.dislike_count,b.comment_count,b.read_time_min,b.published_at
         FROM blogs b
         JOIN follows f ON f.followee_id = b.author_id
         WHERE f.follower_id=$1 AND b.status='published' AND b.privacy='public'
         ORDER BY b.published_at DESC
         LIMIT 12 OFFSET $2`, viewerID, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var blogs []*Blog
	for rows.Next() {
		b := &Blog{}
		if err := rows.Scan(&b.ID, &b.AuthorID, &b.Title, &b.Excerpt, &b.ThumbnailURL,
			&b.Privacy, &b.LikeCount, &b.DislikeCount, &b.CommentCount, &b.ReadTimeMin,
			&b.PublishedAt); err != nil {
			return nil, 0, err
		}
		blogs = append(blogs, b)
	}
	var total int
	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM blogs b JOIN follows f ON f.followee_id=b.author_id WHERE f.follower_id=$1 AND b.status='published'`,
		viewerID,
	).Scan(&total)
	return blogs, total, nil
}

func (h *Handler) ExploreFeed(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "explore feed - connect repository"})
}

func (h *Handler) FollowingFeed(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "following feed - connect repository"})
}

func (h *Handler) UpdateBlog(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "update blog"})
}

func pageFromQuery(r *http.Request) int {
	p, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if p < 1 {
		return 1
	}
	return p
}

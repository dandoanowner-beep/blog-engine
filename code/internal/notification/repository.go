package notification

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, n *Notification) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO notifications (id,user_id,type,actor_id,blog_id,comment_id,read,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		n.ID, n.UserID, n.Type, n.ActorID, n.BlogID, n.CommentID, n.Read, n.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) ListForUser(ctx context.Context, userID uuid.UUID, page int) ([]*Notification, int, error) {
	offset := (page - 1) * 20
	rows, err := r.db.Query(ctx,
		`SELECT id,user_id,type,actor_id,blog_id,comment_id,read,created_at FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT 20 OFFSET $2`,
		userID, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Notification
	for rows.Next() {
		n := &Notification{}
		rows.Scan(&n.ID, &n.UserID, &n.Type, &n.ActorID, &n.BlogID, &n.CommentID, &n.Read, &n.CreatedAt)
		items = append(items, n)
	}
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id=$1`, userID).Scan(&total)
	return items, total, nil
}

func (r *PostgresRepository) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE notifications SET read=true WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func (r *PostgresRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE notifications SET read=true WHERE user_id=$1`, userID)
	return err
}

func (r *PostgresRepository) GetModsAndAdmins(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `SELECT id FROM users WHERE role IN ('moderator','admin','owner')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids, nil
}

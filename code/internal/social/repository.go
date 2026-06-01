package social

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

func (r *PostgresRepository) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `INSERT INTO follows (follower_id,followee_id) VALUES ($1,$2)`, followerID, followeeID)
	return err
}

func (r *PostgresRepository) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM follows WHERE follower_id=$1 AND followee_id=$2`, followerID, followeeID)
	return err
}

func (r *PostgresRepository) IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id=$1 AND followee_id=$2)`,
		followerID, followeeID,
	).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) CreateFriendRequest(ctx context.Context, senderID, receiverID uuid.UUID) (*FriendRequest, error) {
	req := &FriendRequest{ID: uuid.New(), SenderID: senderID, ReceiverID: receiverID, Status: "pending", CreatedAt: time.Now()}
	_, err := r.db.Exec(ctx,
		`INSERT INTO friend_requests (id,sender_id,receiver_id,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$5)`,
		req.ID, req.SenderID, req.ReceiverID, req.Status, req.CreatedAt,
	)
	return req, err
}

func (r *PostgresRepository) GetFriendRequest(ctx context.Context, id uuid.UUID) (*FriendRequest, error) {
	req := &FriendRequest{}
	err := r.db.QueryRow(ctx,
		`SELECT id,sender_id,receiver_id,status FROM friend_requests WHERE id=$1`, id,
	).Scan(&req.ID, &req.SenderID, &req.ReceiverID, &req.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return req, err
}

func (r *PostgresRepository) GetPendingRequest(ctx context.Context, senderID, receiverID uuid.UUID) (*FriendRequest, error) {
	req := &FriendRequest{}
	err := r.db.QueryRow(ctx,
		`SELECT id,sender_id,receiver_id,status FROM friend_requests WHERE sender_id=$1 AND receiver_id=$2 AND status='pending'`,
		senderID, receiverID,
	).Scan(&req.ID, &req.SenderID, &req.ReceiverID, &req.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return req, err
}

func (r *PostgresRepository) UpdateFriendRequest(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE friend_requests SET status=$2, updated_at=NOW() WHERE id=$1`, id, status)
	return err
}

func (r *PostgresRepository) CreateFriendship(ctx context.Context, userA, userB uuid.UUID) error {
	a, b := orderUUIDs(userA, userB)
	_, err := r.db.Exec(ctx, `INSERT INTO friends (user_id_1, user_id_2) VALUES ($1,$2) ON CONFLICT DO NOTHING`, a, b)
	return err
}

func (r *PostgresRepository) DeleteFriendship(ctx context.Context, userA, userB uuid.UUID) error {
	a, b := orderUUIDs(userA, userB)
	_, err := r.db.Exec(ctx, `DELETE FROM friends WHERE user_id_1=$1 AND user_id_2=$2`, a, b)
	return err
}

func (r *PostgresRepository) UpsertReaction(ctx context.Context, reaction *Reaction) (int, int, error) {
	_, err := r.db.Exec(ctx,
		`INSERT INTO reactions (user_id,blog_id,type) VALUES ($1,$2,$3)
         ON CONFLICT (user_id,blog_id) DO UPDATE SET type=EXCLUDED.type`,
		reaction.UserID, reaction.BlogID, reaction.Type,
	)
	if err != nil {
		return 0, 0, err
	}
	return r.getCounts(ctx, reaction.BlogID)
}

func (r *PostgresRepository) DeleteReaction(ctx context.Context, userID, blogID uuid.UUID) (int, int, error) {
	_, err := r.db.Exec(ctx, `DELETE FROM reactions WHERE user_id=$1 AND blog_id=$2`, userID, blogID)
	if err != nil {
		return 0, 0, err
	}
	return r.getCounts(ctx, blogID)
}

func (r *PostgresRepository) GetReaction(ctx context.Context, userID, blogID uuid.UUID) (*Reaction, error) {
	reaction := &Reaction{}
	err := r.db.QueryRow(ctx, `SELECT user_id,blog_id,type FROM reactions WHERE user_id=$1 AND blog_id=$2`, userID, blogID).
		Scan(&reaction.UserID, &reaction.BlogID, &reaction.Type)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return reaction, err
}

func (r *PostgresRepository) getCounts(ctx context.Context, blogID uuid.UUID) (int, int, error) {
	var likes, dislikes int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM reactions WHERE blog_id=$1 AND type='like'`, blogID).Scan(&likes)
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM reactions WHERE blog_id=$1 AND type='dislike'`, blogID).Scan(&dislikes)
	return likes, dislikes, nil
}

func (r *PostgresRepository) CreateComment(ctx context.Context, c *Comment) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO comments (id,blog_id,author_id,parent_id,content,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$6)`,
		c.ID, c.BlogID, c.AuthorID, c.ParentID, c.Content, c.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) GetComment(ctx context.Context, id uuid.UUID) (*Comment, error) {
	c := &Comment{}
	err := r.db.QueryRow(ctx, `SELECT id,blog_id,author_id,parent_id,content FROM comments WHERE id=$1`, id).
		Scan(&c.ID, &c.BlogID, &c.AuthorID, &c.ParentID, &c.Content)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *PostgresRepository) DeleteComment(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM comments WHERE id=$1`, id)
	return err
}

func (r *PostgresRepository) CreateReport(ctx context.Context, report *Report) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO reports (id,reporter_id,blog_id,comment_id,reason,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		report.ID, report.ReporterID, nullUUID(report.BlogID), nullUUID(report.CommentID), report.Reason, report.Status, report.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) ReportExists(ctx context.Context, reporterID, blogID, commentID uuid.UUID) (bool, error) {
	var exists bool
	if blogID != uuid.Nil {
		r.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM reports WHERE reporter_id=$1 AND blog_id=$2)`, reporterID, blogID,
		).Scan(&exists)
	} else {
		r.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM reports WHERE reporter_id=$1 AND comment_id=$2)`, reporterID, commentID,
		).Scan(&exists)
	}
	return exists, nil
}

func orderUUIDs(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if a.String() < b.String() {
		return a, b
	}
	return b, a
}

func nullUUID(id uuid.UUID) interface{} {
	if id == uuid.Nil {
		return nil
	}
	return id
}

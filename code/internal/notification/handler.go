package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"blog-engine/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotifService interface {
	MarkRead(ctx context.Context, notifID, userID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
}

type Handler struct {
	svc  NotifService
	repo Repository
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, repo: svc.repo}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	items, total, err := h.repo.ListForUser(r.Context(), userID, page)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed")
		return
	}
	unread := 0
	for _, n := range items {
		if !n.Read {
			unread++
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": items,
		"unread_count":  unread,
		"total":         total,
		"page":          page,
	})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	notifID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.MarkRead(r.Context(), notifID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"read": true})
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if err := h.svc.MarkAllRead(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "All marked as read"})
}

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

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

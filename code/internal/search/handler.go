package search

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"blog-engine/internal/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	viewerID := middleware.UserIDFromContext(r.Context())
	result, err := h.svc.Search(r.Context(), q, viewerID, page)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

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

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

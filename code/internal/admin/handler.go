package admin

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

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	role := r.URL.Query().Get("role")
	users, total, err := h.svc.ListUsers(r.Context(), page, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"users": users, "total": total, "page": page})
}

func (h *Handler) ChangeRole(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var body struct {
		Role string `json:"role"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if err := h.svc.ChangeRole(r.Context(), userID, body.Role); err != nil {
		switch err {
		case ErrInvalidRole, ErrCannotAssignOwner:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed")
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "role updated"})
}

func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}
	reports, total, err := h.svc.ListReports(r.Context(), status, page)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"reports": reports, "total": total, "page": page})
}

func (h *Handler) ResolveReport(w http.ResponseWriter, r *http.Request) {
	reportID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid report id")
		return
	}
	resolverID := middleware.UserIDFromContext(r.Context())
	var body struct {
		Action string `json:"action"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if err := h.svc.ResolveReport(r.Context(), reportID, body.Action, resolverID); err != nil {
		switch err {
		case ErrInvalidReportAction:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed")
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "resolved"})
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

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

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

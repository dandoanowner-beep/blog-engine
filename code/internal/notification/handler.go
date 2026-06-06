package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"blog-engine/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

package site

import (
	"context"
	"encoding/json"
	"net/http"
)

type ContentService interface {
	GetAbout(ctx context.Context) (string, error)
	UpdateAbout(ctx context.Context, content string) error
}

type Handler struct {
	svc ContentService
}

func NewHandler(svc ContentService) *Handler {
	return &Handler{svc: svc}
}

// GetAbout is public — the Author page is visible to everyone.
func (h *Handler) GetAbout(w http.ResponseWriter, r *http.Request) {
	content, err := h.svc.GetAbout(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load content")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"content": content})
}

// UpdateAbout is owner-only (enforced at routing level, CR-001 pattern).
func (h *Handler) UpdateAbout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateAbout(r.Context(), req.Content); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save content")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

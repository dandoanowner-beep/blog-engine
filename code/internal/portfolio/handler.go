package portfolio

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProjectService interface {
	Create(ctx context.Context, input CreateInput) (*Project, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*Project, error)
}

type Handler struct {
	svc ProjectService
}

func NewHandler(svc ProjectService) *Handler {
	return &Handler{svc: svc}
}

func projectJSON(p *Project) map[string]interface{} {
	return map[string]interface{}{
		"id":            p.ID,
		"title":         p.Title,
		"description":   p.Description,
		"tech_stack":    p.TechStack,
		"repo_url":      p.RepoURL,
		"demo_url":      p.DemoURL,
		"thumbnail_url": p.ThumbnailURL,
		"sort_order":    p.SortOrder,
		"created_at":    p.CreatedAt,
		"updated_at":    p.UpdatedAt,
	}
}

// ListProjects is public — the portfolio page is visible to everyone.
func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load projects")
		return
	}
	out := make([]map[string]interface{}, 0, len(projects))
	for _, p := range projects {
		out = append(out, projectJSON(p))
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"projects": out})
}

type projectRequest struct {
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	TechStack    *string `json:"tech_stack"`
	RepoURL      *string `json:"repo_url"`
	DemoURL      *string `json:"demo_url"`
	ThumbnailURL *string `json:"thumbnail_url"`
	SortOrder    *int    `json:"sort_order"`
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// CreateProject is owner-only (enforced at routing level, CR-001 pattern).
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req projectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	input := CreateInput{
		Title:        deref(req.Title),
		Description:  deref(req.Description),
		TechStack:    deref(req.TechStack),
		RepoURL:      deref(req.RepoURL),
		DemoURL:      deref(req.DemoURL),
		ThumbnailURL: deref(req.ThumbnailURL),
	}
	if req.SortOrder != nil {
		input.SortOrder = *req.SortOrder
	}
	p, err := h.svc.Create(r.Context(), input)
	if err != nil {
		switch err {
		case ErrMissingTitle:
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to create project")
		}
		return
	}
	writeJSON(w, http.StatusCreated, projectJSON(p))
}

// UpdateProject is owner-only (enforced at routing level).
func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	var req projectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	input := UpdateInput{
		Title:        req.Title,
		Description:  req.Description,
		TechStack:    req.TechStack,
		RepoURL:      req.RepoURL,
		DemoURL:      req.DemoURL,
		ThumbnailURL: req.ThumbnailURL,
		SortOrder:    req.SortOrder,
	}
	p, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		switch err {
		case ErrNotFound:
			writeError(w, http.StatusNotFound, "project not found")
		case ErrMissingTitle:
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to update project")
		}
		return
	}
	writeJSON(w, http.StatusOK, projectJSON(p))
}

// DeleteProject is owner-only (enforced at routing level).
func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete project")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

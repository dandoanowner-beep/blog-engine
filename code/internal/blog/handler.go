package blog

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	CtxUserID contextKey = "user_id"
	CtxRole   contextKey = "role"
)

type BlogService interface {
	Create(ctx context.Context, input CreateInput) (*Blog, error)
	GetForViewer(ctx context.Context, blogID, viewerID uuid.UUID) (*Blog, bool, error)
	Delete(ctx context.Context, blogID, requesterID uuid.UUID, role string) error
}

type Handler struct {
	svc BlogService
}

func NewHandler(svc BlogService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetBlog(w http.ResponseWriter, r *http.Request) {
	blogID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid blog id")
		return
	}
	viewerID, _ := r.Context().Value(CtxUserID).(uuid.UUID)

	b, partial, err := h.svc.GetForViewer(r.Context(), blogID, viewerID)
	if err != nil {
		switch err {
		case ErrNotFound:
			writeError(w, http.StatusNotFound, "blog not found")
		case ErrAccessDenied:
			writeError(w, http.StatusForbidden, "access denied")
		default:
			writeError(w, http.StatusInternalServerError, "failed to fetch blog")
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":      b.ID,
		"title":   b.Title,
		"content": b.Content,
		"privacy": b.Privacy,
		"partial": partial,
	})
}

func (h *Handler) CreateBlog(w http.ResponseWriter, r *http.Request) {
	authorID, _ := r.Context().Value(CtxUserID).(uuid.UUID)
	if authorID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Title       string      `json:"title"`
		Content     string      `json:"content"`
		ThumbnailURL string     `json:"thumbnail_url"`
		Privacy     Privacy     `json:"privacy"`
		Status      Status      `json:"status"`
		TagNames    []string    `json:"tag_names"`
		CategoryIDs []uuid.UUID `json:"category_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	input := CreateInput{
		AuthorID:    authorID,
		Title:       req.Title,
		Content:     req.Content,
		ThumbnailURL: req.ThumbnailURL,
		Privacy:     req.Privacy,
		Status:      req.Status,
		TagNames:    req.TagNames,
		CategoryIDs: req.CategoryIDs,
	}
	b, err := h.svc.Create(r.Context(), input)
	if err != nil {
		switch err {
		case ErrMissingTitle, ErrMissingTags:
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to create blog")
		}
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) DeleteBlog(w http.ResponseWriter, r *http.Request) {
	blogID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid blog id")
		return
	}
	requesterID, _ := r.Context().Value(CtxUserID).(uuid.UUID)
	role, _ := r.Context().Value(CtxRole).(string)

	if err := h.svc.Delete(r.Context(), blogID, requesterID, role); err != nil {
		switch err {
		case ErrForbidden:
			writeError(w, http.StatusForbidden, "forbidden")
		case ErrNotFound:
			writeError(w, http.StatusNotFound, "blog not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete blog")
		}
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

package blog

import (
	"context"
	"encoding/json"
	"net/http"

	"blog-engine/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// CtxUserID and CtxRole are kept for handler tests that inject context values directly.
type contextKey string

const (
	CtxUserID contextKey = "user_id"
	CtxRole   contextKey = "role"
)

type BlogService interface {
	Create(ctx context.Context, input CreateInput) (*Blog, error)
	GetForViewer(ctx context.Context, blogID, viewerID uuid.UUID) (*Blog, bool, error)
	Update(ctx context.Context, blogID uuid.UUID, input UpdateInput) (*Blog, error)
	Delete(ctx context.Context, blogID, requesterID uuid.UUID, role string) error
	ArticlesFeed(ctx context.Context, page int, category string) ([]*Blog, int, error)
	ListCategories(ctx context.Context) ([]CategoryWithCount, error)
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
	viewerID := middleware.UserIDFromContext(r.Context())

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
		"id":                 b.ID,
		"title":              b.Title,
		"content":            b.Content,
		"privacy":            b.Privacy,
		"partial":            partial,
		"title_en":           b.TitleEn,
		"body_en":            b.BodyEn,
		"translation_status": b.TranslationStatus,
	})
}

func (h *Handler) CreateBlog(w http.ResponseWriter, r *http.Request) {
	authorID := middleware.UserIDFromContext(r.Context())
	if authorID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Title         string   `json:"title"`
		Content       string   `json:"content"`
		ThumbnailURL  string   `json:"thumbnail_url"`
		Privacy       Privacy  `json:"privacy"`
		Status        Status   `json:"status"`
		TagNames      []string `json:"tag_names"`
		CategoryNames []string `json:"category_names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	input := CreateInput{
		AuthorID:      authorID,
		Title:         req.Title,
		Content:       req.Content,
		ThumbnailURL:  req.ThumbnailURL,
		Privacy:       req.Privacy,
		Status:        req.Status,
		TagNames:      req.TagNames,
		CategoryNames: req.CategoryNames,
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

func (h *Handler) UpdateBlog(w http.ResponseWriter, r *http.Request) {
	blogID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid blog id")
		return
	}
	requesterID := middleware.UserIDFromContext(r.Context())
	if requesterID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Title         *string  `json:"title"`
		Content       *string  `json:"content"`
		ThumbnailURL  *string  `json:"thumbnail_url"`
		Privacy       *Privacy `json:"privacy"`
		Status        *Status  `json:"status"`
		TagNames      []string `json:"tag_names"`
		CategoryNames []string `json:"category_names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	input := UpdateInput{
		RequesterID:   requesterID,
		Title:         req.Title,
		Content:       req.Content,
		ThumbnailURL:  req.ThumbnailURL,
		Privacy:       req.Privacy,
		Status:        req.Status,
		TagNames:      req.TagNames,
		CategoryNames: req.CategoryNames,
	}
	b, err := h.svc.Update(r.Context(), blogID, input)
	if err != nil {
		switch err {
		case ErrForbidden:
			writeError(w, http.StatusForbidden, "forbidden")
		case ErrNotFound:
			writeError(w, http.StatusNotFound, "blog not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update blog")
		}
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) DeleteBlog(w http.ResponseWriter, r *http.Request) {
	blogID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid blog id")
		return
	}
	requesterID := middleware.UserIDFromContext(r.Context())
	role := middleware.RoleFromContext(r.Context())

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

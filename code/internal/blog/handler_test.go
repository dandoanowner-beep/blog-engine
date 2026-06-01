package blog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-engine/internal/blog"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- mock blog service ---

type mockBlogService struct{ mock.Mock }

func (m *mockBlogService) Create(ctx context.Context, input blog.CreateInput) (*blog.Blog, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*blog.Blog), args.Error(1)
}
func (m *mockBlogService) GetForViewer(ctx context.Context, blogID, viewerID uuid.UUID) (*blog.Blog, bool, error) {
	args := m.Called(ctx, blogID, viewerID)
	if args.Get(0) == nil {
		return nil, false, args.Error(2)
	}
	return args.Get(0).(*blog.Blog), args.Bool(1), args.Error(2)
}
func (m *mockBlogService) Delete(ctx context.Context, blogID, requesterID uuid.UUID, role string) error {
	return m.Called(ctx, blogID, requesterID, role).Error(0)
}

// ════════════════════════════════════════════════════════════
// Blog handler tests
// ════════════════════════════════════════════════════════════

func TestGetBlogHandler_PublicBlog_Returns200(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)

	blogID := uuid.New()
	b := &blog.Blog{ID: blogID, Title: "Hello World", Content: "full content", Privacy: blog.PrivacyPublic}
	svc.On("GetForViewer", mock.Anything, blogID, uuid.Nil).Return(b, false, nil)

	r := chi.NewRouter()
	r.Get("/blogs/{id}", h.GetBlog)

	req := httptest.NewRequest(http.MethodGet, "/blogs/"+blogID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "Hello World", resp["title"])
}

func TestGetBlogHandler_NotFound_Returns404(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)

	blogID := uuid.New()
	svc.On("GetForViewer", mock.Anything, blogID, uuid.Nil).Return(nil, false, blog.ErrNotFound)

	r := chi.NewRouter()
	r.Get("/blogs/{id}", h.GetBlog)

	req := httptest.NewRequest(http.MethodGet, "/blogs/"+blogID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetBlogHandler_AccessDenied_Returns403(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)

	blogID := uuid.New()
	svc.On("GetForViewer", mock.Anything, blogID, uuid.Nil).Return(nil, false, blog.ErrAccessDenied)

	r := chi.NewRouter()
	r.Get("/blogs/{id}", h.GetBlog)

	req := httptest.NewRequest(http.MethodGet, "/blogs/"+blogID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestGetBlogHandler_GuestGetPartialBlog_Returns200WithPartialFlag(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)

	blogID := uuid.New()
	b := &blog.Blog{ID: blogID, Title: "Partial", Content: "some content", Privacy: blog.PrivacyPublic}
	svc.On("GetForViewer", mock.Anything, blogID, uuid.Nil).Return(b, true, nil)

	r := chi.NewRouter()
	r.Get("/blogs/{id}", h.GetBlog)

	req := httptest.NewRequest(http.MethodGet, "/blogs/"+blogID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, true, resp["partial"])
}

func TestDeleteBlogHandler_Returns204(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)

	authorID := uuid.New()
	blogID := uuid.New()
	svc.On("Delete", mock.Anything, blogID, authorID, "user").Return(nil)

	r := chi.NewRouter()
	r.Delete("/blogs/{id}", func(w http.ResponseWriter, r *http.Request) {
		// inject auth context
		ctx := context.WithValue(r.Context(), blog.CtxUserID, authorID)
		ctx = context.WithValue(ctx, blog.CtxRole, "user")
		h.DeleteBlog(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest(http.MethodDelete, "/blogs/"+blogID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestCreateBlogHandler_MissingTitle_Returns422(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)

	authorID := uuid.New()
	svc.On("Create", mock.Anything, mock.Anything).Return(nil, blog.ErrMissingTitle)

	body := `{"title":"","content":"some","tag_names":["go"],"category_ids":[],"privacy":"public","status":"published"}`
	r := chi.NewRouter()
	r.Post("/blogs", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), blog.CtxUserID, authorID)
		ctx = context.WithValue(ctx, blog.CtxRole, "user")
		h.CreateBlog(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest(http.MethodPost, "/blogs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

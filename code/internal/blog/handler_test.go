package blog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-engine/internal/blog"
	"blog-engine/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- mock blog service ---

var _ blog.BlogService = (*mockBlogService)(nil)

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
func (m *mockBlogService) Update(ctx context.Context, blogID uuid.UUID, input blog.UpdateInput) (*blog.Blog, error) {
	args := m.Called(ctx, blogID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*blog.Blog), args.Error(1)
}
func (m *mockBlogService) Delete(ctx context.Context, blogID, requesterID uuid.UUID, role string) error {
	return m.Called(ctx, blogID, requesterID, role).Error(0)
}
func (m *mockBlogService) ArticlesFeed(ctx context.Context, page int, category string) ([]*blog.Blog, int, error) {
	args := m.Called(ctx, page, category)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*blog.Blog), args.Int(1), args.Error(2)
}
func (m *mockBlogService) ListCategories(ctx context.Context) ([]blog.CategoryWithCount, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]blog.CategoryWithCount), args.Error(1)
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
	assert.NotNil(t, resp["translation_status"])
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
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
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
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
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
		ctx := context.WithValue(r.Context(), middleware.ContextKey("user_id"), authorID)
		ctx = context.WithValue(ctx, middleware.ContextKey("role"), "user")
		h.DeleteBlog(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest(http.MethodDelete, "/blogs/"+blogID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// ════════════════════════════════════════════════════════════
// PATCH /blogs/:id — UpdateBlog handler
// ════════════════════════════════════════════════════════════

func TestUpdateBlogHandler_Success_Returns200(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)

	authorID := uuid.New()
	blogID := uuid.New()
	newTitle := "Updated Title"
	updated := &blog.Blog{ID: blogID, Title: "Updated Title", Privacy: blog.PrivacyPublic}

	svc.On("Update", mock.Anything, blogID, mock.MatchedBy(func(inp blog.UpdateInput) bool {
		return inp.RequesterID == authorID && inp.Title != nil && *inp.Title == "Updated Title"
	})).Return(updated, nil)

	r := chi.NewRouter()
	r.Patch("/blogs/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), middleware.ContextKey("user_id"), authorID)
		h.UpdateBlog(w, req.WithContext(ctx))
	})

	body, _ := json.Marshal(map[string]interface{}{"title": newTitle})
	req := httptest.NewRequest(http.MethodPatch, "/blogs/"+blogID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp)
}

func TestUpdateBlogHandler_ForbiddenForNonAuthor_Returns403(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)

	otherID := uuid.New()
	blogID := uuid.New()
	svc.On("Update", mock.Anything, blogID, mock.Anything).Return(nil, blog.ErrForbidden)

	r := chi.NewRouter()
	r.Patch("/blogs/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), middleware.ContextKey("user_id"), otherID)
		h.UpdateBlog(w, req.WithContext(ctx))
	})

	body, _ := json.Marshal(map[string]string{"title": "hacked"})
	req := httptest.NewRequest(http.MethodPatch, "/blogs/"+blogID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}

func TestCreateBlogHandler_MissingTitle_Returns422(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)

	authorID := uuid.New()
	svc.On("Create", mock.Anything, mock.Anything).Return(nil, blog.ErrMissingTitle)

	body := `{"title":"","content":"some","tag_names":["go"],"category_ids":[],"privacy":"public","status":"published"}`
	r := chi.NewRouter()
	r.Post("/blogs", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.ContextKey("user_id"), authorID)
		ctx = context.WithValue(ctx, middleware.ContextKey("role"), "user")
		h.CreateBlog(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest(http.MethodPost, "/blogs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}

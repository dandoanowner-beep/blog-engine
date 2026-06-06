package search_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-engine/internal/search"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ search.Repository = (*mockSearchRepo)(nil)

type mockSearchRepo struct{ mock.Mock }

func (m *mockSearchRepo) SearchBlogs(ctx context.Context, q string, viewerID uuid.UUID, page int) ([]*search.BlogResult, int, error) {
	args := m.Called(ctx, q, viewerID, page)
	return args.Get(0).([]*search.BlogResult), args.Int(1), args.Error(2)
}
func (m *mockSearchRepo) SearchUsers(ctx context.Context, q string, page int) ([]*search.UserResult, int, error) {
	args := m.Called(ctx, q, page)
	return args.Get(0).([]*search.UserResult), args.Int(1), args.Error(2)
}
func (m *mockSearchRepo) SearchTags(ctx context.Context, q string, page int) ([]*search.TagResult, int, error) {
	args := m.Called(ctx, q, page)
	return args.Get(0).([]*search.TagResult), args.Int(1), args.Error(2)
}

func TestSearchHandler_ReturnsResults(t *testing.T) {
	repo := &mockSearchRepo{}
	svc := search.NewService(repo)
	h := search.NewHandler(svc)

	blogs := []*search.BlogResult{{ID: uuid.New(), Title: "Go tips"}}
	repo.On("SearchBlogs", mock.Anything, "go", uuid.Nil, 1).Return(blogs, 1, nil)
	repo.On("SearchUsers", mock.Anything, "go", 1).Return([]*search.UserResult{}, 0, nil)
	repo.On("SearchTags", mock.Anything, "go", 1).Return([]*search.TagResult{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/search?q=go&page=1", nil)
	rec := httptest.NewRecorder()
	h.Search(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "go", resp["Query"])
	assert.NotNil(t, resp["Blogs"])
}

func TestSearchHandler_EmptyQuery_ReturnsEmpty(t *testing.T) {
	repo := &mockSearchRepo{}
	svc := search.NewService(repo)
	h := search.NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/search?q=", nil)
	rec := httptest.NewRecorder()
	h.Search(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "", resp["Query"])
}

func TestSearchHandler_RepoError_Returns500(t *testing.T) {
	repo := &mockSearchRepo{}
	svc := search.NewService(repo)
	h := search.NewHandler(svc)

	repo.On("SearchBlogs", mock.Anything, "fail", uuid.Nil, 1).Return([]*search.BlogResult{}, 0, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/search?q=fail", nil)
	rec := httptest.NewRecorder()
	h.Search(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}

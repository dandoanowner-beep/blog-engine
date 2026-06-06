package blog_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-engine/internal/blog"
	"github.com/stretchr/testify/assert"
)

func TestExploreFeedHandler_Returns200(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/feed/explore", nil)
	rec := httptest.NewRecorder()
	h.ExploreFeed(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotNil(t, resp["message"])
}

func TestFollowingFeedHandler_Returns200(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/feed/following", nil)
	rec := httptest.NewRecorder()
	h.FollowingFeed(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotNil(t, resp["message"])
}

func TestUpdateBlogHandler_Returns200(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/blogs/1", nil)
	rec := httptest.NewRecorder()
	h.UpdateBlog(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

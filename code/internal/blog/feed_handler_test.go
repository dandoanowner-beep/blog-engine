package blog_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-engine/internal/blog"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// CR-001 / BUG-007: the articles feed is wired to the repository and must
// return the BlogCard contract the frontend consumes.
func TestArticlesFeedHandler_ReturnsBlogCardContract(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)

	authorID := uuid.New()
	feed := []*blog.Blog{{
		ID:                uuid.New(),
		AuthorID:          authorID,
		AuthorUsername:    "chubeunu",
		Title:             "Bài viết đầu tiên",
		Excerpt:           "Mở đầu...",
		ReadTimeMin:       3,
		LikeCount:         5,
		CommentCount:      2,
		Privacy:           blog.PrivacyPublic,
		Tags:              []blog.Tag{{ID: uuid.New(), Name: "travel", Slug: "travel"}},
		TitleEn:           "First Post",
		TranslationStatus: blog.TranslationStatusDone,
	}}
	svc.On("ArticlesFeed", mock.Anything, 1, "").Return(feed, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/blogs/feed", nil)
	rec := httptest.NewRecorder()
	h.ArticlesFeed(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Blogs []struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Author struct {
				ID       string `json:"id"`
				Username string `json:"username"`
			} `json:"author"`
			Tags []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"tags"`
			LikeCount         int    `json:"like_count"`
			TitleEn           string `json:"title_en"`
			TranslationStatus string `json:"translation_status"`
		} `json:"blogs"`
		Total   int `json:"total"`
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	}
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Len(t, resp.Blogs, 1)
	assert.Equal(t, "chubeunu", resp.Blogs[0].Author.Username)
	assert.Equal(t, authorID.String(), resp.Blogs[0].Author.ID)
	assert.Equal(t, "travel", resp.Blogs[0].Tags[0].Name)
	assert.Equal(t, 5, resp.Blogs[0].LikeCount)
	assert.Equal(t, "First Post", resp.Blogs[0].TitleEn)
	assert.Equal(t, 1, resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, blog.ArticlesPerPage, resp.PerPage)
}

func TestArticlesFeedHandler_EmptyFeed_ReturnsEmptyArrayNotNull(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)
	svc.On("ArticlesFeed", mock.Anything, 1, "").Return([]*blog.Blog{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/blogs/feed", nil)
	rec := httptest.NewRecorder()
	h.ArticlesFeed(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// frontend does data.blogs.map — null would crash it
	assert.Contains(t, rec.Body.String(), `"blogs":[]`)
}

func TestArticlesFeedHandler_PageFromQuery(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)
	svc.On("ArticlesFeed", mock.Anything, 3, "").Return([]*blog.Blog{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/blogs/feed?page=3", nil)
	rec := httptest.NewRecorder()
	h.ArticlesFeed(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertCalled(t, "ArticlesFeed", mock.Anything, 3, "")
}

// CR-002: ?category=slug filters the feed (Categories browse page)
func TestArticlesFeedHandler_CategoryParam(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)
	svc.On("ArticlesFeed", mock.Anything, 1, "tutorials").Return([]*blog.Blog{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/blogs/feed?category=tutorials", nil)
	rec := httptest.NewRecorder()
	h.ArticlesFeed(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertCalled(t, "ArticlesFeed", mock.Anything, 1, "tutorials")
}

// CR-002: public categories list with article counts
func TestListCategoriesHandler_ReturnsContract(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)
	svc.On("ListCategories", mock.Anything).Return([]blog.CategoryWithCount{
		{Category: blog.Category{ID: uuid.New(), Name: "Tutorials", Slug: "tutorials"}, BlogCount: 4},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	rec := httptest.NewRecorder()
	h.ListCategories(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Categories []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Slug      string `json:"slug"`
			BlogCount int    `json:"blog_count"`
		} `json:"categories"`
	}
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Len(t, resp.Categories, 1)
	assert.Equal(t, "tutorials", resp.Categories[0].Slug)
	assert.Equal(t, 4, resp.Categories[0].BlogCount)
}

func TestListCategoriesHandler_Empty_ReturnsEmptyArrayNotNull(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)
	svc.On("ListCategories", mock.Anything).Return([]blog.CategoryWithCount{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	rec := httptest.NewRecorder()
	h.ListCategories(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"categories":[]`)
}

func TestArticlesFeedHandler_ServiceError_Returns500(t *testing.T) {
	svc := &mockBlogService{}
	h := blog.NewHandler(svc)
	svc.On("ArticlesFeed", mock.Anything, 1, "").Return(nil, 0, errors.New("db down"))

	req := httptest.NewRequest(http.MethodGet, "/blogs/feed", nil)
	rec := httptest.NewRecorder()
	h.ArticlesFeed(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

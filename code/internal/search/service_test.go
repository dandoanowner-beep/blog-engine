package search_test

import (
	"context"
	"testing"

	"blog-engine/internal/search"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock ---

type mockRepo struct{ mock.Mock }

func (m *mockRepo) SearchBlogs(ctx context.Context, q string, viewerID uuid.UUID, page int) ([]*search.BlogResult, int, error) {
	args := m.Called(ctx, q, viewerID, page)
	return args.Get(0).([]*search.BlogResult), args.Int(1), args.Error(2)
}
func (m *mockRepo) SearchUsers(ctx context.Context, q string, page int) ([]*search.UserResult, int, error) {
	args := m.Called(ctx, q, page)
	return args.Get(0).([]*search.UserResult), args.Int(1), args.Error(2)
}
func (m *mockRepo) SearchTags(ctx context.Context, q string, page int) ([]*search.TagResult, int, error) {
	args := m.Called(ctx, q, page)
	return args.Get(0).([]*search.TagResult), args.Int(1), args.Error(2)
}

// ════════════════════════════════════════════════════════════
// AC-SEARCH-001
// ════════════════════════════════════════════════════════════

func TestSearch_ReturnsBlogsUsersAndTags(t *testing.T) {
	repo := &mockRepo{}
	svc := search.NewService(repo)

	viewerID := uuid.New()
	blogs := []*search.BlogResult{{ID: uuid.New(), Title: "Go Tutorial"}}
	users := []*search.UserResult{{ID: uuid.New(), Username: "gopher"}}
	tags := []*search.TagResult{{Name: "go"}}

	repo.On("SearchBlogs", mock.Anything, "go", viewerID, 1).Return(blogs, 1, nil)
	repo.On("SearchUsers", mock.Anything, "go", 1).Return(users, 1, nil)
	repo.On("SearchTags", mock.Anything, "go", 1).Return(tags, 1, nil)

	result, err := svc.Search(context.Background(), "go", viewerID, 1)
	assert.NoError(t, err)
	assert.Len(t, result.Blogs.Items, 1)
	assert.Len(t, result.Users.Items, 1)
	assert.Len(t, result.Tags.Items, 1)
	assert.Equal(t, "Go Tutorial", result.Blogs.Items[0].Title)
}

func TestSearch_EmptyQuery_ReturnsEmpty(t *testing.T) {
	repo := &mockRepo{}
	svc := search.NewService(repo)

	result, err := svc.Search(context.Background(), "", uuid.Nil, 1)
	assert.NoError(t, err)
	assert.Empty(t, result.Blogs.Items)
	assert.Empty(t, result.Users.Items)
	assert.Empty(t, result.Tags.Items)
	repo.AssertNotCalled(t, "SearchBlogs")
}

func TestSearch_WhitespaceOnlyQuery_ReturnsEmpty(t *testing.T) {
	repo := &mockRepo{}
	svc := search.NewService(repo)

	result, err := svc.Search(context.Background(), "   ", uuid.Nil, 1)
	assert.NoError(t, err)
	assert.Empty(t, result.Blogs.Items)
	repo.AssertNotCalled(t, "SearchBlogs")
}

func TestSearch_GuestCannotSeePrivateBlogs(t *testing.T) {
	// Privacy filtering is enforced at repo layer (tsvector + privacy WHERE clause)
	// Service passes uuid.Nil as viewerID for guests — repo handles the rest
	repo := &mockRepo{}
	svc := search.NewService(repo)

	repo.On("SearchBlogs", mock.Anything, "secret", uuid.Nil, 1).
		Return([]*search.BlogResult{}, 0, nil) // repo returns nothing for guest + private
	repo.On("SearchUsers", mock.Anything, "secret", 1).Return([]*search.UserResult{}, 0, nil)
	repo.On("SearchTags", mock.Anything, "secret", 1).Return([]*search.TagResult{}, 0, nil)

	result, err := svc.Search(context.Background(), "secret", uuid.Nil, 1)
	assert.NoError(t, err)
	assert.Empty(t, result.Blogs.Items)
}

func TestSearch_QueryTrimmed(t *testing.T) {
	repo := &mockRepo{}
	svc := search.NewService(repo)

	viewerID := uuid.New()
	repo.On("SearchBlogs", mock.Anything, "go", viewerID, 1).Return([]*search.BlogResult{}, 0, nil)
	repo.On("SearchUsers", mock.Anything, "go", 1).Return([]*search.UserResult{}, 0, nil)
	repo.On("SearchTags", mock.Anything, "go", 1).Return([]*search.TagResult{}, 0, nil)

	_, err := svc.Search(context.Background(), "  go  ", viewerID, 1)
	assert.NoError(t, err)
	repo.AssertCalled(t, "SearchBlogs", mock.Anything, "go", viewerID, 1)
}

func TestSearch_PaginationPassedThrough(t *testing.T) {
	repo := &mockRepo{}
	svc := search.NewService(repo)

	viewerID := uuid.New()
	repo.On("SearchBlogs", mock.Anything, "rust", viewerID, 3).Return([]*search.BlogResult{}, 0, nil)
	repo.On("SearchUsers", mock.Anything, "rust", 3).Return([]*search.UserResult{}, 0, nil)
	repo.On("SearchTags", mock.Anything, "rust", 3).Return([]*search.TagResult{}, 0, nil)

	_, err := svc.Search(context.Background(), "rust", viewerID, 3)
	assert.NoError(t, err)
	repo.AssertCalled(t, "SearchBlogs", mock.Anything, "rust", viewerID, 3)
}

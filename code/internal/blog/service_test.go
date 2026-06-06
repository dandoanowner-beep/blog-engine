package blog_test

import (
	"context"
	"testing"

	"blog-engine/internal/blog"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

var _ blog.Repository = (*mockRepo)(nil)
var _ blog.Sanitizer = (*mockSanitizer)(nil)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Create(ctx context.Context, b *blog.Blog) error {
	return m.Called(ctx, b).Error(0)
}
func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*blog.Blog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*blog.Blog), args.Error(1)
}
func (m *mockRepo) Update(ctx context.Context, b *blog.Blog) error {
	return m.Called(ctx, b).Error(0)
}
func (m *mockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRepo) IsBlocked(ctx context.Context, viewerID, authorID uuid.UUID) (bool, error) {
	args := m.Called(ctx, viewerID, authorID)
	return args.Bool(0), args.Error(1)
}
func (m *mockRepo) AreFriends(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	args := m.Called(ctx, userA, userB)
	return args.Bool(0), args.Error(1)
}
func (m *mockRepo) UpsertTags(ctx context.Context, names []string) ([]blog.Tag, error) {
	args := m.Called(ctx, names)
	return args.Get(0).([]blog.Tag), args.Error(1)
}
func (m *mockRepo) UpsertCategories(ctx context.Context, ids []uuid.UUID) error {
	return m.Called(ctx, ids).Error(0)
}
func (m *mockRepo) UpdateFeedScore(ctx context.Context, blogID uuid.UUID) error {
	return m.Called(ctx, blogID).Error(0)
}

type mockSanitizer struct{ mock.Mock }

func (m *mockSanitizer) Sanitize(html string) string {
	return m.Called(html).String(0)
}

// --- Tests: AC-BLOG-001 ---

func TestCreateBlog_Success(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san)

	authorID := uuid.New()
	san.On("Sanitize", "<p>content</p>").Return("<p>content</p>")
	repo.On("UpsertTags", mock.Anything, []string{"go", "tech"}).Return([]blog.Tag{{ID: uuid.New(), Name: "go"}, {ID: uuid.New(), Name: "tech"}}, nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*blog.Blog")).Return(nil)

	input := blog.CreateInput{
		AuthorID:    authorID,
		Title:       "My First Blog",
		Content:     "<p>content</p>",
		Privacy:     blog.PrivacyPublic,
		Status:      blog.StatusPublished,
		TagNames:    []string{"go", "tech"},
		CategoryIDs: []uuid.UUID{uuid.New()},
	}
	b, err := svc.Create(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, "My First Blog", b.Title)
	assert.Equal(t, authorID, b.AuthorID)
}

func TestCreateBlog_MissingTitle(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san)

	input := blog.CreateInput{
		AuthorID: uuid.New(),
		Title:    "",
		Content:  "<p>content</p>",
		TagNames: []string{"go"},
	}
	_, err := svc.Create(context.Background(), input)
	assert.ErrorIs(t, err, blog.ErrMissingTitle)
}

func TestCreateBlog_MissingTags(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san)

	input := blog.CreateInput{
		AuthorID: uuid.New(),
		Title:    "A Blog",
		Content:  "<p>content</p>",
		TagNames: []string{},
	}
	_, err := svc.Create(context.Background(), input)
	assert.ErrorIs(t, err, blog.ErrMissingTags)
}

// --- Tests: AC-BLOG-003 Privacy ---

func TestGetBlog_PublicVisibleToGuest(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san)

	blogID := uuid.New()
	b := &blog.Blog{
		ID:      blogID,
		Privacy: blog.PrivacyPublic,
		Status:  blog.StatusPublished,
		Content: "full content here for reading",
	}
	repo.On("GetByID", mock.Anything, blogID).Return(b, nil)

	result, partial, err := svc.GetForViewer(context.Background(), blogID, uuid.Nil)
	assert.NoError(t, err)
	assert.True(t, partial) // guest gets partial
	assert.NotEmpty(t, result.Content)
}

func TestGetBlog_FriendOnlyHiddenFromStranger(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san)

	authorID := uuid.New()
	viewerID := uuid.New()
	blogID := uuid.New()
	b := &blog.Blog{
		ID:       blogID,
		AuthorID: authorID,
		Privacy:  blog.PrivacyFriendOnly,
		Status:   blog.StatusPublished,
	}
	repo.On("GetByID", mock.Anything, blogID).Return(b, nil)
	repo.On("AreFriends", mock.Anything, viewerID, authorID).Return(false, nil)
	repo.On("IsBlocked", mock.Anything, viewerID, authorID).Return(false, nil)

	_, _, err := svc.GetForViewer(context.Background(), blogID, viewerID)
	assert.ErrorIs(t, err, blog.ErrAccessDenied)
}

func TestGetBlog_FriendOnlyVisibleToFriend(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san)

	authorID := uuid.New()
	friendID := uuid.New()
	blogID := uuid.New()
	b := &blog.Blog{
		ID:       blogID,
		AuthorID: authorID,
		Privacy:  blog.PrivacyFriendOnly,
		Status:   blog.StatusPublished,
		Content:  "friend only content",
	}
	repo.On("GetByID", mock.Anything, blogID).Return(b, nil)
	repo.On("AreFriends", mock.Anything, friendID, authorID).Return(true, nil)
	repo.On("IsBlocked", mock.Anything, friendID, authorID).Return(false, nil)

	result, partial, err := svc.GetForViewer(context.Background(), blogID, friendID)
	assert.NoError(t, err)
	assert.False(t, partial)
	assert.Equal(t, "friend only content", result.Content)
}

func TestGetBlog_OnlyMeHiddenFromEveryone(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san)

	authorID := uuid.New()
	otherID := uuid.New()
	blogID := uuid.New()
	b := &blog.Blog{
		ID:       blogID,
		AuthorID: authorID,
		Privacy:  blog.PrivacyOnlyMe,
		Status:   blog.StatusPublished,
	}
	repo.On("GetByID", mock.Anything, blogID).Return(b, nil)
	repo.On("IsBlocked", mock.Anything, otherID, authorID).Return(false, nil)

	_, _, err := svc.GetForViewer(context.Background(), blogID, otherID)
	assert.ErrorIs(t, err, blog.ErrAccessDenied)
}

func TestGetBlog_OnlyMeVisibleToAuthor(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san)

	authorID := uuid.New()
	blogID := uuid.New()
	b := &blog.Blog{
		ID:       blogID,
		AuthorID: authorID,
		Privacy:  blog.PrivacyOnlyMe,
		Status:   blog.StatusPublished,
		Content:  "private content",
	}
	repo.On("GetByID", mock.Anything, blogID).Return(b, nil)
	repo.On("IsBlocked", mock.Anything, authorID, authorID).Return(false, nil)

	result, partial, err := svc.GetForViewer(context.Background(), blogID, authorID)
	assert.NoError(t, err)
	assert.False(t, partial)
	assert.Equal(t, "private content", result.Content)
}

// --- Tests: AC-BLOG-001 (XSS sanitization) ---

func TestCreateBlog_ContentSanitized(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san)

	dirty := `<p>Hello</p><script>alert('xss')</script>`
	clean := `<p>Hello</p>`

	san.On("Sanitize", dirty).Return(clean)
	repo.On("UpsertTags", mock.Anything, mock.Anything).Return([]blog.Tag{{ID: uuid.New(), Name: "test"}}, nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*blog.Blog")).Return(nil)

	input := blog.CreateInput{
		AuthorID: uuid.New(),
		Title:    "Test",
		Content:  dirty,
		TagNames: []string{"test"},
		CategoryIDs: []uuid.UUID{uuid.New()},
	}
	b, err := svc.Create(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, clean, b.Content)
}

// --- Tests: AC-BLOG-004 Delete ---

func TestDeleteBlog_AuthorCanDelete(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san)

	authorID := uuid.New()
	blogID := uuid.New()
	b := &blog.Blog{ID: blogID, AuthorID: authorID}

	repo.On("GetByID", mock.Anything, blogID).Return(b, nil)
	repo.On("Delete", mock.Anything, blogID).Return(nil)

	err := svc.Delete(context.Background(), blogID, authorID, "user")
	assert.NoError(t, err)
}

func TestDeleteBlog_NonAuthorCannotDelete(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san)

	authorID := uuid.New()
	otherID := uuid.New()
	blogID := uuid.New()
	b := &blog.Blog{ID: blogID, AuthorID: authorID}

	repo.On("GetByID", mock.Anything, blogID).Return(b, nil)

	err := svc.Delete(context.Background(), blogID, otherID, "user")
	assert.ErrorIs(t, err, blog.ErrForbidden)
}

func TestDeleteBlog_ModeratorCanDeleteAny(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san)

	authorID := uuid.New()
	modID := uuid.New()
	blogID := uuid.New()
	b := &blog.Blog{ID: blogID, AuthorID: authorID}

	repo.On("GetByID", mock.Anything, blogID).Return(b, nil)
	repo.On("Delete", mock.Anything, blogID).Return(nil)

	err := svc.Delete(context.Background(), blogID, modID, "moderator")
	assert.NoError(t, err)
}

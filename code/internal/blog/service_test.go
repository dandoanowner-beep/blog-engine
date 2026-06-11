package blog_test

import (
	"context"
	"strings"
	"testing"
	"time"

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
func (m *mockRepo) UpsertCategories(ctx context.Context, names []string) ([]blog.Category, error) {
	args := m.Called(ctx, names)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]blog.Category), args.Error(1)
}
func (m *mockRepo) SetBlogCategories(ctx context.Context, blogID uuid.UUID, categoryIDs []uuid.UUID) error {
	return m.Called(ctx, blogID, categoryIDs).Error(0)
}
func (m *mockRepo) ListCategories(ctx context.Context) ([]blog.CategoryWithCount, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]blog.CategoryWithCount), args.Error(1)
}
func (m *mockRepo) UpdateFeedScore(ctx context.Context, blogID uuid.UUID) error {
	return m.Called(ctx, blogID).Error(0)
}
func (m *mockRepo) UpdateTranslation(ctx context.Context, blogID uuid.UUID, titleEN, bodyEN, status string) error {
	return m.Called(ctx, blogID, titleEN, bodyEN, status).Error(0)
}
func (m *mockRepo) GetArticlesFeed(ctx context.Context, page, perPage int, category string) ([]*blog.Blog, int, error) {
	args := m.Called(ctx, page, perPage, category)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*blog.Blog), args.Int(1), args.Error(2)
}

// --- Mock Translator ---

var _ blog.Translator = (*mockTranslator)(nil)

type mockTranslator struct {
	mock.Mock
	done chan struct{}
}

func (m *mockTranslator) Translate(ctx context.Context, titleVI, bodyVI string) (string, string, error) {
	args := m.Called(ctx, titleVI, bodyVI)
	if m.done != nil {
		select {
		case <-m.done: // already closed
		default:
			close(m.done)
		}
	}
	return args.String(0), args.String(1), args.Error(2)
}

type mockSanitizer struct{ mock.Mock }

func (m *mockSanitizer) Sanitize(html string) string {
	return m.Called(html).String(0)
}

// --- Tests: AC-BLOG-001 ---

func TestCreateBlog_Success(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

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
	}
	b, err := svc.Create(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, "My First Blog", b.Title)
	assert.Equal(t, authorID, b.AuthorID)
}

// CR-002: categories are upserted by name (like tags) and associated to the blog
func TestCreateBlog_WithCategories_AssociatesThem(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

	catID := uuid.New()
	san.On("Sanitize", "<p>c</p>").Return("<p>c</p>")
	repo.On("UpsertTags", mock.Anything, []string{"go"}).Return([]blog.Tag{{ID: uuid.New(), Name: "go"}}, nil)
	repo.On("UpsertCategories", mock.Anything, []string{"Tutorials"}).Return([]blog.Category{{ID: catID, Name: "Tutorials", Slug: "tutorials"}}, nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*blog.Blog")).Return(nil)
	repo.On("SetBlogCategories", mock.Anything, mock.AnythingOfType("uuid.UUID"), []uuid.UUID{catID}).Return(nil)

	input := blog.CreateInput{
		AuthorID:      uuid.New(),
		Title:         "With categories",
		Content:       "<p>c</p>",
		TagNames:      []string{"go"},
		CategoryNames: []string{"Tutorials"},
	}
	b, err := svc.Create(context.Background(), input)
	assert.NoError(t, err)
	assert.Len(t, b.Categories, 1)
	repo.AssertCalled(t, "SetBlogCategories", mock.Anything, b.ID, []uuid.UUID{catID})
}

func TestListCategories_ReturnsCountsFromRepo(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

	repo.On("ListCategories", mock.Anything).Return([]blog.CategoryWithCount{
		{Category: blog.Category{ID: uuid.New(), Name: "Tutorials", Slug: "tutorials"}, BlogCount: 4},
	}, nil)

	cats, err := svc.ListCategories(context.Background())
	assert.NoError(t, err)
	assert.Len(t, cats, 1)
	assert.Equal(t, 4, cats[0].BlogCount)
}

func TestCreateBlog_MissingTitle(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

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
	svc := blog.NewService(repo, san, nil)

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
	svc := blog.NewService(repo, san, nil)

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

// --- Tests: CR-001 / BUG-007 — articles feed wired to repository ---

func TestArticlesFeed_ReturnsBlogsAndTotal(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

	feed := []*blog.Blog{
		{ID: uuid.New(), Title: "Post A", AuthorUsername: "chubeunu"},
		{ID: uuid.New(), Title: "Post B", AuthorUsername: "chubeunu"},
	}
	repo.On("GetArticlesFeed", mock.Anything, 2, blog.ArticlesPerPage, "").Return(feed, 11, nil)

	blogs, total, err := svc.ArticlesFeed(context.Background(), 2, "")
	assert.NoError(t, err)
	assert.Len(t, blogs, 2)
	assert.Equal(t, 11, total)
}

func TestArticlesFeed_ClampsPageToOne(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

	repo.On("GetArticlesFeed", mock.Anything, 1, blog.ArticlesPerPage, "").Return([]*blog.Blog{}, 0, nil)

	_, _, err := svc.ArticlesFeed(context.Background(), 0, "")
	assert.NoError(t, err)
	repo.AssertCalled(t, "GetArticlesFeed", mock.Anything, 1, blog.ArticlesPerPage, "")
}

// CR-002: category slug filters the article feed (Categories browse page)
func TestArticlesFeed_FiltersByCategory(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

	repo.On("GetArticlesFeed", mock.Anything, 1, blog.ArticlesPerPage, "tutorials").
		Return([]*blog.Blog{{ID: uuid.New(), Title: "Filtered"}}, 1, nil)

	blogs, total, err := svc.ArticlesFeed(context.Background(), 1, "tutorials")
	assert.NoError(t, err)
	assert.Len(t, blogs, 1)
	assert.Equal(t, 1, total)
}

// BUG-006: guest gate must be enforced server-side — full content must never
// leave the server for a guest viewer (FR-BLOG-006: guest reads ~30%).
func TestGetBlog_GuestContentTruncatedServerSide(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

	blogID := uuid.New()
	fullContent := strings.Repeat("a", 1000)
	b := &blog.Blog{
		ID:      blogID,
		Privacy: blog.PrivacyPublic,
		Status:  blog.StatusPublished,
		Content: fullContent,
		BodyEn:  strings.Repeat("b", 1000),
	}
	repo.On("GetByID", mock.Anything, blogID).Return(b, nil)

	result, partial, err := svc.GetForViewer(context.Background(), blogID, uuid.Nil)
	assert.NoError(t, err)
	assert.True(t, partial)
	assert.NotEmpty(t, result.Content)
	assert.LessOrEqual(t, len([]rune(result.Content)), 300, "guest must receive at most ~30%% of content")
	assert.LessOrEqual(t, len([]rune(result.BodyEn)), 300, "translation must be truncated too — it would leak gated content")
	// the repo's object must not be mutated — authenticated viewers still need full content
	assert.Equal(t, fullContent, b.Content)
}

func TestGetBlog_AuthenticatedViewerGetsFullContent(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

	blogID := uuid.New()
	viewerID := uuid.New()
	fullContent := strings.Repeat("a", 1000)
	b := &blog.Blog{
		ID:       blogID,
		AuthorID: uuid.New(),
		Privacy:  blog.PrivacyPublic,
		Status:   blog.StatusPublished,
		Content:  fullContent,
	}
	repo.On("GetByID", mock.Anything, blogID).Return(b, nil)
	repo.On("IsBlocked", mock.Anything, viewerID, b.AuthorID).Return(false, nil)

	result, partial, err := svc.GetForViewer(context.Background(), blogID, viewerID)
	assert.NoError(t, err)
	assert.False(t, partial)
	assert.Equal(t, fullContent, result.Content)
}

func TestGetBlog_FriendOnlyHiddenFromStranger(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

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
	svc := blog.NewService(repo, san, nil)

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
	svc := blog.NewService(repo, san, nil)

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
	svc := blog.NewService(repo, san, nil)

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
	svc := blog.NewService(repo, san, nil)

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
	}
	b, err := svc.Create(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, clean, b.Content)
}

// --- Tests: AC-BLOG-004 Delete ---

func TestDeleteBlog_AuthorCanDelete(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

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
	svc := blog.NewService(repo, san, nil)

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
	svc := blog.NewService(repo, san, nil)

	authorID := uuid.New()
	modID := uuid.New()
	blogID := uuid.New()
	b := &blog.Blog{ID: blogID, AuthorID: authorID}

	repo.On("GetByID", mock.Anything, blogID).Return(b, nil)
	repo.On("Delete", mock.Anything, blogID).Return(nil)

	err := svc.Delete(context.Background(), blogID, modID, "moderator")
	assert.NoError(t, err)
}

// ════════════════════════════════════════════════════════════
// AC-I18N-003: translation_status set on create
// ════════════════════════════════════════════════════════════

func TestCreate_WithTranslator_SetsStatusPending(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	done := make(chan struct{})
	tr := &mockTranslator{done: done}
	svc := blog.NewService(repo, san, tr)

	authorID := uuid.New()
	san.On("Sanitize", mock.Anything).Return("<p>content</p>")
	repo.On("UpsertTags", mock.Anything, mock.Anything).Return([]blog.Tag{{ID: uuid.New(), Name: "go"}}, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(b *blog.Blog) bool {
		return b.TranslationStatus == blog.TranslationStatusPending
	})).Return(nil)
	tr.On("Translate", mock.Anything, mock.Anything, mock.Anything).Return("Title EN", "Body EN", nil)
	repo.On("UpdateTranslation", mock.Anything, mock.Anything, "Title EN", "Body EN", blog.TranslationStatusDone).Return(nil)

	input := blog.CreateInput{
		AuthorID: authorID, Title: "Tiêu đề", Content: "<p>content</p>",
		Privacy: blog.PrivacyPublic, Status: blog.StatusPublished, TagNames: []string{"go"},
	}
	b, err := svc.Create(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, blog.TranslationStatusPending, b.TranslationStatus)

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("translateAsync was not called")
	}
	repo.AssertExpectations(t)
	tr.AssertExpectations(t)
}

func TestCreate_WithoutTranslator_SetsStatusNone(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

	san.On("Sanitize", mock.Anything).Return("<p>content</p>")
	repo.On("UpsertTags", mock.Anything, mock.Anything).Return([]blog.Tag{{ID: uuid.New(), Name: "go"}}, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(b *blog.Blog) bool {
		return b.TranslationStatus == blog.TranslationStatusNone
	})).Return(nil)

	input := blog.CreateInput{
		AuthorID: uuid.New(), Title: "Tiêu đề", Content: "<p>content</p>",
		Privacy: blog.PrivacyPublic, Status: blog.StatusPublished, TagNames: []string{"go"},
	}
	b, err := svc.Create(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, blog.TranslationStatusNone, b.TranslationStatus)
}

// AC-I18N-004: translation failure → status=failed; blog was already saved (not rolled back)
func TestCreate_TranslationFailure_SetsStatusFailed(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	done := make(chan struct{})
	tr := &mockTranslator{done: done}
	svc := blog.NewService(repo, san, tr)

	san.On("Sanitize", mock.Anything).Return("<p>content</p>")
	repo.On("UpsertTags", mock.Anything, mock.Anything).Return([]blog.Tag{{ID: uuid.New(), Name: "go"}}, nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)
	tr.On("Translate", mock.Anything, mock.Anything, mock.Anything).Return("", "", assert.AnError)
	repo.On("UpdateTranslation", mock.Anything, mock.Anything, "", "", blog.TranslationStatusFailed).Return(nil)

	input := blog.CreateInput{
		AuthorID: uuid.New(), Title: "Tiêu đề", Content: "<p>content</p>",
		Privacy: blog.PrivacyPublic, Status: blog.StatusDraft, TagNames: []string{"go"},
	}
	_, err := svc.Create(context.Background(), input)
	assert.NoError(t, err) // blog saved successfully, translation failure is async

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("translateAsync was not called")
	}
	repo.AssertCalled(t, "UpdateTranslation", mock.Anything, mock.Anything, "", "", blog.TranslationStatusFailed)
}

// ════════════════════════════════════════════════════════════
// AC-I18N-004: Update triggers re-translation on content change
// ════════════════════════════════════════════════════════════

func TestUpdate_Success_AuthorCanUpdate(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

	authorID := uuid.New()
	blogID := uuid.New()
	existing := &blog.Blog{ID: blogID, AuthorID: authorID, Title: "Old", Content: "<p>old</p>", Privacy: blog.PrivacyPublic}

	newTitle := "New Title"
	repo.On("GetByID", mock.Anything, blogID).Return(existing, nil)
	san.On("Sanitize", mock.Anything).Return("<p>old</p>")
	repo.On("Update", mock.Anything, mock.AnythingOfType("*blog.Blog")).Return(nil)

	input := blog.UpdateInput{RequesterID: authorID, Title: &newTitle}
	b, err := svc.Update(context.Background(), blogID, input)
	assert.NoError(t, err)
	assert.Equal(t, "New Title", b.Title)
}

func TestUpdate_ForbiddenForNonAuthor(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := blog.NewService(repo, san, nil)

	authorID := uuid.New()
	otherID := uuid.New()
	blogID := uuid.New()
	existing := &blog.Blog{ID: blogID, AuthorID: authorID, Title: "Old"}
	repo.On("GetByID", mock.Anything, blogID).Return(existing, nil)

	input := blog.UpdateInput{RequesterID: otherID}
	_, err := svc.Update(context.Background(), blogID, input)
	assert.ErrorIs(t, err, blog.ErrForbidden)
}

func TestUpdate_ContentChange_TriggersTranslation(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	done := make(chan struct{})
	tr := &mockTranslator{done: done}
	svc := blog.NewService(repo, san, tr)

	authorID := uuid.New()
	blogID := uuid.New()
	existing := &blog.Blog{ID: blogID, AuthorID: authorID, Title: "Old", Content: "<p>old</p>"}

	newContent := "<p>new content</p>"
	repo.On("GetByID", mock.Anything, blogID).Return(existing, nil)
	san.On("Sanitize", newContent).Return(newContent)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(b *blog.Blog) bool {
		return b.TranslationStatus == blog.TranslationStatusPending
	})).Return(nil)
	tr.On("Translate", mock.Anything, mock.Anything, mock.Anything).Return("EN Title", "EN Body", nil)
	repo.On("UpdateTranslation", mock.Anything, blogID, "EN Title", "EN Body", blog.TranslationStatusDone).Return(nil)

	input := blog.UpdateInput{RequesterID: authorID, Content: &newContent}
	_, err := svc.Update(context.Background(), blogID, input)
	assert.NoError(t, err)

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("translation not triggered on content change")
	}
	tr.AssertExpectations(t)
}

func TestUpdate_NoContentChange_DoesNotTriggerTranslation(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	tr := &mockTranslator{}
	svc := blog.NewService(repo, san, tr)

	authorID := uuid.New()
	blogID := uuid.New()
	priv := blog.PrivacyFriendOnly
	existing := &blog.Blog{ID: blogID, AuthorID: authorID, Title: "Same", Content: "<p>same</p>", Privacy: blog.PrivacyPublic}

	repo.On("GetByID", mock.Anything, blogID).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	input := blog.UpdateInput{RequesterID: authorID, Privacy: &priv}
	_, err := svc.Update(context.Background(), blogID, input)
	assert.NoError(t, err)
	tr.AssertNotCalled(t, "Translate", mock.Anything, mock.Anything, mock.Anything)
}

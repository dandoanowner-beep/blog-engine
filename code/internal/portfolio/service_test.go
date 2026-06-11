package portfolio_test

import (
	"context"
	"testing"

	"blog-engine/internal/portfolio"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

var _ portfolio.Repository = (*mockRepo)(nil)
var _ portfolio.Sanitizer = (*mockSanitizer)(nil)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Create(ctx context.Context, p *portfolio.Project) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*portfolio.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*portfolio.Project), args.Error(1)
}
func (m *mockRepo) Update(ctx context.Context, p *portfolio.Project) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRepo) List(ctx context.Context) ([]*portfolio.Project, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*portfolio.Project), args.Error(1)
}

type mockSanitizer struct{ mock.Mock }

func (m *mockSanitizer) Sanitize(html string) string {
	return m.Called(html).String(0)
}

// --- Tests: FR-CR002-001 Portfolio ---

func TestCreateProject_Success_SanitizesDescription(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := portfolio.NewService(repo, san)

	dirty := `<p>My project</p><script>x</script>`
	san.On("Sanitize", dirty).Return("<p>My project</p>")
	repo.On("Create", mock.Anything, mock.AnythingOfType("*portfolio.Project")).Return(nil)

	p, err := svc.Create(context.Background(), portfolio.CreateInput{
		Title:       "Blog Engine",
		Description: dirty,
		TechStack:   "Go, React, PostgreSQL",
		RepoURL:     "https://github.com/dandoanowner-beep/blog-engine",
	})
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, p.ID)
	assert.Equal(t, "Blog Engine", p.Title)
	assert.Equal(t, "<p>My project</p>", p.Description)
}

func TestCreateProject_MissingTitle(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := portfolio.NewService(repo, san)

	_, err := svc.Create(context.Background(), portfolio.CreateInput{Title: "   "})
	assert.ErrorIs(t, err, portfolio.ErrMissingTitle)
}

func TestUpdateProject_Success(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := portfolio.NewService(repo, san)

	id := uuid.New()
	existing := &portfolio.Project{ID: id, Title: "Old", Description: "old desc"}
	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	san.On("Sanitize", "new desc").Return("new desc")
	repo.On("Update", mock.Anything, mock.AnythingOfType("*portfolio.Project")).Return(nil)

	newTitle := "New Title"
	newDesc := "new desc"
	p, err := svc.Update(context.Background(), id, portfolio.UpdateInput{Title: &newTitle, Description: &newDesc})
	assert.NoError(t, err)
	assert.Equal(t, "New Title", p.Title)
	assert.Equal(t, "new desc", p.Description)
}

func TestUpdateProject_NotFound(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := portfolio.NewService(repo, san)

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, portfolio.ErrNotFound)

	newTitle := "x"
	_, err := svc.Update(context.Background(), id, portfolio.UpdateInput{Title: &newTitle})
	assert.ErrorIs(t, err, portfolio.ErrNotFound)
}

func TestUpdateProject_BlankTitleRejected(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := portfolio.NewService(repo, san)

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(&portfolio.Project{ID: id, Title: "Keep"}, nil)

	blank := "  "
	_, err := svc.Update(context.Background(), id, portfolio.UpdateInput{Title: &blank})
	assert.ErrorIs(t, err, portfolio.ErrMissingTitle)
}

func TestListProjects_ReturnsAll(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := portfolio.NewService(repo, san)

	repo.On("List", mock.Anything).Return([]*portfolio.Project{
		{ID: uuid.New(), Title: "A"},
		{ID: uuid.New(), Title: "B"},
	}, nil)

	list, err := svc.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestDeleteProject_Delegates(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := portfolio.NewService(repo, san)

	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(nil)

	assert.NoError(t, svc.Delete(context.Background(), id))
}

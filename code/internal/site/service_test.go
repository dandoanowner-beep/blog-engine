package site_test

import (
	"context"
	"testing"

	"blog-engine/internal/site"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

var _ site.Repository = (*mockRepo)(nil)
var _ site.Sanitizer = (*mockSanitizer)(nil)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}
func (m *mockRepo) Upsert(ctx context.Context, key, content string) error {
	return m.Called(ctx, key, content).Error(0)
}

type mockSanitizer struct{ mock.Mock }

func (m *mockSanitizer) Sanitize(html string) string {
	return m.Called(html).String(0)
}

// --- Tests: FR-CR002-002 Author page ---

func TestGetAbout_ReturnsContent(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := site.NewService(repo, san)

	repo.On("Get", mock.Anything, "about").Return("<p>My story</p>", nil)

	content, err := svc.GetAbout(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "<p>My story</p>", content)
}

func TestGetAbout_NotWrittenYet_ReturnsEmptyNoError(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := site.NewService(repo, san)

	repo.On("Get", mock.Anything, "about").Return("", site.ErrNotFound)

	content, err := svc.GetAbout(context.Background())
	assert.NoError(t, err) // empty state is not an error (FR-CR002-002)
	assert.Equal(t, "", content)
}

func TestUpdateAbout_SanitizesBeforeStore(t *testing.T) {
	repo := &mockRepo{}
	san := &mockSanitizer{}
	svc := site.NewService(repo, san)

	dirty := `<p>Story</p><script>alert(1)</script>`
	san.On("Sanitize", dirty).Return("<p>Story</p>")
	repo.On("Upsert", mock.Anything, "about", "<p>Story</p>").Return(nil)

	err := svc.UpdateAbout(context.Background(), dirty)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Upsert", mock.Anything, "about", "<p>Story</p>")
}

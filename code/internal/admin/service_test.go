package admin_test

import (
	"context"
	"testing"

	"blog-engine/internal/admin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock ---

type mockRepo struct{ mock.Mock }

func (m *mockRepo) ListUsers(ctx context.Context, page int, role string) ([]*admin.UserRow, int, error) {
	args := m.Called(ctx, page, role)
	return args.Get(0).([]*admin.UserRow), args.Int(1), args.Error(2)
}
func (m *mockRepo) ChangeUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	return m.Called(ctx, userID, role).Error(0)
}
func (m *mockRepo) ListReports(ctx context.Context, status string, page int) ([]*admin.ReportRow, int, error) {
	args := m.Called(ctx, status, page)
	return args.Get(0).([]*admin.ReportRow), args.Int(1), args.Error(2)
}
func (m *mockRepo) ResolveReport(ctx context.Context, reportID uuid.UUID, action string, resolverID uuid.UUID) error {
	return m.Called(ctx, reportID, action, resolverID).Error(0)
}
func (m *mockRepo) DeleteContent(ctx context.Context, reportID uuid.UUID) error {
	return m.Called(ctx, reportID).Error(0)
}
func (m *mockRepo) GetStats(ctx context.Context) (*admin.Stats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*admin.Stats), args.Error(1)
}

// ════════════════════════════════════════════════════════════
// AC-ADMIN-001: User management
// ════════════════════════════════════════════════════════════

func TestListUsers_ReturnsPagedResults(t *testing.T) {
	repo := &mockRepo{}
	svc := admin.NewService(repo)

	users := []*admin.UserRow{
		{ID: uuid.New(), Username: "alice", Role: "user"},
		{ID: uuid.New(), Username: "bob", Role: "moderator"},
	}
	repo.On("ListUsers", mock.Anything, 1, "").Return(users, 2, nil)

	result, total, err := svc.ListUsers(context.Background(), 1, "")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, total)
}

func TestPromoteToModerator_Success(t *testing.T) {
	repo := &mockRepo{}
	svc := admin.NewService(repo)

	userID := uuid.New()
	repo.On("ChangeUserRole", mock.Anything, userID, "moderator").Return(nil)

	err := svc.ChangeRole(context.Background(), userID, "moderator")
	assert.NoError(t, err)
}

func TestChangeRole_InvalidRole_Error(t *testing.T) {
	repo := &mockRepo{}
	svc := admin.NewService(repo)

	err := svc.ChangeRole(context.Background(), uuid.New(), "superadmin")
	assert.ErrorIs(t, err, admin.ErrInvalidRole)
}

func TestChangeRole_CannotDemoteOwner(t *testing.T) {
	repo := &mockRepo{}
	svc := admin.NewService(repo)

	err := svc.ChangeRole(context.Background(), uuid.New(), "owner")
	assert.ErrorIs(t, err, admin.ErrCannotAssignOwner)
}

// ════════════════════════════════════════════════════════════
// AC-ADMIN-001: Reports queue
// ════════════════════════════════════════════════════════════

func TestListReports_ReturnsPendingReports(t *testing.T) {
	repo := &mockRepo{}
	svc := admin.NewService(repo)

	reports := []*admin.ReportRow{
		{ID: uuid.New(), Reason: "spam", Status: "pending"},
	}
	repo.On("ListReports", mock.Anything, "pending", 1).Return(reports, 1, nil)

	result, total, err := svc.ListReports(context.Background(), "pending", 1)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)
}

func TestResolveReport_DeleteContent(t *testing.T) {
	repo := &mockRepo{}
	svc := admin.NewService(repo)

	reportID := uuid.New()
	resolverID := uuid.New()
	repo.On("DeleteContent", mock.Anything, reportID).Return(nil)
	repo.On("ResolveReport", mock.Anything, reportID, "delete_content", resolverID).Return(nil)

	err := svc.ResolveReport(context.Background(), reportID, "delete_content", resolverID)
	assert.NoError(t, err)
	repo.AssertCalled(t, "DeleteContent", mock.Anything, reportID)
	repo.AssertCalled(t, "ResolveReport", mock.Anything, reportID, "delete_content", resolverID)
}

func TestResolveReport_Dismiss(t *testing.T) {
	repo := &mockRepo{}
	svc := admin.NewService(repo)

	reportID := uuid.New()
	resolverID := uuid.New()
	repo.On("ResolveReport", mock.Anything, reportID, "dismiss", resolverID).Return(nil)

	err := svc.ResolveReport(context.Background(), reportID, "dismiss", resolverID)
	assert.NoError(t, err)
	repo.AssertNotCalled(t, "DeleteContent")
}

func TestResolveReport_InvalidAction_Error(t *testing.T) {
	repo := &mockRepo{}
	svc := admin.NewService(repo)

	err := svc.ResolveReport(context.Background(), uuid.New(), "ban_user", uuid.New())
	assert.ErrorIs(t, err, admin.ErrInvalidReportAction)
}

// ════════════════════════════════════════════════════════════
// AC-ADMIN-001: Platform statistics
// ════════════════════════════════════════════════════════════

func TestGetStats_ReturnsAllFields(t *testing.T) {
	repo := &mockRepo{}
	svc := admin.NewService(repo)

	stats := &admin.Stats{
		TotalUsers:        1200,
		TotalBlogs:        4500,
		TotalComments:     18000,
		NewSignupsToday:   12,
		NewSignupsThisWeek: 67,
	}
	repo.On("GetStats", mock.Anything).Return(stats, nil)

	result, err := svc.GetStats(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1200, result.TotalUsers)
	assert.Equal(t, 4500, result.TotalBlogs)
	assert.Equal(t, 18000, result.TotalComments)
	assert.Equal(t, 12, result.NewSignupsToday)
	assert.Equal(t, 67, result.NewSignupsThisWeek)
}

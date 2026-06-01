package admin_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-engine/internal/admin"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAdminRepo struct{ mock.Mock }

func (m *mockAdminRepo) ListUsers(ctx context.Context, page int, role string) ([]*admin.UserRow, int, error) {
	args := m.Called(ctx, page, role)
	return args.Get(0).([]*admin.UserRow), args.Int(1), args.Error(2)
}
func (m *mockAdminRepo) ChangeUserRole(ctx context.Context, id uuid.UUID, role string) error {
	return m.Called(ctx, id, role).Error(0)
}
func (m *mockAdminRepo) ListReports(ctx context.Context, status string, page int) ([]*admin.ReportRow, int, error) {
	args := m.Called(ctx, status, page)
	return args.Get(0).([]*admin.ReportRow), args.Int(1), args.Error(2)
}
func (m *mockAdminRepo) ResolveReport(ctx context.Context, id uuid.UUID, action string, resolverID uuid.UUID) error {
	return m.Called(ctx, id, action, resolverID).Error(0)
}
func (m *mockAdminRepo) DeleteContent(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockAdminRepo) GetStats(ctx context.Context) (*admin.Stats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*admin.Stats), args.Error(1)
}

func TestListUsersHandler_Returns200(t *testing.T) {
	repo := &mockAdminRepo{}
	svc := admin.NewService(repo)
	h := admin.NewHandler(svc)

	users := []*admin.UserRow{{ID: uuid.New(), Username: "alice", Role: "user"}}
	repo.On("ListUsers", mock.Anything, 1, "").Return(users, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/users?page=1", nil)
	rec := httptest.NewRecorder()
	h.ListUsers(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestChangeRoleHandler_Success(t *testing.T) {
	repo := &mockAdminRepo{}
	svc := admin.NewService(repo)
	h := admin.NewHandler(svc)

	userID := uuid.New()
	repo.On("ChangeUserRole", mock.Anything, userID, "moderator").Return(nil)

	r := chi.NewRouter()
	r.Patch("/admin/users/{id}/role", h.ChangeRole)

	body := `{"role":"moderator"}`
	req := httptest.NewRequest(http.MethodPatch, "/admin/users/"+userID.String()+"/role", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestChangeRoleHandler_InvalidRole_Returns400(t *testing.T) {
	repo := &mockAdminRepo{}
	svc := admin.NewService(repo)
	h := admin.NewHandler(svc)

	userID := uuid.New()
	r := chi.NewRouter()
	r.Patch("/admin/users/{id}/role", h.ChangeRole)

	body := `{"role":"superadmin"}`
	req := httptest.NewRequest(http.MethodPatch, "/admin/users/"+userID.String()+"/role", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetStatsHandler_Returns200(t *testing.T) {
	repo := &mockAdminRepo{}
	svc := admin.NewService(repo)
	h := admin.NewHandler(svc)

	stats := &admin.Stats{TotalUsers: 100, TotalBlogs: 500}
	repo.On("GetStats", mock.Anything).Return(stats, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/stats", nil)
	rec := httptest.NewRecorder()
	h.GetStats(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, float64(100), resp["TotalUsers"])
}

func TestListReportsHandler_Returns200(t *testing.T) {
	repo := &mockAdminRepo{}
	svc := admin.NewService(repo)
	h := admin.NewHandler(svc)

	reports := []*admin.ReportRow{{ID: uuid.New(), Reason: "spam", Status: "pending"}}
	repo.On("ListReports", mock.Anything, "pending", 1).Return(reports, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/reports?status=pending", nil)
	rec := httptest.NewRecorder()
	h.ListReports(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

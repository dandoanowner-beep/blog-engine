package notification_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"blog-engine/internal/middleware"
	"blog-engine/internal/notification"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- mocks ---

type mockNotifRepo struct{ mock.Mock }

func (m *mockNotifRepo) Create(ctx context.Context, n *notification.Notification) error {
	return m.Called(ctx, n).Error(0)
}
func (m *mockNotifRepo) ListForUser(ctx context.Context, userID uuid.UUID, page int) ([]*notification.Notification, int, error) {
	args := m.Called(ctx, userID, page)
	return args.Get(0).([]*notification.Notification), args.Int(1), args.Error(2)
}
func (m *mockNotifRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mockNotifRepo) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockNotifRepo) GetModsAndAdmins(ctx context.Context) ([]uuid.UUID, error) {
	args := m.Called(ctx)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func withUser(r *http.Request, userID uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.ContextKey("user_id"), userID)
	return r.WithContext(ctx)
}

var _ = middleware.UserIDFromContext

func TestListHandler_Returns200WithFields(t *testing.T) {
	repo := &mockNotifRepo{}
	svc := notification.NewService(repo)
	h := notification.NewHandler(svc)

	userID := uuid.New()
	actorID := uuid.New()
	items := []*notification.Notification{
		{ID: uuid.New(), UserID: userID, Type: "like_blog", ActorID: &actorID, Read: false, CreatedAt: time.Now()},
	}
	repo.On("ListForUser", mock.Anything, userID, 1).Return(items, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/notifications?page=1", nil)
	rec := httptest.NewRecorder()
	h.List(rec, withUser(req, userID))

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotNil(t, resp["notifications"])
	assert.Equal(t, float64(1), resp["unread_count"])
	assert.Equal(t, float64(1), resp["total"])
	assert.Equal(t, float64(1), resp["page"])
}

func TestMarkReadHandler_Returns200(t *testing.T) {
	repo := &mockNotifRepo{}
	svc := notification.NewService(repo)
	h := notification.NewHandler(svc)

	userID := uuid.New()
	notifID := uuid.New()
	repo.On("MarkRead", mock.Anything, notifID, userID).Return(nil)

	r := chi.NewRouter()
	r.Put("/notifications/{id}/read", func(w http.ResponseWriter, req *http.Request) {
		h.MarkRead(w, withUser(req, userID))
	})

	req := httptest.NewRequest(http.MethodPut, "/notifications/"+notifID.String()+"/read", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, true, resp["read"])
}

func TestMarkReadHandler_InvalidID_Returns400(t *testing.T) {
	repo := &mockNotifRepo{}
	svc := notification.NewService(repo)
	h := notification.NewHandler(svc)

	userID := uuid.New()
	r := chi.NewRouter()
	r.Put("/notifications/{id}/read", func(w http.ResponseWriter, req *http.Request) {
		h.MarkRead(w, withUser(req, userID))
	})

	req := httptest.NewRequest(http.MethodPut, "/notifications/not-a-uuid/read", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}

func TestMarkAllReadHandler_Returns200(t *testing.T) {
	repo := &mockNotifRepo{}
	svc := notification.NewService(repo)
	h := notification.NewHandler(svc)

	userID := uuid.New()
	repo.On("MarkAllRead", mock.Anything, userID).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/notifications/read-all", nil)
	rec := httptest.NewRecorder()
	h.MarkAllRead(rec, withUser(req, userID))

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "All marked as read", resp["message"])
}

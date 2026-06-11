package site_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"blog-engine/internal/site"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ site.ContentService = (*mockService)(nil)

type mockService struct{ mock.Mock }

func (m *mockService) GetAbout(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}
func (m *mockService) UpdateAbout(ctx context.Context, content string) error {
	return m.Called(ctx, content).Error(0)
}

func TestGetAboutHandler_Returns200WithContent(t *testing.T) {
	svc := &mockService{}
	h := site.NewHandler(svc)
	svc.On("GetAbout", mock.Anything).Return("<p>My story</p>", nil)

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()
	h.GetAbout(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Content string `json:"content"`
	}
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "<p>My story</p>", resp.Content)
}

func TestUpdateAboutHandler_Returns200(t *testing.T) {
	svc := &mockService{}
	h := site.NewHandler(svc)
	svc.On("UpdateAbout", mock.Anything, "<p>new</p>").Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/about", strings.NewReader(`{"content":"<p>new</p>"}`))
	rec := httptest.NewRecorder()
	h.UpdateAbout(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertCalled(t, "UpdateAbout", mock.Anything, "<p>new</p>")
}

func TestUpdateAboutHandler_BadBody_Returns400(t *testing.T) {
	svc := &mockService{}
	h := site.NewHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/about", strings.NewReader(`{not json`))
	rec := httptest.NewRecorder()
	h.UpdateAbout(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

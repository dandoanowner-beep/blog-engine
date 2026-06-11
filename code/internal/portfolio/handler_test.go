package portfolio_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"blog-engine/internal/portfolio"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ portfolio.ProjectService = (*mockService)(nil)

type mockService struct{ mock.Mock }

func (m *mockService) Create(ctx context.Context, input portfolio.CreateInput) (*portfolio.Project, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*portfolio.Project), args.Error(1)
}
func (m *mockService) Update(ctx context.Context, id uuid.UUID, input portfolio.UpdateInput) (*portfolio.Project, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*portfolio.Project), args.Error(1)
}
func (m *mockService) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockService) List(ctx context.Context) ([]*portfolio.Project, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*portfolio.Project), args.Error(1)
}

func TestListProjectsHandler_ReturnsContract(t *testing.T) {
	svc := &mockService{}
	h := portfolio.NewHandler(svc)

	svc.On("List", mock.Anything).Return([]*portfolio.Project{{
		ID:           uuid.New(),
		Title:        "Blog Engine",
		Description:  "<p>desc</p>",
		TechStack:    "Go, React",
		RepoURL:      "https://github.com/x/y",
		DemoURL:      "https://demo.example.com",
		ThumbnailURL: "https://img.example.com/t.png",
	}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	rec := httptest.NewRecorder()
	h.ListProjects(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Projects []struct {
			ID           string `json:"id"`
			Title        string `json:"title"`
			Description  string `json:"description"`
			TechStack    string `json:"tech_stack"`
			RepoURL      string `json:"repo_url"`
			DemoURL      string `json:"demo_url"`
			ThumbnailURL string `json:"thumbnail_url"`
		} `json:"projects"`
	}
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Len(t, resp.Projects, 1)
	assert.Equal(t, "Blog Engine", resp.Projects[0].Title)
	assert.Equal(t, "Go, React", resp.Projects[0].TechStack)
}

func TestListProjectsHandler_Empty_ReturnsEmptyArrayNotNull(t *testing.T) {
	svc := &mockService{}
	h := portfolio.NewHandler(svc)
	svc.On("List", mock.Anything).Return([]*portfolio.Project{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	rec := httptest.NewRecorder()
	h.ListProjects(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"projects":[]`)
}

func TestCreateProjectHandler_Returns201(t *testing.T) {
	svc := &mockService{}
	h := portfolio.NewHandler(svc)

	svc.On("Create", mock.Anything, mock.AnythingOfType("portfolio.CreateInput")).
		Return(&portfolio.Project{ID: uuid.New(), Title: "New"}, nil)

	body := `{"title":"New","description":"d","tech_stack":"Go","repo_url":"","demo_url":"","thumbnail_url":""}`
	req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.CreateProject(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreateProjectHandler_MissingTitle_Returns422(t *testing.T) {
	svc := &mockService{}
	h := portfolio.NewHandler(svc)
	svc.On("Create", mock.Anything, mock.Anything).Return(nil, portfolio.ErrMissingTitle)

	req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(`{"title":""}`))
	rec := httptest.NewRecorder()
	h.CreateProject(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestCreateProjectHandler_BadBody_Returns400(t *testing.T) {
	svc := &mockService{}
	h := portfolio.NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(`{not json`))
	rec := httptest.NewRecorder()
	h.CreateProject(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateProjectHandler_NotFound_Returns404(t *testing.T) {
	svc := &mockService{}
	h := portfolio.NewHandler(svc)
	id := uuid.New()
	svc.On("Update", mock.Anything, id, mock.Anything).Return(nil, portfolio.ErrNotFound)

	r := chi.NewRouter()
	r.Patch("/projects/{id}", h.UpdateProject)
	req := httptest.NewRequest(http.MethodPatch, "/projects/"+id.String(), strings.NewReader(`{"title":"x"}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateProjectHandler_InvalidID_Returns400(t *testing.T) {
	svc := &mockService{}
	h := portfolio.NewHandler(svc)

	r := chi.NewRouter()
	r.Patch("/projects/{id}", h.UpdateProject)
	req := httptest.NewRequest(http.MethodPatch, "/projects/not-a-uuid", strings.NewReader(`{"title":"x"}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleteProjectHandler_Returns204(t *testing.T) {
	svc := &mockService{}
	h := portfolio.NewHandler(svc)
	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(nil)

	r := chi.NewRouter()
	r.Delete("/projects/{id}", h.DeleteProject)
	req := httptest.NewRequest(http.MethodDelete, "/projects/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

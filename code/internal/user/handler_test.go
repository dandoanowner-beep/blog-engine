package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-engine/internal/middleware"
	"blog-engine/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProfileHandler_Returns200(t *testing.T) {
	repo := &mockRepo{}
	svc := user.NewService(repo)
	h := user.NewHandler(svc)

	profileID := uuid.New()
	profile := &user.Profile{ID: profileID, Username: "alice", Bio: "Hello"}
	repo.On("GetByUsername", mock.Anything, "alice").Return(profile, nil)
	repo.On("IsFriend", mock.Anything, uuid.Nil, profileID).Return(false, nil)

	r := chi.NewRouter()
	r.Get("/users/{username}", h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/users/alice", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotNil(t, resp["user"])
}

func TestGetProfileHandler_NotFound_Returns404(t *testing.T) {
	repo := &mockRepo{}
	svc := user.NewService(repo)
	h := user.NewHandler(svc)

	repo.On("GetByUsername", mock.Anything, "ghost").Return(nil, user.ErrNotFound)

	r := chi.NewRouter()
	r.Get("/users/{username}", h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/users/ghost", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}

func TestUpdateProfileHandler_Success(t *testing.T) {
	repo := &mockRepo{}
	svc := user.NewService(repo)
	h := user.NewHandler(svc)

	userID := uuid.New()
	updated := &user.Profile{ID: userID, Bio: "Updated bio"}
	input := user.UpdateInput{Bio: strPtr("Updated bio")}

	repo.On("Update", mock.Anything, userID, input).Return(updated, nil)

	r := chi.NewRouter()
	r.Patch("/users/me", func(w http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), middleware.ContextKey("user_id"), userID)
		h.UpdateProfile(w, req.WithContext(ctx))
	})

	body, _ := json.Marshal(map[string]string{"bio": "Updated bio"})
	req := httptest.NewRequest(http.MethodPatch, "/users/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotNil(t, resp["user"])
}

func TestUpdateProfileHandler_DuplicateUsername_Returns409(t *testing.T) {
	repo := &mockRepo{}
	svc := user.NewService(repo)
	h := user.NewHandler(svc)

	userID := uuid.New()
	taken := "taken"
	repo.On("UsernameExists", mock.Anything, taken, userID).Return(true, nil)

	r := chi.NewRouter()
	r.Patch("/users/me", func(w http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), middleware.ContextKey("user_id"), userID)
		h.UpdateProfile(w, req.WithContext(ctx))
	})

	body, _ := json.Marshal(map[string]string{"username": "taken"})
	req := httptest.NewRequest(http.MethodPatch, "/users/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}


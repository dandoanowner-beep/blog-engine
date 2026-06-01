package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"blog-engine/internal/auth"
	"blog-engine/internal/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func makeJWT(t *testing.T, role string) string {
	t.Helper()
	j := auth.NewJWT("test-secret", "test-refresh-secret")
	tok, err := j.GenerateAccessToken(uuid.New(), role)
	assert.NoError(t, err)
	return tok
}

func TestAuthMiddleware_ValidToken_SetsContext(t *testing.T) {
	j := auth.NewJWT("test-secret", "test-refresh-secret")
	m := middleware.NewAuth(j)

	called := false
	handler := m.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		userID := middleware.UserIDFromContext(r.Context())
		assert.NotEqual(t, uuid.Nil, userID)
		w.WriteHeader(http.StatusOK)
	}))

	tok := makeJWT(t, "user")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_MissingToken_Returns401(t *testing.T) {
	j := auth.NewJWT("test-secret", "test-refresh-secret")
	m := middleware.NewAuth(j)

	handler := m.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_ExpiredToken_Returns401(t *testing.T) {
	j := auth.NewJWT("test-secret", "test-refresh-secret")
	m := middleware.NewAuth(j)
	tok, _ := j.GenerateAccessTokenWithExpiry(uuid.New(), "user", time.Now().Add(-time.Minute))

	handler := m.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRBAC_AdminCanAccessAdminRoute(t *testing.T) {
	j := auth.NewJWT("test-secret", "test-refresh-secret")
	authM := middleware.NewAuth(j)
	rbacM := middleware.NewRBAC()

	handler := authM.Authenticate(rbacM.RequireRole("admin", "owner")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	))

	tok := makeJWT(t, "admin")
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRBAC_UserCannotAccessAdminRoute(t *testing.T) {
	j := auth.NewJWT("test-secret", "test-refresh-secret")
	authM := middleware.NewAuth(j)
	rbacM := middleware.NewRBAC()

	handler := authM.Authenticate(rbacM.RequireRole("admin", "owner")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	))

	tok := makeJWT(t, "user")
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRBAC_ModeratorCanAccessModeratorRoute(t *testing.T) {
	j := auth.NewJWT("test-secret", "test-refresh-secret")
	authM := middleware.NewAuth(j)
	rbacM := middleware.NewRBAC()

	handler := authM.Authenticate(rbacM.RequireRole("moderator", "admin", "owner")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	))

	tok := makeJWT(t, "moderator")
	req := httptest.NewRequest("GET", "/mod", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

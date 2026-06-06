package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"blog-engine/internal/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- mock service ---

type mockAuthService struct{ mock.Mock }

func (m *mockAuthService) Register(ctx context.Context, email, username, password string) (*auth.User, error) {
	args := m.Called(ctx, email, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}
func (m *mockAuthService) Login(ctx context.Context, email, password string) (*auth.TokenPair, *auth.User, error) {
	args := m.Called(ctx, email, password)
	var tp *auth.TokenPair
	var u *auth.User
	if args.Get(0) != nil {
		tp = args.Get(0).(*auth.TokenPair)
	}
	if args.Get(1) != nil {
		u = args.Get(1).(*auth.User)
	}
	return tp, u, args.Error(2)
}
func (m *mockAuthService) VerifyEmail(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}
func (m *mockAuthService) ForgotPassword(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}
func (m *mockAuthService) ResetPassword(ctx context.Context, token, password string) error {
	return m.Called(ctx, token, password).Error(0)
}
func (m *mockAuthService) BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	return m.Called(ctx, blockerID, blockedID).Error(0)
}
func (m *mockAuthService) UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	return m.Called(ctx, blockerID, blockedID).Error(0)
}
func (m *mockAuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.Error(1)
}

// ════════════════════════════════════════════════════════════
// Handler tests
// ════════════════════════════════════════════════════════════

func TestRegisterHandler_Success(t *testing.T) {
	svc := &mockAuthService{}
	h := auth.NewHandler(svc)

	user := &auth.User{ID: uuid.New(), Email: "test@example.com", Username: "testuser", Verified: false}
	svc.On("Register", mock.Anything, "test@example.com", "testuser", "password123").Return(user, nil)

	body := `{"email":"test@example.com","username":"testuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "Verification email sent", resp["message"])
}

func TestRegisterHandler_DuplicateEmail_Returns400(t *testing.T) {
	svc := &mockAuthService{}
	h := auth.NewHandler(svc)

	svc.On("Register", mock.Anything, "dup@example.com", "user", "pass1234").Return(nil, auth.ErrEmailTaken)

	body := `{"email":"dup@example.com","username":"user","password":"pass1234"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegisterHandler_InvalidJSON_Returns400(t *testing.T) {
	svc := &mockAuthService{}
	h := auth.NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLoginHandler_Success(t *testing.T) {
	svc := &mockAuthService{}
	h := auth.NewHandler(svc)

	tokens := &auth.TokenPair{AccessToken: "access.tok.en", RefreshToken: "refresh.tok.en"}
	user := &auth.User{ID: uuid.New(), Username: "testuser", Email: "test@example.com", Role: "user", Verified: true}
	svc.On("Login", mock.Anything, "test@example.com", "password123").Return(tokens, user, nil)

	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "access.tok.en", resp["access_token"])
	userMap, ok := resp["user"].(map[string]interface{})
	assert.True(t, ok, "response must contain a user object")
	assert.NotEmpty(t, userMap["id"])
	assert.Equal(t, "testuser", userMap["username"])
	assert.Equal(t, "test@example.com", userMap["email"])
	assert.Equal(t, "user", userMap["role"])
	assert.Equal(t, true, userMap["verified"])
}

func TestLoginHandler_InvalidCredentials_Returns401(t *testing.T) {
	svc := &mockAuthService{}
	h := auth.NewHandler(svc)

	svc.On("Login", mock.Anything, "x@x.com", "wrong").Return(nil, nil, auth.ErrInvalidCredentials)

	body := `{"email":"x@x.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLoginHandler_LockedAccount_Returns423(t *testing.T) {
	svc := &mockAuthService{}
	h := auth.NewHandler(svc)

	svc.On("Login", mock.Anything, "locked@x.com", "pass").Return(nil, nil, auth.ErrAccountLocked)

	body := `{"email":"locked@x.com","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)
	assert.Equal(t, http.StatusLocked, rec.Code)
}

func TestVerifyEmailHandler_Success(t *testing.T) {
	svc := &mockAuthService{}
	h := auth.NewHandler(svc)

	svc.On("VerifyEmail", mock.Anything, "valid-token").Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/verify?token=valid-token", nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestVerifyEmailHandler_ExpiredToken_Returns400(t *testing.T) {
	svc := &mockAuthService{}
	h := auth.NewHandler(svc)

	svc.On("VerifyEmail", mock.Anything, "bad-token").Return(auth.ErrTokenExpired)

	req := httptest.NewRequest(http.MethodGet, "/auth/verify?token=bad-token", nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestForgotPasswordHandler_AlwaysReturns200(t *testing.T) {
	svc := &mockAuthService{}
	h := auth.NewHandler(svc)

	// Even for unknown emails — must not leak
	svc.On("ForgotPassword", mock.Anything, "any@x.com").Return(nil)

	body := `{"email":"any@x.com"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ForgotPassword(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHealthHandler_Returns200(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	auth.HealthHandler(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ensure unused import doesn't break
var _ = time.Now

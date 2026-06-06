package auth_test

import (
	"context"
	"testing"
	"time"

	"blog-engine/internal/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

var _ auth.Repository = (*mockRepo)(nil)
var _ auth.EmailSender = (*mockEmail)(nil)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) CreateUser(ctx context.Context, u *auth.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockRepo) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}
func (m *mockRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}
func (m *mockRepo) GetUserByGoogleID(ctx context.Context, googleID string) (*auth.User, error) {
	args := m.Called(ctx, googleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}
func (m *mockRepo) UpdateUser(ctx context.Context, u *auth.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockRepo) SaveVerificationToken(ctx context.Context, userID uuid.UUID, token string, exp time.Time) error {
	return m.Called(ctx, userID, token, exp).Error(0)
}
func (m *mockRepo) GetVerificationToken(ctx context.Context, token string) (*auth.VerificationToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.VerificationToken), args.Error(1)
}
func (m *mockRepo) MarkTokenUsed(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}
func (m *mockRepo) SavePasswordReset(ctx context.Context, userID uuid.UUID, token string, exp time.Time) error {
	return m.Called(ctx, userID, token, exp).Error(0)
}
func (m *mockRepo) GetPasswordReset(ctx context.Context, token string) (*auth.PasswordResetToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.PasswordResetToken), args.Error(1)
}
func (m *mockRepo) IncrementLoginAttempts(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}
func (m *mockRepo) LockAccount(ctx context.Context, email string, until time.Time) error {
	return m.Called(ctx, email, until).Error(0)
}
func (m *mockRepo) ResetLoginAttempts(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}
func (m *mockRepo) BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	return m.Called(ctx, blockerID, blockedID).Error(0)
}
func (m *mockRepo) UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	return m.Called(ctx, blockerID, blockedID).Error(0)
}

type mockEmail struct{ mock.Mock }

func (m *mockEmail) SendVerification(to, token string) error {
	return m.Called(to, token).Error(0)
}
func (m *mockEmail) SendPasswordReset(to, token string) error {
	return m.Called(to, token).Error(0)
}

// --- Tests: AC-AUTH-001 ---

func TestRegister_Success(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	repo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(nil, auth.ErrNotFound)
	repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)
	repo.On("SaveVerificationToken", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
	email.On("SendVerification", "test@example.com", mock.AnythingOfType("string")).Return(nil)

	user, err := svc.Register(context.Background(), "test@example.com", "testuser", "password123")
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "user", user.Role)
	assert.False(t, user.Verified)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	existing := &auth.User{ID: uuid.New(), Email: "test@example.com"}
	repo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(existing, nil)

	_, err := svc.Register(context.Background(), "test@example.com", "testuser", "password123")
	assert.ErrorIs(t, err, auth.ErrEmailTaken)
}

func TestRegister_PasswordTooShort(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	_, err := svc.Register(context.Background(), "test@example.com", "testuser", "short")
	assert.ErrorIs(t, err, auth.ErrPasswordTooShort)
}

// --- Tests: AC-AUTH-003 ---

func TestVerifyEmail_Success(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	userID := uuid.New()
	tok := &auth.VerificationToken{
		Token:     "valid-token",
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour),
		Used:      false,
	}
	user := &auth.User{ID: userID, Email: "test@example.com", Verified: false}

	repo.On("GetVerificationToken", mock.Anything, "valid-token").Return(tok, nil)
	repo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	repo.On("UpdateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)
	repo.On("MarkTokenUsed", mock.Anything, "valid-token").Return(nil)

	err := svc.VerifyEmail(context.Background(), "valid-token")
	assert.NoError(t, err)
}

func TestVerifyEmail_ExpiredToken(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	tok := &auth.VerificationToken{
		Token:     "expired-token",
		UserID:    uuid.New(),
		ExpiresAt: time.Now().Add(-time.Hour),
		Used:      false,
	}
	repo.On("GetVerificationToken", mock.Anything, "expired-token").Return(tok, nil)

	err := svc.VerifyEmail(context.Background(), "expired-token")
	assert.ErrorIs(t, err, auth.ErrTokenExpired)
}

// --- Tests: AC-AUTH-004 ---

func TestLogin_Success(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	user := &auth.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Role:     "user",
		Verified: true,
	}
	// pre-hash a known password
	hash, _ := auth.HashPassword("password123")
	user.PasswordHash = hash

	repo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	repo.On("ResetLoginAttempts", mock.Anything, "test@example.com").Return(nil)

	tokens, user, err := svc.Login(context.Background(), "test@example.com", "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	hash, _ := auth.HashPassword("correctpassword")
	user := &auth.User{
		ID: uuid.New(), Email: "test@example.com",
		PasswordHash: hash, Role: "user", Verified: true,
		LoginAttempts: 0,
	}

	repo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	repo.On("IncrementLoginAttempts", mock.Anything, "test@example.com").Return(nil)

	_, _, err := svc.Login(context.Background(), "test@example.com", "wrongpassword")
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestLogin_LockedAccount(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	lockedUntil := time.Now().Add(10 * time.Minute)
	user := &auth.User{
		ID: uuid.New(), Email: "test@example.com",
		Role: "user", Verified: true, LockedUntil: &lockedUntil,
	}

	repo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)

	_, _, err := svc.Login(context.Background(), "test@example.com", "anypassword")
	assert.ErrorIs(t, err, auth.ErrAccountLocked)
}

func TestLogin_FifthFailedAttemptLocksAccount(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	hash, _ := auth.HashPassword("correctpassword")
	user := &auth.User{
		ID: uuid.New(), Email: "test@example.com",
		PasswordHash: hash, Role: "user", Verified: true, LoginAttempts: 4,
	}

	repo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	repo.On("IncrementLoginAttempts", mock.Anything, "test@example.com").Return(nil)
	repo.On("LockAccount", mock.Anything, "test@example.com", mock.Anything).Return(nil)

	_, _, err := svc.Login(context.Background(), "test@example.com", "wrongpassword")
	assert.ErrorIs(t, err, auth.ErrAccountLocked)
	repo.AssertCalled(t, "LockAccount", mock.Anything, "test@example.com", mock.Anything)
}

// --- Tests: AC-AUTH-005 ---

func TestPasswordReset_Success(t *testing.T) {
	repo := &mockRepo{}
	emailSvc := &mockEmail{}
	svc := auth.NewService(repo, emailSvc, "jwt-secret", "refresh-secret", "http://app.test")

	userID := uuid.New()
	user := &auth.User{ID: userID, Email: "test@example.com"}
	repo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	repo.On("SavePasswordReset", mock.Anything, userID, mock.AnythingOfType("string"), mock.Anything).Return(nil)
	emailSvc.On("SendPasswordReset", "test@example.com", mock.AnythingOfType("string")).Return(nil)

	err := svc.ForgotPassword(context.Background(), "test@example.com")
	assert.NoError(t, err)
}

func TestPasswordReset_UnknownEmail_NoError(t *testing.T) {
	// Security: should not reveal whether email exists
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	repo.On("GetUserByEmail", mock.Anything, "unknown@example.com").Return(nil, auth.ErrNotFound)

	err := svc.ForgotPassword(context.Background(), "unknown@example.com")
	assert.NoError(t, err) // must not leak whether email exists
}

// --- Tests: AC-AUTH-006 (unverified cannot publish) ---

func TestUnverifiedUser_CannotPublish(t *testing.T) {
	user := &auth.User{ID: uuid.New(), Verified: false, Role: "user"}
	err := auth.AssertVerified(user)
	assert.ErrorIs(t, err, auth.ErrNotVerified)
}

func TestVerifiedUser_CanPublish(t *testing.T) {
	user := &auth.User{ID: uuid.New(), Verified: true, Role: "user"}
	err := auth.AssertVerified(user)
	assert.NoError(t, err)
}

// --- Tests: AC-AUTH-006 (block) ---

func TestBlockUser_MutualBlind(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	blockerID := uuid.New()
	blockedID := uuid.New()

	repo.On("BlockUser", mock.Anything, blockerID, blockedID).Return(nil)

	err := svc.BlockUser(context.Background(), blockerID, blockedID)
	assert.NoError(t, err)
}

func TestBlockUser_CannotBlockSelf(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	id := uuid.New()
	err := svc.BlockUser(context.Background(), id, id)
	assert.ErrorIs(t, err, auth.ErrCannotBlockSelf)
}

func TestResetPassword_Success(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	userID := uuid.New()
	tok := &auth.PasswordResetToken{Token: "reset-token", UserID: userID, ExpiresAt: time.Now().Add(time.Hour), Used: false}
	user := &auth.User{ID: userID, Email: "test@example.com", PasswordHash: "old"}

	repo.On("GetPasswordReset", mock.Anything, "reset-token").Return(tok, nil)
	repo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	repo.On("UpdateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)
	repo.On("MarkTokenUsed", mock.Anything, "reset-token").Return(nil)

	err := svc.ResetPassword(context.Background(), "reset-token", "newpassword123")
	assert.NoError(t, err)
}

func TestResetPassword_ExpiredToken(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	tok := &auth.PasswordResetToken{Token: "expired-token", UserID: uuid.New(), ExpiresAt: time.Now().Add(-time.Hour), Used: false}
	repo.On("GetPasswordReset", mock.Anything, "expired-token").Return(tok, nil)

	err := svc.ResetPassword(context.Background(), "expired-token", "newpassword123")
	assert.ErrorIs(t, err, auth.ErrTokenExpired)
}

func TestResetPassword_PasswordTooShort(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	err := svc.ResetPassword(context.Background(), "any-token", "short")
	assert.ErrorIs(t, err, auth.ErrPasswordTooShort)
}

func TestRefreshToken_InvalidToken_ReturnsError(t *testing.T) {
	repo := &mockRepo{}
	email := &mockEmail{}
	svc := auth.NewService(repo, email, "jwt-secret", "refresh-secret", "http://app.test")

	// An invalid token exercises the ValidateRefreshToken error path
	_, err := svc.RefreshToken(context.Background(), "not.a.valid.jwt")
	assert.Error(t, err)
}

func TestRefreshToken_RoundTrip(t *testing.T) {
	repo := &mockRepo{}
	emailSvc := &mockEmail{}
	svc := auth.NewService(repo, emailSvc, "jwt-secret", "refresh-secret", "http://app.test")

	userID := uuid.New()
	hash, _ := auth.HashPassword("password123")
	user := &auth.User{ID: userID, Email: "rt@test.com", Role: "user", PasswordHash: hash}

	repo.On("GetUserByEmail", mock.Anything, "rt@test.com").Return(user, nil)
	repo.On("ResetLoginAttempts", mock.Anything, "rt@test.com").Return(nil)
	repo.On("GetUserByID", mock.Anything, userID).Return(user, nil)

	tokens, _, err := svc.Login(context.Background(), "rt@test.com", "password123")
	assert.NoError(t, err)

	newAccess, err := svc.RefreshToken(context.Background(), tokens.RefreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newAccess)
}

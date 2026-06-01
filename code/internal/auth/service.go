package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	CreateUser(ctx context.Context, u *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*User, error)
	UpdateUser(ctx context.Context, u *User) error
	SaveVerificationToken(ctx context.Context, userID uuid.UUID, token string, exp time.Time) error
	GetVerificationToken(ctx context.Context, token string) (*VerificationToken, error)
	MarkTokenUsed(ctx context.Context, token string) error
	SavePasswordReset(ctx context.Context, userID uuid.UUID, token string, exp time.Time) error
	GetPasswordReset(ctx context.Context, token string) (*PasswordResetToken, error)
	IncrementLoginAttempts(ctx context.Context, email string) error
	LockAccount(ctx context.Context, email string, until time.Time) error
	ResetLoginAttempts(ctx context.Context, email string) error
	BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error
	UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error
}

type EmailSender interface {
	SendVerification(to, token string) error
	SendPasswordReset(to, token string) error
}

type Service struct {
	repo          Repository
	email         EmailSender
	jwt           *JWT
	appURL        string
}

func NewService(repo Repository, email EmailSender, jwtSecret, refreshSecret, appURL string) *Service {
	return &Service{
		repo:   repo,
		email:  email,
		jwt:    NewJWT(jwtSecret, refreshSecret),
		appURL: appURL,
	}
}

func (s *Service) Register(ctx context.Context, email, username, password string) (*User, error) {
	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}
	existing, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil && err != ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           uuid.New(),
		Email:        email,
		Username:     username,
		PasswordHash: hash,
		Role:         "user",
		Verified:     false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	token := generateToken()
	exp := time.Now().Add(24 * time.Hour)
	if err := s.repo.SaveVerificationToken(ctx, user.ID, token, exp); err != nil {
		return nil, err
	}
	_ = s.email.SendVerification(email, token)
	return user, nil
}

func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	tok, err := s.repo.GetVerificationToken(ctx, token)
	if err != nil {
		return ErrTokenInvalid
	}
	if time.Now().After(tok.ExpiresAt) {
		return ErrTokenExpired
	}
	if tok.Used {
		return ErrTokenInvalid
	}
	user, err := s.repo.GetUserByID(ctx, tok.UserID)
	if err != nil {
		return err
	}
	user.Verified = true
	user.UpdatedAt = time.Now()
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}
	return s.repo.MarkTokenUsed(ctx, token)
}

func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, ErrAccountLocked
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		_ = s.repo.IncrementLoginAttempts(ctx, email)
		if user.LoginAttempts+1 >= 5 {
			lockUntil := time.Now().Add(15 * time.Minute)
			_ = s.repo.LockAccount(ctx, email, lockUntil)
			return nil, ErrAccountLocked
		}
		return nil, ErrInvalidCredentials
	}
	_ = s.repo.ResetLoginAttempts(ctx, email)

	access, err := s.jwt.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	refresh, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil // do not leak whether email exists
	}
	token := generateToken()
	exp := time.Now().Add(time.Hour)
	if err := s.repo.SavePasswordReset(ctx, user.ID, token, exp); err != nil {
		return err
	}
	_ = s.email.SendPasswordReset(email, token)
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}
	tok, err := s.repo.GetPasswordReset(ctx, token)
	if err != nil {
		return ErrTokenInvalid
	}
	if time.Now().After(tok.ExpiresAt) || tok.Used {
		return ErrTokenExpired
	}
	user, err := s.repo.GetUserByID(ctx, tok.UserID)
	if err != nil {
		return err
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	user.UpdatedAt = time.Now()
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}
	return s.repo.MarkTokenUsed(ctx, token)
}

func (s *Service) UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	if blockerID == blockedID {
		return ErrCannotBlockSelf
	}
	return s.repo.UnblockUser(ctx, blockerID, blockedID)
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	userID, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return s.jwt.GenerateAccessToken(user.ID, user.Role)
}

func (s *Service) BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	if blockerID == blockedID {
		return ErrCannotBlockSelf
	}
	return s.repo.BlockUser(ctx, blockerID, blockedID)
}

func AssertVerified(user *User) error {
	if !user.Verified {
		return ErrNotVerified
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func generateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b)
}

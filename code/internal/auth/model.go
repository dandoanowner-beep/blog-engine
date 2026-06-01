package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	GoogleID     string
	Role         string
	Verified     bool
	AvatarURL    string
	Bio          string
	FavoriteQuote string
	LoginAttempts int
	LockedUntil  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type VerificationToken struct {
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
	Used      bool
}

type PasswordResetToken struct {
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
	Used      bool
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

package auth_test

import (
	"testing"
	"time"

	"blog-engine/internal/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	j := auth.NewJWT("secret", "refresh-secret")
	userID := uuid.New()

	token, err := j.GenerateAccessToken(userID, "user")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := j.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, "user", claims.Role)
}

func TestAccessToken_Expiry(t *testing.T) {
	j := auth.NewJWT("secret", "refresh-secret")
	userID := uuid.New()

	// generate token that expired 1 second ago
	token, err := j.GenerateAccessTokenWithExpiry(userID, "user", time.Now().Add(-time.Second))
	assert.NoError(t, err)

	_, err = j.ValidateAccessToken(token)
	assert.ErrorIs(t, err, auth.ErrTokenExpired)
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	j := auth.NewJWT("secret", "refresh-secret")
	userID := uuid.New()

	token, err := j.GenerateRefreshToken(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	id, err := j.ValidateRefreshToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, id)
}

func TestInvalidToken_ReturnsError(t *testing.T) {
	j := auth.NewJWT("secret", "refresh-secret")

	_, err := j.ValidateAccessToken("not.a.valid.token")
	assert.Error(t, err)
}

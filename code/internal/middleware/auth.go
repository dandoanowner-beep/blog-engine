package middleware

import (
	"context"
	"net/http"
	"strings"

	"blog-engine/internal/auth"
	"github.com/google/uuid"
)

type contextKey string

// ContextKey is the exported type for injecting auth values in tests.
type ContextKey = contextKey

const (
	keyUserID contextKey = "user_id"
	keyRole   contextKey = "role"
)

type Auth struct {
	jwt *auth.JWT
}

func NewAuth(jwt *auth.JWT) *Auth {
	return &Auth{jwt: jwt}
}

func (a *Auth) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearer(r)
		if token == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		claims, err := a.jwt.ValidateAccessToken(token)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), keyUserID, claims.UserID)
		ctx = context.WithValue(ctx, keyRole, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(keyUserID).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

func RoleFromContext(ctx context.Context) string {
	if role, ok := ctx.Value(keyRole).(string); ok {
		return role
	}
	return ""
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

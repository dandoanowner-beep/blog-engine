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

// OptionalAuthenticate is for guest-allowed routes (e.g. public blog reads).
// No Authorization header → request proceeds as guest with no identity in context.
// A token that IS present must be valid — an invalid or expired token is rejected
// with 401 before any controller logic, never silently downgraded to guest.
func (a *Auth) OptionalAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearer(r)
		if token == "" {
			next.ServeHTTP(w, r)
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

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"mini-ecommerce/pkg/jwt"
)

type contextKey string

const UserIDContextKey contextKey = "user_id"

func Auth(tokenManager *jwt.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				respondUnauthorized(w, "missing or invalid authorization header")
				return
			}
			claims, err := tokenManager.Validate(strings.TrimPrefix(authHeader, "Bearer "))
			if err != nil {
				respondUnauthorized(w, "invalid or expired token")
				return
			}
			ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": message})
}

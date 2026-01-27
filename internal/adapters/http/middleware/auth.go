package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/golang-jwt/jwt/v5"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserEmailKey is the context key for authenticated user email
	UserEmailKey contextKey = "user_email"
	// UserRoleKey is the context key for authenticated user role
	UserRoleKey contextKey = "user_role"
)

// Claims represents JWT token claims
type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// Auth creates middleware that validates JWT tokens
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				sendUnauthorized(w, "Missing Authorization header")
				return
			}

			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				sendUnauthorized(w, "Invalid Authorization header format (expected 'Bearer <token>')")
				return
			}

			tokenString := parts[1]

			// Parse and validate token
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				sendUnauthorized(w, fmt.Sprintf("Invalid token: %v", err))
				return
			}

			if !token.Valid {
				sendUnauthorized(w, "Token is not valid")
				return
			}

			// Check expiration
			if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
				sendUnauthorized(w, "Token has expired")
				return
			}

			// Add user email and role to request context
			ctx := context.WithValue(r.Context(), UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserEmail extracts authenticated user email from request context
func GetUserEmail(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(UserEmailKey).(string)
	return email, ok
}

// GetUserRole extracts authenticated user role from request context
func GetUserRole(r *http.Request) (string, bool) {
	role, ok := r.Context().Value(UserRoleKey).(string)
	return role, ok
}

func sendUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	resp := dto.ErrorResponse{
		Error:   "Unauthorized",
		Message: message,
	}
	json.NewEncoder(w).Encode(resp)
}

package middleware

import (
	"net/http"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

// RequireAdmin verifies the authenticated user has Admin role
// Assumes Auth middleware has already run and populated context
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(UserRoleKey).(string)
		if !ok {
			// No role found? Should have been set by Auth middleware
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if role != string(domain.RoleAdmin) {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

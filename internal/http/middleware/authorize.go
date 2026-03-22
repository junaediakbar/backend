package middleware

import (
	"net/http"

	"laundry-backend/internal/httpapi"
)

func RequireRole(allowed ...string) func(http.Handler) http.Handler {
	set := map[string]struct{}{}
	for _, r := range allowed {
		set[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r.Context())
			if !ok {
				httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", nil)
				return
			}
			if _, ok := set[claims.Role]; !ok {
				httpapi.WriteError(w, http.StatusForbidden, "forbidden", "Forbidden", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

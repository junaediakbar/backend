package middleware

import (
	"net/http"
	"strings"

	"laundry-backend/internal/httpapi"
)

// RequireNotEmployee blocks users with role "employee" (UI/staff-only API).
func RequireNotEmployee(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaims(r.Context())
		if !ok {
			httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", nil)
			return
		}
		if claims.Role == "employee" {
			httpapi.WriteError(w, http.StatusForbidden, "forbidden", "Akses ditolak", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// OrderEmployeeReadOnly allows employees to use GET/HEAD on /orders, and PUT on
// .../work-assignments/{taskType} (isi Performa Karyawan). Metode lain tetap ditolak.
func OrderEmployeeReadOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaims(r.Context())
		if !ok {
			httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", nil)
			return
		}
		if claims.Role != "employee" {
			next.ServeHTTP(w, r)
			return
		}
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/work-assignments/") {
			next.ServeHTTP(w, r)
			return
		}
		httpapi.WriteError(w, http.StatusForbidden, "forbidden", "Akses ditolak", nil)
	})
}

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

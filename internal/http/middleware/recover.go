package middleware

import (
	"net/http"

	"laundry-backend/internal/httpapi"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				httpapi.WriteError(w, http.StatusInternalServerError, "panic", "Terjadi kesalahan internal", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

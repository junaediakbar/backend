package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			next.ServeHTTP(ww, r)

			d := time.Since(start)
			method := r.Method
			path := r.URL.Path
			status := ww.Status()
			bytes := ww.BytesWritten()
			reqID := chimw.GetReqID(r.Context())

			user := ""
			if c, ok := GetClaims(r.Context()); ok {
				user = c.Email
				if strings.TrimSpace(c.Role) != "" {
					user = fmt.Sprintf("%s(%s)", user, c.Role)
				}
			}

			fmt.Fprintf(os.Stdout, "http %s %s status=%d bytes=%d dur=%s req_id=%s user=%s\n",
				method, path, status, bytes, d.Round(time.Millisecond), reqID, user,
			)
		})
	}
}

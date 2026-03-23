package httpapi

import (
	"errors"
	"log"
	"net/http"
	"runtime/debug"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		reqID := chimw.GetReqID(r.Context())
		var appErr *AppError
		if errors.As(err, &appErr) {
			log.Printf("http_error req_id=%s method=%s path=%s status=%d code=%s msg=%s err=%v", reqID, r.Method, r.URL.Path, appErr.Status, appErr.Code, appErr.Message, err)
			WriteError(w, appErr.Status, appErr.Code, appErr.Message, appErr.Details)
			return
		}
		log.Printf("http_error req_id=%s method=%s path=%s status=%d err=%v stack=%s", reqID, r.Method, r.URL.Path, http.StatusInternalServerError, err, string(debug.Stack()))
		WriteError(w, http.StatusInternalServerError, "internal", "Terjadi kesalahan internal", nil)
	}
}

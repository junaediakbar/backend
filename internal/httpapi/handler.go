package httpapi

import (
	"errors"
	"net/http"
)

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		var appErr *AppError
		if errors.As(err, &appErr) {
			WriteError(w, appErr.Status, appErr.Code, appErr.Message, appErr.Details)
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal", "Terjadi kesalahan internal", nil)
	}
}

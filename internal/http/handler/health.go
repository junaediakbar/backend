package handler

import (
	"net/http"

	"laundry-backend/internal/httpapi"
)

func Health() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		httpapi.WriteOK(w, http.StatusOK, map[string]string{"status": "ok"})
		return nil
	})
}

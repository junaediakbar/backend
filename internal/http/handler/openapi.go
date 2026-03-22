package handler

import (
	"net/http"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/openapi"
)

func OpenAPIJSON() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		data, err := openapi.JSON()
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
		return nil
	})
}

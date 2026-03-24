package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/repository"
	"laundry-backend/internal/service"
)

type EmployeeHandler struct {
	svc *service.EmployeeService
}

func NewEmployeeHandler(svc *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

func (h *EmployeeHandler) List() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		active, err := parseBoolQuery(r, "active")
		if err != nil {
			return err
		}
		out, err := h.svc.List(r.Context(), active)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

func (h *EmployeeHandler) Get() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "id")
		out, err := h.svc.Get(r.Context(), id)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

type employeeBody struct {
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
}

func (h *EmployeeHandler) Create() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		var body employeeBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}
		out, err := h.svc.Create(r.Context(), repository.CreateEmployeeParams{Name: strings.TrimSpace(body.Name), IsActive: body.IsActive})
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusCreated, out)
		return nil
	})
}

func (h *EmployeeHandler) Update() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "id")
		var body employeeBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}
		out, err := h.svc.Update(r.Context(), id, repository.UpdateEmployeeParams{Name: strings.TrimSpace(body.Name), IsActive: body.IsActive})
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

func (h *EmployeeHandler) Delete() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "id")
		if err := h.svc.Delete(r.Context(), id); err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, map[string]bool{"ok": true})
		return nil
	})
}

func (h *EmployeeHandler) Performance() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		start, err := parseDateQuery(r, "startDate", false)
		if err != nil {
			return err
		}
		end, err := parseDateQuery(r, "endDate", true)
		if err != nil {
			return err
		}
		out, err := h.svc.Performance(r.Context(), start, end)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

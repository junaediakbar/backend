package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"laundry-backend/internal/http/middleware"
	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/model"
	"laundry-backend/internal/service"
)

type EmployeeHandler struct {
	svc *service.EmployeeService
	loc *time.Location
}

func NewEmployeeHandler(svc *service.EmployeeService, loc *time.Location) *EmployeeHandler {
	if loc == nil {
		loc = time.UTC
	}
	return &EmployeeHandler{svc: svc, loc: loc}
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
		// Karyawan: cukup id + nama (+ status) untuk dropdown penugasan; sembunyikan email & role.
		if c, ok := middleware.GetClaims(r.Context()); ok && strings.EqualFold(strings.TrimSpace(c.Role), "employee") {
			sanitized := make([]model.Employee, len(out))
			for i, e := range out {
				sanitized[i] = model.Employee{
					ID:        e.ID,
					Name:      e.Name,
					Email:     "",
					Role:      "",
					IsActive:  e.IsActive,
					CreatedAt: e.CreatedAt,
					UpdatedAt: e.UpdatedAt,
				}
			}
			httpapi.WriteOK(w, http.StatusOK, sanitized)
			return nil
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

type employeeCreateBody struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	IsActive bool   `json:"isActive"`
}

type employeeUpdateBody struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Role     string  `json:"role"`
	Password *string `json:"password"`
	IsActive bool    `json:"isActive"`
}

func (h *EmployeeHandler) Create() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		var body employeeCreateBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}
		out, err := h.svc.Create(r.Context(), service.CreateEmployeeInput{
			Name:     body.Name,
			Email:    body.Email,
			Password: body.Password,
			Role:     body.Role,
			IsActive: body.IsActive,
		})
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
		var body employeeUpdateBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}
		out, err := h.svc.Update(r.Context(), id, service.UpdateEmployeeInput{
			Name:     body.Name,
			Email:    body.Email,
			Role:     body.Role,
			IsActive: body.IsActive,
			Password: body.Password,
		})
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
		c, ok := middleware.GetClaims(r.Context())
		if !ok {
			return httpapi.Unauthorized("Unauthorized")
		}
		start, err := parseDateQuery(r, "startDate", false, h.loc)
		if err != nil {
			return err
		}
		end, err := parseDateQuery(r, "endDate", true, h.loc)
		if err != nil {
			return err
		}
		var filter *string
		if c.Role == "employee" {
			eid := strings.TrimSpace(c.EmployeeID)
			if eid == "" {
				eid = strings.TrimSpace(c.UserID)
			}
			if eid == "" {
				return httpapi.BadRequest("validation_error", "Akun karyawan belum tertaut ke data karyawan", nil)
			}
			filter = &eid
		} else {
			q := strings.TrimSpace(r.URL.Query().Get("employeeId"))
			if q != "" {
				filter = &q
			}
		}
		out, err := h.svc.Performance(r.Context(), start, end, filter)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

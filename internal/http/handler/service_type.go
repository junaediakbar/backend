package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/repository"
	"laundry-backend/internal/service"
	"laundry-backend/internal/util"
)

type ServiceTypeHandler struct {
	svc *service.ServiceTypeService
}

func NewServiceTypeHandler(svc *service.ServiceTypeService) *ServiceTypeHandler {
	return &ServiceTypeHandler{svc: svc}
}

func (h *ServiceTypeHandler) List() http.Handler {
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

func (h *ServiceTypeHandler) Get() http.Handler {
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

type serviceTypeBody struct {
	Name         string  `json:"name"`
	Unit         string  `json:"unit"`
	DefaultPrice float64 `json:"defaultPrice"`
	IsActive     bool    `json:"isActive"`
}

func (h *ServiceTypeHandler) Create() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		var body serviceTypeBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}
		out, err := h.svc.Create(r.Context(), repository.CreateServiceTypeParams{
			Name:         strings.TrimSpace(body.Name),
			Unit:         strings.TrimSpace(body.Unit),
			DefaultPrice: util.Money2(body.DefaultPrice),
			IsActive:     body.IsActive,
		})
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusCreated, out)
		return nil
	})
}

func (h *ServiceTypeHandler) Update() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "id")
		var body serviceTypeBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}
		out, err := h.svc.Update(r.Context(), id, repository.UpdateServiceTypeParams{
			Name:         strings.TrimSpace(body.Name),
			Unit:         strings.TrimSpace(body.Unit),
			DefaultPrice: util.Money2(body.DefaultPrice),
			IsActive:     body.IsActive,
		})
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

func (h *ServiceTypeHandler) Delete() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "id")
		if err := h.svc.Delete(r.Context(), id); err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, map[string]bool{"ok": true})
		return nil
	})
}

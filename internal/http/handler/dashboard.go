package handler

import (
	"net/http"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/service"
)

type DashboardHandler struct {
	svc *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) Summary() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		start, err := parseDateQuery(r, "startDate", false)
		if err != nil {
			return err
		}
		end, err := parseDateQuery(r, "endDate", true)
		if err != nil {
			return err
		}
		out, err := h.svc.Summary(r.Context(), start, end)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

func (h *DashboardHandler) RevenueSeries() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		start, err := parseDateQuery(r, "startDate", false)
		if err != nil {
			return err
		}
		end, err := parseDateQuery(r, "endDate", true)
		if err != nil {
			return err
		}
		if start == nil || end == nil {
			return httpapi.BadRequest("validation_error", "startDate dan endDate wajib diisi", nil)
		}
		out, err := h.svc.RevenueSeries(r.Context(), *start, *end)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

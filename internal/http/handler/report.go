package handler

import (
	"net/http"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/service"
)

type ReportHandler struct {
	svc *service.ReportService
}

func NewReportHandler(svc *service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

func (h *ReportHandler) OrdersCSV() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		start, err := parseDateQuery(r, "startDate", false)
		if err != nil {
			return err
		}
		end, err := parseDateQuery(r, "endDate", true)
		if err != nil {
			return err
		}

		data, filename, err := h.svc.OrdersCSV(r.Context(), start, end)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
		return nil
	})
}

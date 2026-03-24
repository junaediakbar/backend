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

type DeliveryHandler struct {
	svc *service.DeliveryService
}

func NewDeliveryHandler(svc *service.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{svc: svc}
}

func (h *DeliveryHandler) ListPlans() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		limit := parseIntQuery(r, "limit", 50)
		out, err := h.svc.ListPlans(r.Context(), limit)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

func (h *DeliveryHandler) GetPlan() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "id")
		out, err := h.svc.GetPlan(r.Context(), id)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

type createPlanBody struct {
	Name         string  `json:"name"`
	PlannedDate  string  `json:"plannedDate"`
	StartAddress *string `json:"startAddress"`
	StartLat     float64 `json:"startLat"`
	StartLng     float64 `json:"startLng"`
	Stops        []struct {
		CustomerID string  `json:"customerId"`
		Sequence   int     `json:"sequence"`
		DistanceKm float64 `json:"distanceKm"`
	} `json:"stops"`
}

func (h *DeliveryHandler) CreatePlan() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		var body createPlanBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}

		plannedDate, ok := parseDateOnly(body.PlannedDate)
		if !ok {
			return httpapi.BadRequest("validation_error", "Tanggal tidak valid", nil)
		}

		stops := make([]repository.CreateStopParams, 0, len(body.Stops))
		for _, s0 := range body.Stops {
			customerID := strings.TrimSpace(s0.CustomerID)
			if customerID == "" || s0.Sequence <= 0 {
				continue
			}
			stops = append(stops, repository.CreateStopParams{
				CustomerID: customerID,
				Sequence:   s0.Sequence,
				DistanceKm: util.Money2(s0.DistanceKm),
			})
		}

		out, err := h.svc.CreatePlan(r.Context(), service.CreatePlanInput{
			Name:         body.Name,
			PlannedDate:  plannedDate,
			StartAddress: trimNotePtr(body.StartAddress),
			StartLat:     body.StartLat,
			StartLng:     body.StartLng,
			Stops:        stops,
		})
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusCreated, out)
		return nil
	})
}

func (h *DeliveryHandler) DeletePlan() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "id")
		if err := h.svc.DeletePlan(r.Context(), id); err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, map[string]bool{"ok": true})
		return nil
	})
}

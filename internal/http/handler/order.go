package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) List() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		page := parseIntQuery(r, "page", 1)
		pageSize := parseIntQuery(r, "pageSize", 20)
		out, err := h.svc.List(r.Context(), q, page, pageSize)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

func (h *OrderHandler) Get() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "id")
		out, err := h.svc.GetDetail(r.Context(), id)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

type createOrderBody struct {
	CustomerID    string                `json:"customerId"`
	ReceivedDate  *string               `json:"receivedDate"`
	CompletedDate *string               `json:"completedDate"`
	Note          *string               `json:"note"`
	Items         []createOrderItemBody `json:"items"`
}

type createOrderItemBody struct {
	ServiceTypeID string  `json:"serviceTypeId"`
	Quantity      float64 `json:"quantity"`
	UnitPrice     float64 `json:"unitPrice"`
	Discount      float64 `json:"discount"`
}

func (h *OrderHandler) Create() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		var body createOrderBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}

		receivedDate := time.Now()
		if body.ReceivedDate != nil && strings.TrimSpace(*body.ReceivedDate) != "" {
			if t, ok := parseDateOnly(*body.ReceivedDate); ok {
				receivedDate = t
			}
		}

		var completedDate *time.Time
		if body.CompletedDate != nil && strings.TrimSpace(*body.CompletedDate) != "" {
			if t, ok := parseDateOnly(*body.CompletedDate); ok {
				completedDate = &t
			}
		}

		items := make([]service.CreateOrderItemInput, 0, len(body.Items))
		for _, it := range body.Items {
			items = append(items, service.CreateOrderItemInput{
				ServiceTypeID: it.ServiceTypeID,
				Quantity:      it.Quantity,
				UnitPrice:     it.UnitPrice,
				Discount:      it.Discount,
			})
		}

		out, err := h.svc.Create(r.Context(), service.CreateOrderInput{
			CustomerID:    body.CustomerID,
			ReceivedDate:  receivedDate,
			CompletedDate: completedDate,
			Note:          trimNotePtr(body.Note),
			Items:         items,
		})
		if err != nil {
			return err
		}

		httpapi.WriteOK(w, http.StatusCreated, out)
		return nil
	})
}

type workflowBody struct {
	WorkflowStatus string `json:"workflowStatus"`
}

func (h *OrderHandler) UpdateWorkflow() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		orderID := chi.URLParam(r, "id")
		var body workflowBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}
		if err := h.svc.UpdateWorkflow(r.Context(), orderID, body.WorkflowStatus); err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, map[string]bool{"ok": true})
		return nil
	})
}

type paymentBody struct {
	Amount float64 `json:"amount"`
	Method string  `json:"method"`
	Note   *string `json:"note"`
}

func (h *OrderHandler) CreatePayment() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		orderID := chi.URLParam(r, "id")
		var body paymentBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}
		out, err := h.svc.CreatePayment(r.Context(), orderID, service.CreatePaymentInput{
			Amount: body.Amount,
			Method: body.Method,
			Note:   trimNotePtr(body.Note),
		})
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusCreated, out)
		return nil
	})
}

type workAssignmentBody struct {
	EmployeeID string `json:"employeeId"`
}

func (h *OrderHandler) UpsertWorkAssignment() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		orderID := chi.URLParam(r, "orderId")
		orderItemID := chi.URLParam(r, "orderItemId")
		taskType := chi.URLParam(r, "taskType")

		percent, ok := taskPercent(taskType)
		if !ok {
			return httpapi.BadRequest("validation_error", "Task type tidak valid", nil)
		}

		var body workAssignmentBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}

		if err := h.svc.UpsertWorkAssignment(r.Context(), service.UpsertWorkAssignmentInput{
			OrderID:     orderID,
			OrderItemID: orderItemID,
			TaskType:    taskType,
			EmployeeID:  body.EmployeeID,
			Percent:     percent,
		}); err != nil {
			return err
		}

		httpapi.WriteOK(w, http.StatusOK, map[string]bool{"ok": true})
		return nil
	})
}

type attachmentsBody struct {
	Files []struct {
		FilePath  string  `json:"filePath"`
		MimeType  *string `json:"mimeType"`
		SizeBytes *int    `json:"sizeBytes"`
	} `json:"files"`
}

func (h *OrderHandler) CreateAttachments() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		orderID := chi.URLParam(r, "id")
		var body attachmentsBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}
		files := make([]service.CreateAttachmentInput, 0, len(body.Files))
		for _, f := range body.Files {
			files = append(files, service.CreateAttachmentInput{
				FilePath:  f.FilePath,
				MimeType:  trimNotePtr(f.MimeType),
				SizeBytes: f.SizeBytes,
			})
		}
		if err := h.svc.CreateAttachments(r.Context(), orderID, files); err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, map[string]bool{"ok": true})
		return nil
	})
}

func parseDateOnly(value string) (time.Time, bool) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func trimNotePtr(p *string) *string {
	if p == nil {
		return nil
	}
	s := strings.TrimSpace(*p)
	if s == "" {
		return nil
	}
	return &s
}

func taskPercent(taskType string) (float64, bool) {
	switch taskType {
	case "pickup", "dropoff", "fuel_vehicle", "driver", "dust_removal", "brushing", "rinse_sprayer", "spin_dry":
		return 5, true
	case "finishing_packing":
		return 10, true
	default:
		return 0, false
	}
}

package handler

import (
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/model"
	"laundry-backend/internal/service"
	"laundry-backend/internal/util"
)

type OrderHandler struct {
	svc *service.OrderService
	loc *time.Location
}

func NewOrderHandler(svc *service.OrderService, loc *time.Location) *OrderHandler {
	if loc == nil {
		loc = time.UTC
	}
	return &OrderHandler{svc: svc, loc: loc}
}

func keepWallClock(t time.Time, loc *time.Location) time.Time {
	if loc == nil {
		return t
	}
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

func keepWallClockPtr(t *time.Time, loc *time.Location) *time.Time {
	if t == nil {
		return nil
	}
	v := keepWallClock(*t, loc)
	return &v
}

func (h *OrderHandler) normalizeOrderTimes(o *model.OrderDetail) {
	if o == nil {
		return
	}
	o.ReceivedDate = keepWallClock(o.ReceivedDate, h.loc)
	o.CompletedDate = keepWallClockPtr(o.CompletedDate, h.loc)
	o.PickupDate = keepWallClockPtr(o.PickupDate, h.loc)
	o.CreatedAt = keepWallClock(o.CreatedAt, h.loc)
	o.UpdatedAt = keepWallClock(o.UpdatedAt, h.loc)

	for i := range o.Items {
		o.Items[i].CreatedAt = keepWallClock(o.Items[i].CreatedAt, h.loc)
		o.Items[i].UpdatedAt = keepWallClock(o.Items[i].UpdatedAt, h.loc)
		for j := range o.Items[i].WorkAssignments {
			o.Items[i].WorkAssignments[j].CreatedAt = keepWallClock(o.Items[i].WorkAssignments[j].CreatedAt, h.loc)
		}
	}
	for i := range o.Payments {
		o.Payments[i].PaidAt = keepWallClock(o.Payments[i].PaidAt, h.loc)
		o.Payments[i].CreatedAt = keepWallClock(o.Payments[i].CreatedAt, h.loc)
	}
	for i := range o.Attachments {
		o.Attachments[i].CreatedAt = keepWallClock(o.Attachments[i].CreatedAt, h.loc)
	}
}

func (h *OrderHandler) List() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		page := parseIntQuery(r, "page", 1)
		pageSize := parseIntQuery(r, "pageSize", 20)
		sort := strings.TrimSpace(r.URL.Query().Get("sort"))
		dir := strings.TrimSpace(r.URL.Query().Get("dir"))
		startDate, err := parseDateQuery(r, "startDate", false)
		if err != nil {
			return err
		}
		endDate, err := parseDateQuery(r, "endDate", true)
		if err != nil {
			return err
		}
		out, err := h.svc.List(r.Context(), q, page, pageSize, sort, dir, startDate, endDate)
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
		h.normalizeOrderTimes(out)
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

func (h *OrderHandler) Delete() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "id")
		if err := h.svc.Delete(r.Context(), id); err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, map[string]bool{"ok": true})
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
		var imageParts [][]byte

		ct := r.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "multipart/form-data") {
			if err := r.ParseMultipartForm(80 << 20); err != nil {
				return httpapi.BadRequest("invalid_multipart", "Form upload tidak valid", nil)
			}

			body.CustomerID = r.FormValue("customerId")
			if v := strings.TrimSpace(r.FormValue("receivedDate")); v != "" {
				body.ReceivedDate = &v
			}
			if v := strings.TrimSpace(r.FormValue("completedDate")); v != "" {
				body.CompletedDate = &v
			}
			if v := strings.TrimSpace(r.FormValue("note")); v != "" {
				body.Note = &v
			}

			itemsRaw := strings.TrimSpace(r.FormValue("items"))
			if itemsRaw != "" {
				if err := json.Unmarshal([]byte(itemsRaw), &body.Items); err != nil {
					return httpapi.BadRequest("validation_error", "Items tidak valid", nil)
				}
			}

			if r.MultipartForm != nil {
				nFiles := len(r.MultipartForm.File["images"]) + len(r.MultipartForm.File["image"])
				if nFiles > 3 {
					return httpapi.BadRequest("validation_error", "Maksimal 3 gambar", nil)
				}
			}
			parts, err := collectMultipartOrderImages(r)
			if err != nil {
				return err
			}
			imageParts = parts
		} else {
			if err := decodeJSON(r, &body); err != nil {
				return err
			}
		}

		receivedDate := time.Now().In(h.loc)
		if body.ReceivedDate != nil && strings.TrimSpace(*body.ReceivedDate) != "" {
			if t, ok := parseOrderDateTime(*body.ReceivedDate, h.loc); ok {
				receivedDate = t
			}
		}

		var completedDate *time.Time
		if body.CompletedDate != nil && strings.TrimSpace(*body.CompletedDate) != "" {
			if t, ok := parseOrderDateTime(*body.CompletedDate, h.loc); ok {
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

		if len(imageParts) > 0 {
			urls := make([]string, 0, len(imageParts))
			for i, imageBytes := range imageParts {
				processed, err := util.ProcessImageToJPEGMaxBytes(imageBytes, 5<<20)
				if err != nil {
					return httpapi.BadRequest("validation_error", "Gagal memproses gambar", nil)
				}

				imageURL, err := util.UploadOrderImageToCloudinary(r.Context(), out.ID, i, processed.Bytes)
				if err != nil {
					return httpapi.Internal("Gagal upload gambar")
				}
				log.Printf("order_image_set order_id=%s image_url_prefix=%s", out.ID, imageURLPrefix(imageURL))
				urls = append(urls, imageURL)
			}
			encoded := util.EncodeOrderImagesJSON(urls)
			if encoded != nil {
				if err := h.svc.UpdateImage(r.Context(), out.ID, encoded); err != nil {
					return err
				}
				first, all := util.NormalizeOrderImageColumn(encoded)
				out.Image = first
				out.Images = all
			}
		}

		h.normalizeOrderTimes(out)
		httpapi.WriteOK(w, http.StatusCreated, out)
		return nil
	})
}

func imageURLPrefix(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 60 {
		return s
	}
	return s[:60]
}

func readLimited(r io.Reader, max int64) ([]byte, error) {
	lr := io.LimitReader(r, max+1)
	b, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, httpapi.BadRequest("validation_error", "File terlalu besar", nil)
	}
	return b, nil
}

type workflowBody struct {
	WorkflowStatus string `json:"workflowStatus"`
}

func (h *OrderHandler) UpdateWorkflow() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		orderID := chi.URLParam(r, "id")
		var body workflowBody
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			workflowStatus := strings.TrimSpace(r.URL.Query().Get("workflowStatus"))
			if workflowStatus == "" {
				return httpapi.BadRequest("invalid_json", "Body JSON tidak valid", map[string]string{"detail": err.Error()})
			}
			body.WorkflowStatus = workflowStatus
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

func (h *OrderHandler) DeletePayment() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		orderID := chi.URLParam(r, "id")
		paymentID := chi.URLParam(r, "paymentId")
		out, err := h.svc.DeletePayment(r.Context(), orderID, paymentID)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

type workAssignmentBody struct {
	EmployeeID string `json:"employeeId"`
	Percent    *float64 `json:"percent"`
}

func (h *OrderHandler) UpsertWorkAssignment() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		orderID := chi.URLParam(r, "orderId")
		orderItemID := chi.URLParam(r, "orderItemId")
		taskType := chi.URLParam(r, "taskType")

		var body workAssignmentBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}

		employeeID := strings.TrimSpace(body.EmployeeID)
		percent := 0.0
		if employeeID != "" {
			if body.Percent != nil && *body.Percent > 0 {
				percent = *body.Percent
			} else if p, ok := taskPercent(taskType); ok {
				percent = p
			} else {
				return httpapi.BadRequest("validation_error", "Percent wajib diisi untuk task baru", nil)
			}
		}

		if err := h.svc.UpsertWorkAssignment(r.Context(), service.UpsertWorkAssignmentInput{
			OrderID:     orderID,
			OrderItemID: orderItemID,
			TaskType:    taskType,
			EmployeeID:  employeeID,
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

// parseOrderDateTime accepts date-only (YYYY-MM-DD), HTML datetime-local (YYYY-MM-DDTHH:MM), RFC3339, etc.
func parseOrderDateTime(value string, loc *time.Location) (time.Time, bool) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.In(loc), true
	}
	layouts := []string{
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, loc); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func collectMultipartOrderImages(r *http.Request) ([][]byte, error) {
	const maxFiles = 3
	const maxEach = 25 << 20
	var out [][]byte
	if r.MultipartForm == nil {
		return out, nil
	}
	appendOne := func(fh *multipart.FileHeader) error {
		if len(out) >= maxFiles {
			return nil
		}
		f, err := fh.Open()
		if err != nil {
			return err
		}
		defer f.Close()
		b, err := readLimited(f, maxEach)
		if err != nil {
			return err
		}
		if len(b) > 0 {
			out = append(out, b)
		}
		return nil
	}
	for _, fh := range r.MultipartForm.File["images"] {
		if err := appendOne(fh); err != nil {
			return nil, err
		}
	}
	if len(out) < maxFiles {
		for _, fh := range r.MultipartForm.File["image"] {
			if err := appendOne(fh); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
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
	case
		"pickup_fuel",
		"pickup_driver",
		"pickup_worker_1",
		"pickup_worker_2",
		"dropoff_fuel",
		"dropoff_driver",
		"dropoff_worker_1",
		"dropoff_worker_2":
		return 2.5, true
	case "dust_removal", "brushing", "rinse_sprayer", "spin_dry":
		return 5, true
	case "finishing_packing":
		return 10, true
	default:
		return 0, false
	}
}

package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/repository"
	"laundry-backend/internal/util"
)

type PublicReceiptHandler struct {
	orders repository.OrderRepository
	loc    *time.Location
}

func NewPublicReceiptHandler(orders repository.OrderRepository, loc *time.Location) *PublicReceiptHandler {
	if loc == nil {
		loc = time.UTC
	}
	return &PublicReceiptHandler{orders: orders, loc: loc}
}

type publicReceiptItem struct {
	ServiceName string  `json:"serviceName"`
	Unit        string  `json:"unit"`
	Quantity    string  `json:"quantity"`
	UnitPrice   string  `json:"unitPrice"`
	Discount    string  `json:"discount"`
	Total       string  `json:"total"`
	LengthM     *string `json:"lengthM,omitempty"`
	WidthM      *string `json:"widthM,omitempty"`
}

type PublicReceipt struct {
	PublicToken    string              `json:"publicToken"`
	InvoiceNumber  string              `json:"invoiceNumber"`
	CustomerName   string              `json:"customerName"`
	CustomerPhone  *string             `json:"customerPhone,omitempty"`
	Total          string              `json:"total"`
	PaidAmount     string              `json:"paidAmount"`
	PaymentStatus  string              `json:"paymentStatus"`
	WorkflowStatus string              `json:"workflowStatus"`
	ReceivedDate   time.Time           `json:"receivedDate"`
	CompletedDate  *time.Time          `json:"completedDate,omitempty"`
	PickupDate     *time.Time          `json:"pickupDate,omitempty"`
	Image          *string             `json:"image,omitempty"`
	Images         []string            `json:"images,omitempty"`
	Note           *string             `json:"note,omitempty"`
	Items          []publicReceiptItem `json:"items"`
}

func (h *PublicReceiptHandler) Get() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		token := strings.TrimSpace(chi.URLParam(r, "token"))
		if token == "" {
			return httpapi.BadRequest("validation_error", "Token wajib diisi", nil)
		}

		out, err := h.orders.GetDetailByPublicToken(r.Context(), token)
		if err != nil {
			if err == pgx.ErrNoRows {
				return httpapi.NotFound("Nota tidak ditemukan")
			}
			return err
		}

		paid := 0.0
		for _, p := range out.Payments {
			paid += parseMoneyFloat(p.Amount)
		}

		items := make([]publicReceiptItem, 0, len(out.Items))
		for _, it := range out.Items {
			items = append(items, publicReceiptItem{
				ServiceName: it.ServiceType.Name,
				Unit:        it.ServiceType.Unit,
				Quantity:    it.Quantity,
				UnitPrice:   it.UnitPrice,
				Discount:    it.Discount,
				Total:       it.Total,
				LengthM:     it.LengthM,
				WidthM:      it.WidthM,
			})
		}

		resp := PublicReceipt{
			PublicToken:    out.PublicToken,
			InvoiceNumber:  out.InvoiceNumber,
			CustomerName:   out.Customer.Name,
			CustomerPhone:  out.Customer.Phone,
			Total:          out.Total,
			PaidAmount:     util.Money2(paid),
			PaymentStatus:  out.PaymentStatus,
			WorkflowStatus: out.WorkflowStatus,
			ReceivedDate:   keepWallClock(out.ReceivedDate, h.loc),
			CompletedDate:  keepWallClockPtr(out.CompletedDate, h.loc),
			PickupDate:     keepWallClockPtr(out.PickupDate, h.loc),
			Image:          out.Image,
			Images:         out.Images,
			Note:           out.Note,
			Items:          items,
		}

		httpapi.WriteOK(w, http.StatusOK, resp)
		return nil
	})
}

func parseMoneyFloat(v string) float64 {
	n, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return 0
	}
	return n
}

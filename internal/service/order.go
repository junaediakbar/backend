package service

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
	"laundry-backend/internal/util"
	"laundry-backend/internal/workflow"
)

type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) List(ctx context.Context, q string, page, pageSize int, sort string, dir string, startDate, endDate *time.Time) (model.Paged[model.OrderListItem], error) {
	return s.repo.List(ctx, q, page, pageSize, sort, dir, startDate, endDate)
}

func (s *OrderService) GetDetail(ctx context.Context, id string) (*model.OrderDetail, error) {
	out, err := s.repo.GetDetail(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.NotFound("Nota tidak ditemukan")
		}
		return nil, err
	}
	return out, nil
}

func (s *OrderService) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return httpapi.BadRequest("validation_error", "ID tidak valid", nil)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if err == pgx.ErrNoRows {
			return httpapi.NotFound("Nota tidak ditemukan")
		}
		return err
	}
	return nil
}

type CreateOrderItemInput struct {
	ServiceTypeID string
	Quantity      float64
	UnitPrice     float64
	Discount      float64
}

type CreateOrderInput struct {
	CustomerID    string
	ReceivedDate  time.Time
	CompletedDate *time.Time
	Image         *string
	Note          *string
	Items         []CreateOrderItemInput
}

func (s *OrderService) Create(ctx context.Context, in CreateOrderInput) (*model.OrderDetail, error) {
	in.CustomerID = strings.TrimSpace(in.CustomerID)
	if in.CustomerID == "" {
		return nil, httpapi.BadRequest("validation_error", "Pelanggan wajib dipilih", nil)
	}
	if len(in.Items) == 0 {
		return nil, httpapi.BadRequest("validation_error", "Minimal 1 item pesanan", nil)
	}

	items := make([]repository.CreateOrderItemParams, 0, len(in.Items))
	for _, it := range in.Items {
		serviceTypeID := strings.TrimSpace(it.ServiceTypeID)
		if serviceTypeID == "" {
			return nil, httpapi.BadRequest("validation_error", "Layanan wajib dipilih", nil)
		}
		if it.Quantity <= 0 {
			return nil, httpapi.BadRequest("validation_error", "Qty harus lebih dari 0", nil)
		}
		if it.UnitPrice < 0 || it.Discount < 0 {
			return nil, httpapi.BadRequest("validation_error", "Harga/diskon tidak valid", nil)
		}
		total := it.Quantity*it.UnitPrice - it.Discount
		if total < 0 {
			total = 0
		}
		items = append(items, repository.CreateOrderItemParams{
			ServiceTypeID: serviceTypeID,
			Quantity:      util.Money2(it.Quantity),
			UnitPrice:     util.Money2(it.UnitPrice),
			Discount:      util.Money2(it.Discount),
			Total:         util.Money2(total),
		})
	}

	return s.repo.Create(ctx, repository.CreateOrderParams{
		CustomerID:    in.CustomerID,
		ReceivedDate:  in.ReceivedDate,
		CompletedDate: in.CompletedDate,
		Image:         in.Image,
		Note:          in.Note,
		Items:         items,
	})
}

func (s *OrderService) UpdateImage(ctx context.Context, orderID string, image *string) error {
	if err := s.repo.UpdateImage(ctx, orderID, image); err != nil {
		if err == pgx.ErrNoRows {
			return httpapi.NotFound("Nota tidak ditemukan")
		}
		return err
	}
	return nil
}

func (s *OrderService) UpdateWorkflow(ctx context.Context, orderID string, workflowStatus string) error {
	workflowStatus = strings.TrimSpace(workflowStatus)
	allowed := map[string]bool{
		"received":     true,
		"rontok_done":  true,
		"jemur_done":   true,
		"downy_done":   true,
		"packing_done": true,
		"delivered":    true,
		"picked_up":    true,
		// legacy
		"washing":   true,
		"drying":    true,
		"ironing":   true,
		"finished":  true,
	}
	if !allowed[workflowStatus] {
		return httpapi.BadRequest("validation_error", "Workflow status tidak valid", nil)
	}
	if err := s.repo.UpdateWorkflow(ctx, orderID, workflowStatus); err != nil {
		if err == pgx.ErrNoRows {
			return httpapi.NotFound("Nota tidak ditemukan")
		}
		return err
	}
	return nil
}

type CreatePaymentInput struct {
	Amount float64
	Method string
	Note   *string
}

func (s *OrderService) CreatePayment(ctx context.Context, orderID string, in CreatePaymentInput) (*model.Payment, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return nil, httpapi.BadRequest("validation_error", "ID nota tidak valid", nil)
	}
	in.Method = strings.ToLower(strings.TrimSpace(in.Method))
	if in.Amount <= 0 || in.Method == "" {
		return nil, httpapi.BadRequest("validation_error", "Nominal dan metode wajib diisi", nil)
	}
	if in.Method != "cash" && in.Method != "qris" && in.Method != "lainnya" {
		return nil, httpapi.BadRequest("validation_error", "Metode harus cash, qris, atau lainnya", nil)
	}
	out, err := s.repo.CreatePayment(ctx, orderID, repository.CreatePaymentParams{
		Amount: util.Money2(in.Amount),
		Method: in.Method,
		Note:   in.Note,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.NotFound("Nota tidak ditemukan")
		}
		return nil, err
	}
	return out, nil
}

func (s *OrderService) DeletePayment(ctx context.Context, orderID string, paymentID string) (*model.Payment, error) {
	orderID = strings.TrimSpace(orderID)
	paymentID = strings.TrimSpace(paymentID)
	if orderID == "" || paymentID == "" {
		return nil, httpapi.BadRequest("validation_error", "ID tidak valid", nil)
	}
	out, err := s.repo.DeletePayment(ctx, orderID, paymentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.NotFound("Pembayaran tidak ditemukan")
		}
		return nil, err
	}
	return out, nil
}

type UpsertWorkAssignmentInput struct {
	OrderID     string
	OrderItemID string
	TaskType    string
	EmployeeID  string
	Percent     float64
}

func (s *OrderService) UpsertWorkAssignment(ctx context.Context, in UpsertWorkAssignmentInput) error {
	in.OrderID = strings.TrimSpace(in.OrderID)
	in.OrderItemID = strings.TrimSpace(in.OrderItemID)
	in.TaskType = strings.TrimSpace(in.TaskType)
	in.EmployeeID = strings.TrimSpace(in.EmployeeID)
	if in.OrderID == "" || in.OrderItemID == "" || in.TaskType == "" {
		return httpapi.BadRequest("validation_error", "Input tidak valid", nil)
	}
	if in.EmployeeID == "" {
		if err := s.repo.DeleteWorkAssignment(ctx, in.OrderItemID, in.TaskType); err != nil {
			return err
		}
		return nil
	}
	if in.Percent < 0 {
		return httpapi.BadRequest("validation_error", "Percent tidak valid", nil)
	}

	detail, err := s.repo.GetDetail(ctx, in.OrderID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return httpapi.NotFound("Nota tidak ditemukan")
		}
		return err
	}
	var item *model.OrderItem
	for i := range detail.Items {
		if detail.Items[i].ID == in.OrderItemID {
			item = &detail.Items[i]
			break
		}
	}
	if item == nil {
		return httpapi.BadRequest("validation_error", "Item nota tidak ditemukan", nil)
	}
	canon := workflow.NormalizeTask(in.TaskType)
	if !workflow.CanAssignTask(item.WorkAssignments, canon) {
		return httpapi.BadRequest("validation_error", "Isi tahap produksi sebelumnya terlebih dahulu (rontok opsional boleh dilewati).", nil)
	}

	if err := s.repo.UpsertWorkAssignment(ctx, repository.UpsertWorkAssignmentParams{
		OrderID:     in.OrderID,
		OrderItemID: in.OrderItemID,
		TaskType:    in.TaskType,
		EmployeeID:  in.EmployeeID,
		Percent:     util.Money2(in.Percent),
		Amount:      "0.00",
	}); err != nil {
		return err
	}

	after, err := s.repo.GetDetail(ctx, in.OrderID)
	if err != nil {
		return err
	}
	target := workflow.TargetFromAssignments(after)
	cur := after.WorkflowStatus
	if workflow.Rank(target) > workflow.Rank(cur) && workflow.Rank(target) < workflow.Rank(workflow.PickedUp) {
		_ = s.repo.UpdateWorkflow(ctx, in.OrderID, target)
	}
	return nil
}

type CreateAttachmentInput struct {
	FilePath  string
	MimeType  *string
	SizeBytes *int
}

func (s *OrderService) CreateAttachments(ctx context.Context, orderID string, files []CreateAttachmentInput) error {
	out := make([]repository.CreateAttachmentParams, 0, len(files))
	for _, f := range files {
		fp := strings.TrimSpace(f.FilePath)
		if fp == "" {
			continue
		}
		out = append(out, repository.CreateAttachmentParams{FilePath: fp, MimeType: f.MimeType, SizeBytes: f.SizeBytes})
	}
	if len(out) == 0 {
		return httpapi.BadRequest("validation_error", "No files", nil)
	}
	if err := s.repo.CreateAttachments(ctx, orderID, out); err != nil {
		return err
	}
	return nil
}

package pg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lucsky/cuid"

	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type OrderRepo struct {
	db *DB
}

func NewOrderRepo(db *DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) List(ctx context.Context, q string, page, pageSize int, sort string, dir string) (model.Paged[model.OrderListItem], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	offset := (page - 1) * pageSize

	where := "true"
	args := []any{}
	if strings.TrimSpace(q) != "" {
		where = `(
			LOWER(o.invoice_number) LIKE '%' || LOWER($1) || '%'
			OR LOWER(c.name) LIKE '%' || LOWER($1) || '%'
			OR LOWER(COALESCE(c.phone, '')) LIKE '%' || LOWER($1) || '%'
			OR LOWER(COALESCE(c.email, '')) LIKE '%' || LOWER($1) || '%'
			OR LOWER(o.id) LIKE '%' || LOWER($1) || '%'
		)`
		args = append(args, strings.TrimSpace(q))
	}

	sortKey := strings.ToLower(strings.TrimSpace(sort))
	dirKey := strings.ToLower(strings.TrimSpace(dir))
	if dirKey != "asc" {
		dirKey = "desc"
	}
	orderBy := "o.created_at"
	switch sortKey {
	case "created_at":
		orderBy = "o.created_at"
	case "received_date":
		orderBy = "o.received_date"
	case "total":
		orderBy = "o.total"
	case "invoice_number":
		orderBy = "o.invoice_number"
	case "customer_name":
		orderBy = "c.name"
	}
	orderClause := fmt.Sprintf("%s %s", orderBy, dirKey)

	var total int
	if err := r.db.Pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM laundry_backend.orders o
		JOIN laundry_backend.customers c ON c.id = o.customer_id
		WHERE %s
	`, where), args...).Scan(&total); err != nil {
		return model.Paged[model.OrderListItem]{}, err
	}

	argsList := append([]any{}, args...)
	argsList = append(argsList, pageSize, offset)
	rows, err := r.db.Pool.Query(ctx, fmt.Sprintf(`
		SELECT
			o.id,
			o.invoice_number,
			o.public_token,
			c.id,
			c.name,
			o.total::text,
			o.payment_status::text,
			o.workflow_status::text,
			o.created_at,
			COALESCE(cnt.item_count, 0) AS item_count,
			fi.service_type_id,
			fi.service_type_name
		FROM laundry_backend.orders o
		JOIN laundry_backend.customers c ON c.id = o.customer_id
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS item_count
			FROM laundry_backend.order_items oi
			WHERE oi.order_id = o.id
		) cnt ON true
		LEFT JOIN LATERAL (
			SELECT oi.service_type_id, st.name AS service_type_name
			FROM laundry_backend.order_items oi
			JOIN laundry_backend.service_types st ON st.id = oi.service_type_id
			WHERE oi.order_id = o.id
			ORDER BY oi.created_at ASC
			LIMIT 1
		) fi ON true
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, where, orderClause, len(args)+1, len(args)+2), argsList...)
	if err != nil {
		return model.Paged[model.OrderListItem]{}, err
	}
	defer rows.Close()

	out := make([]model.OrderListItem, 0, pageSize)
	for rows.Next() {
		var item model.OrderListItem
		var stID *string
		var stName *string
		if err := rows.Scan(
			&item.ID,
			&item.InvoiceNumber,
			&item.PublicToken,
			&item.Customer.ID,
			&item.Customer.Name,
			&item.Total,
			&item.PaymentStatus,
			&item.WorkflowStatus,
			&item.CreatedAt,
			&item.ItemCount,
			&stID,
			&stName,
		); err != nil {
			return model.Paged[model.OrderListItem]{}, err
		}
		if stID != nil && stName != nil {
			item.FirstItem = &struct {
				ServiceType struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"serviceType"`
			}{}
			item.FirstItem.ServiceType.ID = *stID
			item.FirstItem.ServiceType.Name = *stName
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return model.Paged[model.OrderListItem]{}, err
	}

	return model.Paged[model.OrderListItem]{Items: out, Page: page, PageSize: pageSize, Total: total}, nil
}

func (r *OrderRepo) GetDetail(ctx context.Context, id string) (*model.OrderDetail, error) {
	var o model.OrderDetail
	err := r.db.Pool.QueryRow(ctx, `
		SELECT
			o.id,
			o.invoice_number,
			o.public_token,
			c.id,
			c.name,
			c.phone,
			o.total::text,
			o.payment_status::text,
			o.workflow_status::text,
			o.received_date,
			o.completed_date,
			o.pickup_date,
			o.image,
			o.note,
			o.created_at,
			o.updated_at
		FROM laundry_backend.orders o
		JOIN laundry_backend.customers c ON c.id = o.customer_id
		WHERE o.id=$1
	`, id).Scan(
		&o.ID,
		&o.InvoiceNumber,
		&o.PublicToken,
		&o.Customer.ID,
		&o.Customer.Name,
		&o.Customer.Phone,
		&o.Total,
		&o.PaymentStatus,
		&o.WorkflowStatus,
		&o.ReceivedDate,
		&o.CompletedDate,
		&o.PickupDate,
		&o.Image,
		&o.Note,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	itemsRows, err := r.db.Pool.Query(ctx, `
		SELECT
			oi.id,
			oi.service_type_id,
			st.name,
			st.unit,
			oi.quantity::text,
			oi.unit_price::text,
			oi.discount::text,
			oi.total::text,
			oi.created_at,
			oi.updated_at
		FROM laundry_backend.order_items oi
		JOIN laundry_backend.service_types st ON st.id = oi.service_type_id
		WHERE oi.order_id=$1
		ORDER BY oi.created_at ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer itemsRows.Close()

	items := []model.OrderItem{}
	itemIndex := map[string]int{}
	for itemsRows.Next() {
		var it model.OrderItem
		if err := itemsRows.Scan(
			&it.ID,
			&it.ServiceType.ID,
			&it.ServiceType.Name,
			&it.ServiceType.Unit,
			&it.Quantity,
			&it.UnitPrice,
			&it.Discount,
			&it.Total,
			&it.CreatedAt,
			&it.UpdatedAt,
		); err != nil {
			return nil, err
		}
		it.WorkAssignments = []model.WorkAssignment{}
		itemIndex[it.ID] = len(items)
		items = append(items, it)
	}
	if err := itemsRows.Err(); err != nil {
		return nil, err
	}

	waRows, err := r.db.Pool.Query(ctx, `
		SELECT
			wa.id,
			wa.order_item_id,
			wa.task_type::text,
			e.id,
			e.name,
			wa.percent::text,
			wa.amount::text,
			wa.created_at
		FROM laundry_backend.work_assignments wa
		JOIN laundry_backend.employees e ON e.id = wa.employee_id
		WHERE wa.order_id=$1
		ORDER BY wa.created_at ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer waRows.Close()

	for waRows.Next() {
		var wa model.WorkAssignment
		if err := waRows.Scan(
			&wa.ID,
			&wa.OrderItemID,
			&wa.TaskType,
			&wa.Employee.ID,
			&wa.Employee.Name,
			&wa.Percent,
			&wa.Amount,
			&wa.CreatedAt,
		); err != nil {
			return nil, err
		}
		if idx, ok := itemIndex[wa.OrderItemID]; ok {
			items[idx].WorkAssignments = append(items[idx].WorkAssignments, wa)
		}
	}
	if err := waRows.Err(); err != nil {
		return nil, err
	}

	payRows, err := r.db.Pool.Query(ctx, `
		SELECT id, order_id, amount::text, method, paid_at, note, created_at
		FROM laundry_backend.payments
		WHERE order_id=$1
		ORDER BY paid_at DESC
	`, id)
	if err != nil {
		return nil, err
	}
	defer payRows.Close()

	payments := []model.Payment{}
	for payRows.Next() {
		var p model.Payment
		if err := payRows.Scan(&p.ID, &p.OrderID, &p.Amount, &p.Method, &p.PaidAt, &p.Note, &p.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	if err := payRows.Err(); err != nil {
		return nil, err
	}

	attRows, err := r.db.Pool.Query(ctx, `
		SELECT id, order_id, file_path, mime_type, size_bytes, created_at
		FROM laundry_backend.order_attachments
		WHERE order_id=$1
		ORDER BY created_at DESC
	`, id)
	if err != nil {
		return nil, err
	}
	defer attRows.Close()

	atts := []model.OrderAttachment{}
	for attRows.Next() {
		var a model.OrderAttachment
		if err := attRows.Scan(&a.ID, &a.OrderID, &a.FilePath, &a.MimeType, &a.SizeBytes, &a.CreatedAt); err != nil {
			return nil, err
		}
		atts = append(atts, a)
	}
	if err := attRows.Err(); err != nil {
		return nil, err
	}

	o.Items = items
	o.Payments = payments
	o.Attachments = atts
	return &o, nil
}

func (r *OrderRepo) GetDetailByPublicToken(ctx context.Context, token string) (*model.OrderDetail, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, pgx.ErrNoRows
	}
	var id string
	if err := r.db.Pool.QueryRow(ctx, `
		SELECT id
		FROM laundry_backend.orders
		WHERE public_token=$1
	`, token).Scan(&id); err != nil {
		return nil, err
	}
	return r.GetDetail(ctx, id)
}

func (r *OrderRepo) Create(ctx context.Context, p repository.CreateOrderParams) (*model.OrderDetail, error) {
	orderID := cuid.New()
	now := time.Now()

	prefix := fmt.Sprintf("LDR-%04d%02d%02d-", now.Year(), int(now.Month()), now.Day())

	var invoice string
	for attempt := 0; attempt < 5; attempt++ {
		var count int
		err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM laundry_backend.orders WHERE invoice_number LIKE $1 || '%'`, prefix).Scan(&count)
		if err != nil {
			return nil, err
		}
		invoice = fmt.Sprintf("%s%03d", prefix, count+1+attempt)

		tx, err := r.db.Pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return nil, err
		}

		err = r.createOrderTx(ctx, tx, orderID, invoice, p)
		if err != nil {
			_ = tx.Rollback(ctx)
			if isUniqueViolation(err) {
				continue
			}
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		break
	}

	return r.GetDetail(ctx, orderID)
}

func (r *OrderRepo) createOrderTx(ctx context.Context, tx pgx.Tx, orderID, invoice string, p repository.CreateOrderParams) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO laundry_backend.orders (
			id,
			invoice_number,
			public_token,
			customer_id,
			total,
			payment_status,
			workflow_status,
			received_date,
			completed_date,
			pickup_date,
			image,
			note,
			created_at,
			updated_at
		) VALUES ($1,$2,$3,$4,$5,'unpaid','received',$6,$7,NULL,$8,$9,now(),now())
	`, orderID, invoice, cuid.New(), p.CustomerID, sumItemTotals(p.Items), p.ReceivedDate, p.CompletedDate, p.Image, p.Note)
	if err != nil {
		return err
	}

	for _, it := range p.Items {
		itemID := cuid.New()
		_, err := tx.Exec(ctx, `
			INSERT INTO laundry_backend.order_items (
				id,
				order_id,
				service_type_id,
				quantity,
				unit_price,
				discount,
				total,
				created_at,
				updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,now(),now())
		`, itemID, orderID, it.ServiceTypeID, it.Quantity, it.UnitPrice, it.Discount, it.Total)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *OrderRepo) UpdateImage(ctx context.Context, orderID string, image *string) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE laundry_backend.orders
		SET image=$2, updated_at=now()
		WHERE id=$1
	`, orderID, image)
	return err
}

func (r *OrderRepo) UpdateWorkflow(ctx context.Context, orderID string, workflowStatus string) error {
	switch workflowStatus {
	case "picked_up":
		_, err := r.db.Pool.Exec(ctx, `
			UPDATE laundry_backend.orders
			SET workflow_status='picked_up', pickup_date=now(), updated_at=now()
			WHERE id=$1
		`, orderID)
		return err
	case "finished":
		_, err := r.db.Pool.Exec(ctx, `
			UPDATE laundry_backend.orders
			SET workflow_status='finished', pickup_date=NULL, completed_date=now(), updated_at=now()
			WHERE id=$1
		`, orderID)
		return err
	case "received", "washing", "drying", "ironing":
		_, err := r.db.Pool.Exec(ctx, `
			UPDATE laundry_backend.orders
			SET workflow_status=$2, pickup_date=NULL, completed_date=NULL, updated_at=now()
			WHERE id=$1
		`, orderID, workflowStatus)
		return err
	default:
		return errors.New("invalid workflow status")
	}
}

func (r *OrderRepo) CreatePayment(ctx context.Context, orderID string, p repository.CreatePaymentParams) (*model.Payment, error) {
	tx, err := r.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	paymentID := cuid.New()
	var out model.Payment

	err = tx.QueryRow(ctx, `
		INSERT INTO laundry_backend.payments (id, order_id, amount, method, paid_at, note, created_at)
		VALUES ($1, $2, $3, $4, now(), $5, now())
		RETURNING id, order_id, amount::text, method, paid_at, note, created_at
	`, paymentID, orderID, p.Amount, p.Method, p.Note).Scan(
		&out.ID, &out.OrderID, &out.Amount, &out.Method, &out.PaidAt, &out.Note, &out.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		WITH sums AS (
			SELECT COALESCE(SUM(amount), 0) AS paid
			FROM laundry_backend.payments
			WHERE order_id = $1
		)
		UPDATE laundry_backend.orders
		SET payment_status = CASE
			WHEN (SELECT paid FROM sums) >= total THEN 'paid'
			WHEN (SELECT paid FROM sums) > 0 THEN 'partial'
			ELSE 'unpaid'
		END,
		updated_at=now()
		WHERE id = $1
	`, orderID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *OrderRepo) UpsertWorkAssignment(ctx context.Context, p repository.UpsertWorkAssignmentParams) error {
	id := cuid.New()
	ct, err := r.db.Pool.Exec(ctx, `
		INSERT INTO laundry_backend.work_assignments (id, order_id, order_item_id, employee_id, task_type, percent, amount, created_at)
		SELECT
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			ROUND(oi.total * ($6::numeric / 100), 2),
			now()
		FROM laundry_backend.order_items oi
		WHERE oi.id = $3 AND oi.order_id = $2
		ON CONFLICT (order_item_id, task_type)
		DO UPDATE SET
			employee_id = EXCLUDED.employee_id,
			percent = EXCLUDED.percent,
			amount = ROUND((SELECT oi.total FROM laundry_backend.order_items oi WHERE oi.id = EXCLUDED.order_item_id) * (EXCLUDED.percent / 100), 2)
	`, id, p.OrderID, p.OrderItemID, p.EmployeeID, p.TaskType, p.Percent)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *OrderRepo) DeleteWorkAssignment(ctx context.Context, orderItemID string, taskType string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM laundry_backend.work_assignments WHERE order_item_id=$1 AND task_type=$2`, orderItemID, taskType)
	return err
}

func (r *OrderRepo) CreateAttachments(ctx context.Context, orderID string, files []repository.CreateAttachmentParams) error {
	if len(files) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, f := range files {
		id := cuid.New()
		batch.Queue(`
			INSERT INTO laundry_backend.order_attachments (id, order_id, file_path, mime_type, size_bytes, created_at)
			VALUES ($1,$2,$3,$4,$5,now())
		`, id, orderID, f.FilePath, f.MimeType, f.SizeBytes)
	}
	br := r.db.Pool.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < len(files); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func sumItemTotals(items []repository.CreateOrderItemParams) string {
	var cents int64
	for _, it := range items {
		cents += parseMoneyCents(it.Total)
	}
	return formatMoneyCents(cents)
}

func parseMoneyCents(s string) int64 {
	var whole int64
	var frac int64
	var neg bool
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}
	parts := strings.SplitN(s, ".", 2)
	whole = parseInt64(parts[0])
	if len(parts) == 2 {
		fs := parts[1]
		if len(fs) >= 2 {
			frac = parseInt64(fs[:2])
		} else if len(fs) == 1 {
			frac = parseInt64(fs) * 10
		}
	}
	c := whole*100 + frac
	if neg {
		return -c
	}
	return c
}

func formatMoneyCents(c int64) string {
	neg := c < 0
	if neg {
		c = -c
	}
	whole := c / 100
	frac := c % 100
	if neg {
		return fmt.Sprintf("-%d.%02d", whole, frac)
	}
	return fmt.Sprintf("%d.%02d", whole, frac)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

var _ repository.OrderRepository = (*OrderRepo)(nil)

package pg

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"time"

	"laundry-backend/internal/repository"
)

var (
	// WITA (Asia/Makassar, UTC+8)
	witaLocation, _ = time.LoadLocation("Asia/Makassar")
)

type ReportRepo struct {
	db *DB
}

func NewReportRepo(db *DB) *ReportRepo {
	return &ReportRepo{db: db}
}

func (r *ReportRepo) OrdersCSV(ctx context.Context, start, end *time.Time) ([]byte, string, error) {
	where := "true"
	args := []any{}
	if start != nil || end != nil {
		where = "o.created_at >= COALESCE($1, o.created_at) AND o.created_at <= COALESCE($2, o.created_at)"
		var s, e any
		if start != nil {
			s = start.UTC()
		}
		if end != nil {
			e = end.UTC()
		}
		args = append(args, s, e)
	}

	rows, err := r.db.Pool.Query(ctx, fmt.Sprintf(`
		SELECT
			o.id,
			o.invoice_number,
			o.created_at,
			c.name,
			o.total::text,
			o.payment_status::text,
			o.workflow_status::text,
			oi.quantity::text,
			st.unit,
			st.name,
			oi.total::text
		FROM laundry_backend.orders o
		JOIN laundry_backend.customers c ON c.id = o.customer_id
		LEFT JOIN laundry_backend.order_items oi ON oi.order_id = o.id
		LEFT JOIN laundry_backend.service_types st ON st.id = oi.service_type_id
		WHERE %s
		ORDER BY o.created_at DESC, oi.created_at ASC
	`, where), args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	type item struct {
		qty   string
		unit  string
		name  string
		total string
	}

	type orderRow struct {
		invoice        string
		createdAt      time.Time
		customerName   string
		total          string
		paymentStatus  string
		workflowStatus string
		items          []item
	}

	byID := map[string]*orderRow{}
	orderIDs := []string{}

	for rows.Next() {
		var orderID string
		var invoice string
		var createdAt time.Time
		var customerName string
		var total string
		var paymentStatus string
		var workflowStatus string
		var qty *string
		var unit *string
		var serviceName *string
		var itemTotal *string

		if err := rows.Scan(
			&orderID,
			&invoice,
			&createdAt,
			&customerName,
			&total,
			&paymentStatus,
			&workflowStatus,
			&qty,
			&unit,
			&serviceName,
			&itemTotal,
		); err != nil {
			return nil, "", err
		}

		o := byID[orderID]
		if o == nil {
			o = &orderRow{
				invoice:        invoice,
				createdAt:      createdAt,
				customerName:   customerName,
				total:          total,
				paymentStatus:  paymentStatus,
				workflowStatus: workflowStatus,
				items:          []item{},
			}
			byID[orderID] = o
			orderIDs = append(orderIDs, orderID)
		}

		if qty != nil && unit != nil && serviceName != nil && itemTotal != nil {
			o.items = append(o.items, item{qty: *qty, unit: *unit, name: *serviceName, total: *itemTotal})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	sort.SliceStable(orderIDs, func(i, j int) bool {
		return byID[orderIDs[i]].createdAt.After(byID[orderIDs[j]].createdAt)
	})

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"invoice_number", "tanggal", "pelanggan", "items", "total", "payment_status", "workflow_status"})

	for _, id := range orderIDs {
		o := byID[id]
		itemsText := ""
		for idx, it := range o.items {
			if idx > 0 {
				itemsText += "; "
			}
			itemsText += fmt.Sprintf("%s %s %s (%s)", it.qty, it.unit, it.name, it.total)
		}
		_ = w.Write([]string{
			o.invoice,
			o.createdAt.In(witaLocation).Format(time.RFC3339),
			o.customerName,
			itemsText,
			o.total,
			o.paymentStatus,
			o.workflowStatus,
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, "", err
	}

	startPart := "all"
	endPart := "all"
	if start != nil {
		startPart = start.Format("2006-01-02")
	}
	if end != nil {
		endPart = end.Format("2006-01-02")
	}
	filename := fmt.Sprintf("report-%s-%s.csv", startPart, endPart)
	return buf.Bytes(), filename, nil
}

var _ repository.ReportRepository = (*ReportRepo)(nil)

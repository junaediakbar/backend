package pg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type DashboardRepo struct {
	db *DB
}

func NewDashboardRepo(db *DB) *DashboardRepo {
	return &DashboardRepo{db: db}
}

func (r *DashboardRepo) Summary(ctx context.Context, start, end *time.Time) (*model.DashboardSummary, error) {
	var s model.DashboardSummary
	var revenue *string

	createdAtWhere := "true"
	paymentWhere := "true"
	args := []any{}

	if start != nil || end != nil {
		createdAtConds := []string{"true"}
		paymentConds := []string{"true"}
		if start != nil {
			args = append(args, pgtype.Timestamp{Time: *start, Valid: true})
			createdAtConds = append(createdAtConds, fmt.Sprintf("created_at >= $%d", len(args)))
			paymentConds = append(paymentConds, fmt.Sprintf("paid_at >= $%d", len(args)))
		}
		if end != nil {
			args = append(args, pgtype.Timestamp{Time: *end, Valid: true})
			createdAtConds = append(createdAtConds, fmt.Sprintf("created_at <= $%d", len(args)))
			paymentConds = append(paymentConds, fmt.Sprintf("paid_at <= $%d", len(args)))
		}
		createdAtWhere = strings.Join(createdAtConds, " AND ")
		paymentWhere = strings.Join(paymentConds, " AND ")
	}

	err := r.db.Pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT
			(SELECT COUNT(*) FROM laundry_backend.customers WHERE %s) AS customer_count,
			(SELECT COUNT(*) FROM laundry_backend.orders WHERE %s) AS order_count,
			(SELECT COUNT(*) FROM laundry_backend.orders WHERE payment_status::text <> 'paid' AND %s) AS unpaid_count,
			(SELECT COALESCE(SUM(amount), 0)::text FROM laundry_backend.payments WHERE %s) AS total_revenue
	`, createdAtWhere, createdAtWhere, createdAtWhere, paymentWhere), args...).Scan(&s.CustomerCount, &s.OrderCount, &s.UnpaidCount, &revenue)
	if err != nil {
		return nil, err
	}
	if revenue != nil {
		s.TotalRevenue = *revenue
	} else {
		s.TotalRevenue = "0.00"
	}
	return &s, nil
}

func (r *DashboardRepo) RevenueSeries(ctx context.Context, start, end time.Time) ([]model.DashboardDailyRow, error) {
	args := []any{
		pgtype.Timestamp{Time: start, Valid: true},
		pgtype.Timestamp{Time: end, Valid: true},
	}

	rows, err := r.db.Pool.Query(ctx, `
		WITH days AS (
			SELECT generate_series($1::timestamp, $2::timestamp, interval '1 day')::date AS day
		),
		o AS (
			SELECT created_at::date AS day, COUNT(*)::int AS order_count
			FROM laundry_backend.orders
			WHERE created_at >= $1 AND created_at <= $2
			GROUP BY 1
		),
		p AS (
			SELECT paid_at::date AS day, COALESCE(SUM(amount), 0)::text AS revenue
			FROM laundry_backend.payments
			WHERE paid_at >= $1 AND paid_at <= $2
			GROUP BY 1
		)
		SELECT
			to_char(d.day, 'YYYY-MM-DD') AS date,
			COALESCE(o.order_count, 0) AS order_count,
			COALESCE(p.revenue, '0.00') AS revenue
		FROM days d
		LEFT JOIN o ON o.day = d.day
		LEFT JOIN p ON p.day = d.day
		ORDER BY d.day ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.DashboardDailyRow{}
	for rows.Next() {
		var rrow model.DashboardDailyRow
		if err := rows.Scan(&rrow.Date, &rrow.OrderCount, &rrow.Revenue); err != nil {
			return nil, err
		}
		out = append(out, rrow)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

var _ repository.DashboardRepository = (*DashboardRepo)(nil)

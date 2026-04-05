package pg

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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
			args = append(args, timestampAsUTCWall(*start))
			createdAtConds = append(createdAtConds, fmt.Sprintf("created_at >= $%d", len(args)))
			paymentConds = append(paymentConds, fmt.Sprintf("paid_at >= $%d", len(args)))
		}
		if end != nil {
			args = append(args, timestampAsUTCWall(*end))
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

func formatMoneyString(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// RevenueSeries mengagregasi per hari kalender WITA dari created_at / paid_at (bukan string invoice).
// Prefix invoice dulu memakai time.Now() UTC sehingga tanggal LDR-* bisa tidak sama dengan hari operasional;
// nota baru memakai zona waktu bisnis di OrderRepo.Create.
func (r *DashboardRepo) RevenueSeries(ctx context.Context, start, end time.Time) ([]model.DashboardDailyRow, error) {
	loc := witaLocation
	if loc == nil {
		loc = time.UTC
	}
	tStart := timestampAsUTCWall(start)
	tEnd := timestampAsUTCWall(end)

	type orderAgg struct {
		count int
		sum   float64
	}
	byDayOrders := map[string]*orderAgg{}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT created_at, total::float8
		FROM laundry_backend.orders
		WHERE created_at >= $1 AND created_at <= $2
	`, tStart, tEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var createdAt time.Time
		var total float64
		if err := rows.Scan(&createdAt, &total); err != nil {
			return nil, err
		}
		day := createdAt.In(loc).Format("2006-01-02")
		a := byDayOrders[day]
		if a == nil {
			a = &orderAgg{}
			byDayOrders[day] = a
		}
		a.count++
		a.sum += total
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	byDayRev := map[string]float64{}
	prows, err := r.db.Pool.Query(ctx, `
		SELECT paid_at, amount::float8
		FROM laundry_backend.payments
		WHERE paid_at >= $1 AND paid_at <= $2
	`, tStart, tEnd)
	if err != nil {
		return nil, err
	}
	defer prows.Close()
	for prows.Next() {
		var paidAt time.Time
		var amount float64
		if err := prows.Scan(&paidAt, &amount); err != nil {
			return nil, err
		}
		day := paidAt.In(loc).Format("2006-01-02")
		byDayRev[day] += amount
	}
	if err := prows.Err(); err != nil {
		return nil, err
	}

	first := time.Date(start.In(loc).Year(), start.In(loc).Month(), start.In(loc).Day(), 0, 0, 0, 0, loc)
	last := time.Date(end.In(loc).Year(), end.In(loc).Month(), end.In(loc).Day(), 0, 0, 0, 0, loc)

	out := make([]model.DashboardDailyRow, 0)
	for d := first; !d.After(last); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		oa := byDayOrders[key]
		oc := 0
		nt := 0.0
		if oa != nil {
			oc = oa.count
			nt = oa.sum
		}
		rv := byDayRev[key]
		out = append(out, model.DashboardDailyRow{
			Date:       key,
			OrderCount: oc,
			NotaTotal:  formatMoneyString(nt),
			Revenue:    formatMoneyString(rv),
		})
	}
	return out, nil
}

var _ repository.DashboardRepository = (*DashboardRepo)(nil)

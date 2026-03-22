package pg

import (
	"context"

	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type DashboardRepo struct {
	db *DB
}

func NewDashboardRepo(db *DB) *DashboardRepo {
	return &DashboardRepo{db: db}
}

func (r *DashboardRepo) Summary(ctx context.Context) (*model.DashboardSummary, error) {
	var s model.DashboardSummary
	var revenue *string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM laundry_backend.customers) AS customer_count,
			(SELECT COUNT(*) FROM laundry_backend.orders) AS order_count,
			(SELECT COUNT(*) FROM laundry_backend.orders WHERE payment_status::text <> 'paid') AS unpaid_count,
			(SELECT COALESCE(SUM(amount), 0)::text FROM laundry_backend.payments) AS total_revenue
	`).Scan(&s.CustomerCount, &s.OrderCount, &s.UnpaidCount, &revenue)
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

var _ repository.DashboardRepository = (*DashboardRepo)(nil)

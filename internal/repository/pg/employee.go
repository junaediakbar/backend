package pg

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/lucsky/cuid"

	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type EmployeeRepo struct {
	db *DB
}

func NewEmployeeRepo(db *DB) *EmployeeRepo {
	return &EmployeeRepo{db: db}
}

func (r *EmployeeRepo) List(ctx context.Context, onlyActive *bool) ([]model.Employee, error) {
	where := "true"
	args := []any{}
	if onlyActive != nil {
		where = "is_active=$1"
		args = append(args, *onlyActive)
	}

	rows, err := r.db.Pool.Query(ctx, fmt.Sprintf(`
		SELECT id, name, is_active, created_at, updated_at
		FROM laundry_backend.employees
		WHERE %s
		ORDER BY name ASC
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.Employee{}
	for rows.Next() {
		var e model.Employee
		if err := rows.Scan(&e.ID, &e.Name, &e.IsActive, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *EmployeeRepo) Get(ctx context.Context, id string) (*model.Employee, error) {
	var e model.Employee
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, is_active, created_at, updated_at
		FROM laundry_backend.employees
		WHERE id=$1
	`, id).Scan(&e.ID, &e.Name, &e.IsActive, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepo) Create(ctx context.Context, p repository.CreateEmployeeParams) (*model.Employee, error) {
	id := cuid.New()
	var e model.Employee
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO laundry_backend.employees (id, name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, now(), now())
		RETURNING id, name, is_active, created_at, updated_at
	`, id, p.Name, p.IsActive).Scan(&e.ID, &e.Name, &e.IsActive, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepo) Update(ctx context.Context, id string, p repository.UpdateEmployeeParams) (*model.Employee, error) {
	var e model.Employee
	err := r.db.Pool.QueryRow(ctx, `
		UPDATE laundry_backend.employees
		SET name=$1, is_active=$2, updated_at=now()
		WHERE id=$3
		RETURNING id, name, is_active, created_at, updated_at
	`, p.Name, p.IsActive, id).Scan(&e.ID, &e.Name, &e.IsActive, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepo) Performance(ctx context.Context, start, end *time.Time) ([]model.EmployeePerformanceRow, error) {
	args := []any{}
	where := "true"
	if start != nil || end != nil {
		where = "o.created_at >= COALESCE($1, o.created_at) AND o.created_at <= COALESCE($2, o.created_at)"
		args = append(args, start, end)
	}

	rows, err := r.db.Pool.Query(ctx, fmt.Sprintf(`
		SELECT
			e.id,
			e.name,
			wa.task_type::text,
			SUM(wa.amount)::text
		FROM laundry_backend.work_assignments wa
		JOIN laundry_backend.employees e ON e.id = wa.employee_id
		JOIN laundry_backend.orders o ON o.id = wa.order_id
		WHERE %s
		GROUP BY e.id, e.name, wa.task_type
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type agg struct {
		name        string
		pickupCents int64
		workCents   int64
		totalCents  int64
	}

	pickupSet := map[string]bool{
		"pickup_fuel":      true,
		"pickup_driver":    true,
		"pickup_worker_1":  true,
		"pickup_worker_2":  true,
		"dropoff_fuel":     true,
		"dropoff_driver":   true,
		"dropoff_worker_1": true,
		"dropoff_worker_2": true,
	}
	workSet := map[string]bool{"dust_removal": true, "brushing": true, "rinse_sprayer": true, "spin_dry": true, "finishing_packing": true}

	byID := map[string]*agg{}
	for rows.Next() {
		var id, name, taskType, sumText string
		if err := rows.Scan(&id, &name, &taskType, &sumText); err != nil {
			return nil, err
		}
		a := byID[id]
		if a == nil {
			a = &agg{name: name}
			byID[id] = a
		}
		cents := parseCents(sumText)
		if pickupSet[taskType] {
			a.pickupCents += cents
		}
		if workSet[taskType] {
			a.workCents += cents
		}
		a.totalCents += cents
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]model.EmployeePerformanceRow, 0, len(byID))
	for id, a := range byID {
		out = append(out, model.EmployeePerformanceRow{
			EmployeeID:   id,
			EmployeeName: a.name,
			PickupAmount: formatCents(a.pickupCents),
			WorkAmount:   formatCents(a.workCents),
			TotalAmount:  formatCents(a.totalCents),
		})
	}
	sort.Slice(out, func(i, j int) bool { return parseCents(out[i].TotalAmount) > parseCents(out[j].TotalAmount) })
	return out, nil
}

func parseCents(s string) int64 {
	var whole int64
	var frac int64
	var neg bool
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}
	parts := splitOnce(s, '.')
	whole = parseInt64(parts[0])
	if len(parts) > 1 {
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

func formatCents(c int64) string {
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

func splitOnce(s string, sep byte) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func parseInt64(s string) int64 {
	var n int64
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch < '0' || ch > '9' {
			break
		}
		n = n*10 + int64(ch-'0')
	}
	return n
}

var _ repository.EmployeeRepository = (*EmployeeRepo)(nil)

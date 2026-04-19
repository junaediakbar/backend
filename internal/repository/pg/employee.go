package pg

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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
		SELECT id, name, email, role::text, is_active, created_at, updated_at
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
		if err := rows.Scan(&e.ID, &e.Name, &e.Email, &e.Role, &e.IsActive, &e.CreatedAt, &e.UpdatedAt); err != nil {
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
		SELECT id, name, email, role::text, is_active, created_at, updated_at
		FROM laundry_backend.employees
		WHERE id=$1
	`, id).Scan(&e.ID, &e.Name, &e.Email, &e.Role, &e.IsActive, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepo) CountEmployeesWithRole(ctx context.Context, role string, excludeEmployeeID *string) (int, error) {
	role = strings.TrimSpace(role)
	if role == "" {
		return 0, nil
	}
	var n int
	var err error
	if excludeEmployeeID != nil && strings.TrimSpace(*excludeEmployeeID) != "" {
		err = r.db.Pool.QueryRow(ctx, `
			SELECT COUNT(*)::int
			FROM laundry_backend.employees
			WHERE role::text = $1 AND id <> $2
		`, role, strings.TrimSpace(*excludeEmployeeID)).Scan(&n)
	} else {
		err = r.db.Pool.QueryRow(ctx, `
			SELECT COUNT(*)::int
			FROM laundry_backend.employees
			WHERE role::text = $1
		`, role).Scan(&n)
	}
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (r *EmployeeRepo) GetByEmailForAuth(ctx context.Context, email string) (*repository.EmployeeAuthRow, error) {
	var e model.Employee
	var passwordHash string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, email, role::text, is_active, created_at, updated_at, password_hash
		FROM laundry_backend.employees
		WHERE lower(email)=lower($1)
	`, email).Scan(&e.ID, &e.Name, &e.Email, &e.Role, &e.IsActive, &e.CreatedAt, &e.UpdatedAt, &passwordHash)
	if err != nil {
		return nil, err
	}
	return &repository.EmployeeAuthRow{Employee: e, PasswordHash: passwordHash}, nil
}

func (r *EmployeeRepo) Create(ctx context.Context, p repository.CreateEmployeeParams) (*model.Employee, error) {
	id := cuid.New()
	var e model.Employee
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO laundry_backend.employees (id, name, email, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now(), now())
		RETURNING id, name, email, role::text, is_active, created_at, updated_at
	`, id, p.Name, p.Email, p.PasswordHash, p.Role, p.IsActive).Scan(&e.ID, &e.Name, &e.Email, &e.Role, &e.IsActive, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepo) Update(ctx context.Context, id string, p repository.UpdateEmployeeParams) (*model.Employee, error) {
	var e model.Employee
	if p.PasswordHash != nil {
		err := r.db.Pool.QueryRow(ctx, `
			UPDATE laundry_backend.employees
			SET name=$1, email=$2, role=$3, is_active=$4, password_hash=$5, updated_at=now()
			WHERE id=$6
			RETURNING id, name, email, role::text, is_active, created_at, updated_at
		`, p.Name, p.Email, p.Role, p.IsActive, *p.PasswordHash, id).Scan(&e.ID, &e.Name, &e.Email, &e.Role, &e.IsActive, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		return &e, nil
	}
	err := r.db.Pool.QueryRow(ctx, `
		UPDATE laundry_backend.employees
		SET name=$1, email=$2, role=$3, is_active=$4, updated_at=now()
		WHERE id=$5
		RETURNING id, name, email, role::text, is_active, created_at, updated_at
	`, p.Name, p.Email, p.Role, p.IsActive, id).Scan(&e.ID, &e.Name, &e.Email, &e.Role, &e.IsActive, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.db.Pool.Exec(ctx, `DELETE FROM laundry_backend.employees WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *EmployeeRepo) Performance(ctx context.Context, start, end *time.Time, onlyEmployeeID *string) ([]model.EmployeePerformanceRow, error) {
	args := []any{}
	conds := []string{"true"}
	if start != nil {
		args = append(args, timestampAsUTCWall(*start))
		conds = append(conds, fmt.Sprintf("o.created_at >= $%d", len(args)))
	}
	if end != nil {
		args = append(args, timestampAsUTCWall(*end))
		conds = append(conds, fmt.Sprintf("o.created_at <= $%d", len(args)))
	}
	if onlyEmployeeID != nil && strings.TrimSpace(*onlyEmployeeID) != "" {
		args = append(args, strings.TrimSpace(*onlyEmployeeID))
		conds = append(conds, fmt.Sprintf("e.id = $%d", len(args)))
	}
	where := strings.Join(conds, " AND ")

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
		a.totalCents += cents
		if strings.HasPrefix(taskType, "pickup_") || strings.HasPrefix(taskType, "dropoff_") {
			a.pickupCents += cents
		} else {
			a.workCents += cents
		}
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

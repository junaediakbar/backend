package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/lucsky/cuid"

	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type CustomerRepo struct {
	db *DB
}

func NewCustomerRepo(db *DB) *CustomerRepo {
	return &CustomerRepo{db: db}
}

func (r *CustomerRepo) List(ctx context.Context, q string, page, pageSize int) (model.Paged[model.Customer], error) {
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
		where = `(name ILIKE '%' || $1 || '%' OR phone ILIKE '%' || $1 || '%')`
		args = append(args, q)
	}

	var total int
	if err := r.db.Pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM laundry_backend.customers WHERE %s`, where), args...).Scan(&total); err != nil {
		return model.Paged[model.Customer]{}, err
	}

	argsList := append([]any{}, args...)
	argsList = append(argsList, pageSize, offset)

	rows, err := r.db.Pool.Query(ctx, fmt.Sprintf(`
		SELECT
			id,
			name,
			phone,
			address,
			latitude::float8,
			longitude::float8,
			email,
			notes,
			created_at,
			updated_at
		FROM laundry_backend.customers
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2), argsList...)
	if err != nil {
		return model.Paged[model.Customer]{}, err
	}
	defer rows.Close()

	items := make([]model.Customer, 0, pageSize)
	for rows.Next() {
		var c model.Customer
		var lat *float64
		var lng *float64
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Phone,
			&c.Address,
			&lat,
			&lng,
			&c.Email,
			&c.Notes,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return model.Paged[model.Customer]{}, err
		}
		c.Latitude = lat
		c.Longitude = lng
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return model.Paged[model.Customer]{}, err
	}

	return model.Paged[model.Customer]{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func (r *CustomerRepo) Get(ctx context.Context, id string) (*model.Customer, error) {
	var c model.Customer
	var lat *float64
	var lng *float64
	err := r.db.Pool.QueryRow(ctx, `
		SELECT
			id,
			name,
			phone,
			address,
			latitude::float8,
			longitude::float8,
			email,
			notes,
			created_at,
			updated_at
		FROM laundry_backend.customers
		WHERE id=$1
	`, id).Scan(
		&c.ID,
		&c.Name,
		&c.Phone,
		&c.Address,
		&lat,
		&lng,
		&c.Email,
		&c.Notes,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	c.Latitude = lat
	c.Longitude = lng
	return &c, nil
}

func (r *CustomerRepo) RecentOrders(ctx context.Context, customerID string, limit int) ([]model.CustomerOrderSummary, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, invoice_number, total::text, workflow_status::text
		FROM laundry_backend.orders
		WHERE customer_id=$1
		ORDER BY created_at DESC
		LIMIT $2
	`, customerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.CustomerOrderSummary{}
	for rows.Next() {
		var o model.CustomerOrderSummary
		if err := rows.Scan(&o.ID, &o.InvoiceNumber, &o.Total, &o.WorkflowStatus); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *CustomerRepo) Create(ctx context.Context, p repository.CreateCustomerParams) (*model.Customer, error) {
	id := cuid.New()
	var c model.Customer
	var lat *float64
	var lng *float64
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO laundry_backend.customers (id, name, phone, address, latitude, longitude, email, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())
		RETURNING id, name, phone, address, latitude::float8, longitude::float8, email, notes, created_at, updated_at
	`, id, p.Name, p.Phone, p.Address, p.Latitude, p.Longitude, p.Email, p.Notes).Scan(
		&c.ID, &c.Name, &c.Phone, &c.Address, &lat, &lng, &c.Email, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	c.Latitude = lat
	c.Longitude = lng
	return &c, nil
}

func (r *CustomerRepo) Update(ctx context.Context, id string, p repository.UpdateCustomerParams) (*model.Customer, error) {
	var c model.Customer
	var lat *float64
	var lng *float64
	err := r.db.Pool.QueryRow(ctx, `
		UPDATE laundry_backend.customers
		SET
			name=$1,
			phone=$2,
			address=$3,
			latitude=$4,
			longitude=$5,
			email=$6,
			notes=$7,
			updated_at=now()
		WHERE id=$8
		RETURNING id, name, phone, address, latitude::float8, longitude::float8, email, notes, created_at, updated_at
	`, p.Name, p.Phone, p.Address, p.Latitude, p.Longitude, p.Email, p.Notes, id).Scan(
		&c.ID, &c.Name, &c.Phone, &c.Address, &lat, &lng, &c.Email, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	c.Latitude = lat
	c.Longitude = lng
	return &c, nil
}

var _ repository.CustomerRepository = (*CustomerRepo)(nil)

package pg

import (
	"context"
	"fmt"

	"github.com/lucsky/cuid"

	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type ServiceTypeRepo struct {
	db *DB
}

func NewServiceTypeRepo(db *DB) *ServiceTypeRepo {
	return &ServiceTypeRepo{db: db}
}

func (r *ServiceTypeRepo) List(ctx context.Context, onlyActive *bool) ([]model.ServiceType, error) {
	where := "true"
	args := []any{}
	if onlyActive != nil {
		where = "is_active=$1"
		args = append(args, *onlyActive)
	}

	rows, err := r.db.Pool.Query(ctx, fmt.Sprintf(`
		SELECT id, name, unit, default_price::text, is_active, created_at, updated_at
		FROM laundry_backend.service_types
		WHERE %s
		ORDER BY name ASC
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.ServiceType{}
	for rows.Next() {
		var s model.ServiceType
		if err := rows.Scan(&s.ID, &s.Name, &s.Unit, &s.DefaultPrice, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ServiceTypeRepo) Get(ctx context.Context, id string) (*model.ServiceType, error) {
	var s model.ServiceType
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, unit, default_price::text, is_active, created_at, updated_at
		FROM laundry_backend.service_types
		WHERE id=$1
	`, id).Scan(&s.ID, &s.Name, &s.Unit, &s.DefaultPrice, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ServiceTypeRepo) Create(ctx context.Context, p repository.CreateServiceTypeParams) (*model.ServiceType, error) {
	id := cuid.New()
	var s model.ServiceType
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO laundry_backend.service_types (id, name, unit, default_price, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())
		RETURNING id, name, unit, default_price::text, is_active, created_at, updated_at
	`, id, p.Name, p.Unit, p.DefaultPrice, p.IsActive).Scan(&s.ID, &s.Name, &s.Unit, &s.DefaultPrice, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ServiceTypeRepo) Update(ctx context.Context, id string, p repository.UpdateServiceTypeParams) (*model.ServiceType, error) {
	var s model.ServiceType
	err := r.db.Pool.QueryRow(ctx, `
		UPDATE laundry_backend.service_types
		SET name=$1, unit=$2, default_price=$3, is_active=$4, updated_at=now()
		WHERE id=$5
		RETURNING id, name, unit, default_price::text, is_active, created_at, updated_at
	`, p.Name, p.Unit, p.DefaultPrice, p.IsActive, id).Scan(&s.ID, &s.Name, &s.Unit, &s.DefaultPrice, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

var _ repository.ServiceTypeRepository = (*ServiceTypeRepo)(nil)

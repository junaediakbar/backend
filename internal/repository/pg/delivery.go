package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/lucsky/cuid"

	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type DeliveryRepo struct {
	db *DB
}

func NewDeliveryRepo(db *DB) *DeliveryRepo {
	return &DeliveryRepo{db: db}
}

func (r *DeliveryRepo) ListPlans(ctx context.Context, limit int) ([]model.DeliveryPlanListItem, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT
			dp.id,
			dp.name,
			dp.planned_date,
			COALESCE(cnt.stop_count, 0) AS stop_count
		FROM laundry_backend.delivery_plans dp
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS stop_count
			FROM laundry_backend.delivery_stops ds
			WHERE ds.plan_id = dp.id
		) cnt ON true
		ORDER BY dp.planned_date DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.DeliveryPlanListItem{}
	for rows.Next() {
		var p model.DeliveryPlanListItem
		if err := rows.Scan(&p.ID, &p.Name, &p.PlannedDate, &p.StopCount); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *DeliveryRepo) GetPlan(ctx context.Context, id string) (*model.DeliveryPlanDetail, error) {
	var p model.DeliveryPlanDetail
	var startLat *float64
	var startLng *float64
	err := r.db.Pool.QueryRow(ctx, `
		SELECT
			id,
			name,
			planned_date,
			start_address,
			start_lat::float8,
			start_lng::float8,
			created_at,
			updated_at
		FROM laundry_backend.delivery_plans
		WHERE id=$1
	`, id).Scan(&p.ID, &p.Name, &p.PlannedDate, &p.StartAddress, &startLat, &startLng, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.StartLat = startLat
	p.StartLng = startLng

	rows, err := r.db.Pool.Query(ctx, `
		SELECT
			ds.id,
			ds.sequence,
			ds.distance_km::text,
			ds.created_at,
			c.id,
			c.name,
			c.phone,
			c.address,
			c.latitude::float8,
			c.longitude::float8,
			c.email,
			c.notes,
			c.created_at,
			c.updated_at
		FROM laundry_backend.delivery_stops ds
		JOIN laundry_backend.customers c ON c.id = ds.customer_id
		WHERE ds.plan_id=$1
		ORDER BY ds.sequence ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stops := []model.DeliveryStop{}
	for rows.Next() {
		var s model.DeliveryStop
		var lat *float64
		var lng *float64
		if err := rows.Scan(
			&s.ID,
			&s.Sequence,
			&s.DistanceKm,
			&s.CreatedAt,
			&s.Customer.ID,
			&s.Customer.Name,
			&s.Customer.Phone,
			&s.Customer.Address,
			&lat,
			&lng,
			&s.Customer.Email,
			&s.Customer.Notes,
			&s.Customer.CreatedAt,
			&s.Customer.UpdatedAt,
		); err != nil {
			return nil, err
		}
		s.Customer.Latitude = lat
		s.Customer.Longitude = lng
		stops = append(stops, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	p.Stops = stops
	return &p, nil
}

func (r *DeliveryRepo) CreatePlan(ctx context.Context, p repository.CreatePlanParams) (*model.DeliveryPlanDetail, error) {
	planID := cuid.New()

	tx, err := r.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		INSERT INTO laundry_backend.delivery_plans (
			id,
			name,
			planned_date,
			start_address,
			start_lat,
			start_lng,
			created_at,
			updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,now(),now())
	`, planID, p.Name, p.PlannedDate, p.StartAddress, p.StartLat, p.StartLng)
	if err != nil {
		return nil, err
	}

	for _, s := range p.Stops {
		stopID := cuid.New()
		_, err := tx.Exec(ctx, `
			INSERT INTO laundry_backend.delivery_stops (id, plan_id, customer_id, sequence, distance_km, created_at)
			VALUES ($1,$2,$3,$4,$5,now())
		`, stopID, planID, s.CustomerID, s.Sequence, s.DistanceKm)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetPlan(ctx, planID)
}

var _ repository.DeliveryRepository = (*DeliveryRepo)(nil)

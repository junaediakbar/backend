package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/lucsky/cuid"

	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type UserRepo struct {
	db *DB
}

func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) List(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, email, role::text, is_active, created_at, updated_at
		FROM laundry_backend.users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *UserRepo) Get(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, email, role::text, is_active, created_at, updated_at
		FROM laundry_backend.users
		WHERE id=$1
	`, id).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByEmailForAuth(ctx context.Context, email string) (*repository.UserAuthRow, error) {
	var u model.User
	var passwordHash string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, email, role::text, is_active, created_at, updated_at, password_hash
		FROM laundry_backend.users
		WHERE lower(email)=lower($1)
	`, email).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &passwordHash)
	if err != nil {
		return nil, err
	}
	return &repository.UserAuthRow{User: u, PasswordHash: passwordHash}, nil
}

func (r *UserRepo) Create(ctx context.Context, p repository.CreateUserParams) (*model.User, error) {
	id := cuid.New()
	authUserID := id
	var u model.User
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO laundry_backend.users (id, auth_user_id, name, email, role, password_hash, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,now(),now())
		RETURNING id, name, email, role::text, is_active, created_at, updated_at
	`, id, authUserID, p.Name, p.Email, p.Role, p.PasswordHash, p.IsActive).Scan(
		&u.ID, &u.Name, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Update(ctx context.Context, id string, p repository.UpdateUserParams) (*model.User, error) {
	var u model.User
	var passExpr string
	args := []any{p.Name, p.Email, p.Role, p.IsActive, id}
	if p.PasswordHash != nil {
		passExpr = ", password_hash=$6"
		args = append(args, *p.PasswordHash)
	}

	sql := `
		UPDATE laundry_backend.users
		SET
			name=$1,
			email=$2,
			role=$3,
			is_active=$4,
			updated_at=now()
	` + passExpr + `
		WHERE id=$5
		RETURNING id, name, email, role::text, is_active, created_at, updated_at
	`
	err := r.db.Pool.QueryRow(ctx, sql, args...).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.db.Pool.Exec(ctx, `DELETE FROM laundry_backend.users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

var _ repository.UserRepository = (*UserRepo)(nil)

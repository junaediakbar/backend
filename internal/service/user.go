package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) Get(ctx context.Context, id string) (*model.User, error) {
	out, err := s.repo.Get(ctx, strings.TrimSpace(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.NotFound("User tidak ditemukan")
		}
		return nil, err
	}
	return out, nil
}

type CreateUserInput struct {
	Name     string
	Email    string
	Role     string
	Password string
	IsActive bool
}

func (s *UserService) Create(ctx context.Context, in CreateUserInput) (*model.User, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(in.Email)
	in.Role = strings.TrimSpace(in.Role)

	if in.Name == "" {
		return nil, httpapi.BadRequest("validation_error", "Nama wajib diisi", nil)
	}
	if in.Email == "" {
		return nil, httpapi.BadRequest("validation_error", "Email wajib diisi", nil)
	}
	if in.Password == "" {
		return nil, httpapi.BadRequest("validation_error", "Password wajib diisi", nil)
	}
	if !isValidRole(in.Role) {
		return nil, httpapi.BadRequest("validation_error", "Role tidak valid", nil)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, repository.CreateUserParams{
		Name:         in.Name,
		Email:        in.Email,
		Role:         in.Role,
		PasswordHash: string(hash),
		IsActive:     in.IsActive,
	})
}

type UpdateUserInput struct {
	Name     string
	Email    string
	Role     string
	Password *string
	IsActive bool
}

func (s *UserService) Update(ctx context.Context, id string, in UpdateUserInput) (*model.User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, httpapi.BadRequest("validation_error", "ID wajib diisi", nil)
	}

	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(in.Email)
	in.Role = strings.TrimSpace(in.Role)

	if in.Name == "" {
		return nil, httpapi.BadRequest("validation_error", "Nama wajib diisi", nil)
	}
	if in.Email == "" {
		return nil, httpapi.BadRequest("validation_error", "Email wajib diisi", nil)
	}
	if !isValidRole(in.Role) {
		return nil, httpapi.BadRequest("validation_error", "Role tidak valid", nil)
	}

	var passHash *string
	if in.Password != nil && strings.TrimSpace(*in.Password) != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		s := string(hash)
		passHash = &s
	}

	out, err := s.repo.Update(ctx, id, repository.UpdateUserParams{
		Name:         in.Name,
		Email:        in.Email,
		Role:         in.Role,
		PasswordHash: passHash,
		IsActive:     in.IsActive,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.NotFound("User tidak ditemukan")
		}
		return nil, err
	}
	return out, nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return httpapi.BadRequest("validation_error", "ID wajib diisi", nil)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if err == pgx.ErrNoRows {
			return httpapi.NotFound("User tidak ditemukan")
		}
		return err
	}
	return nil
}

func isValidRole(role string) bool {
	switch role {
	case "owner", "admin", "cashier":
		return true
	default:
		return false
	}
}

package service

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type EmployeeService struct {
	repo repository.EmployeeRepository
}

func NewEmployeeService(repo repository.EmployeeRepository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) List(ctx context.Context, onlyActive *bool) ([]model.Employee, error) {
	return s.repo.List(ctx, onlyActive)
}

func (s *EmployeeService) Get(ctx context.Context, id string) (*model.Employee, error) {
	out, err := s.repo.Get(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.NotFound("Anggota tim tidak ditemukan")
		}
		return nil, err
	}
	return out, nil
}

type CreateEmployeeInput struct {
	Name     string
	Email    string
	Password string
	Role     string
	IsActive bool
}

func (s *EmployeeService) Create(ctx context.Context, in CreateEmployeeInput) (*model.Employee, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
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
	if !isValidStaffRole(in.Role) {
		return nil, httpapi.BadRequest("validation_error", "Role tidak valid", nil)
	}
	if err := s.ensureAtMostOneOwner(ctx, in.Role, nil); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, repository.CreateEmployeeParams{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: string(hash),
		Role:         in.Role,
		IsActive:     in.IsActive,
	})
}

type UpdateEmployeeInput struct {
	Name     string
	Email    string
	Role     string
	IsActive bool
	Password *string
}

func (s *EmployeeService) Update(ctx context.Context, id string, in UpdateEmployeeInput) (*model.Employee, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.Role = strings.TrimSpace(in.Role)
	if in.Name == "" {
		return nil, httpapi.BadRequest("validation_error", "Nama wajib diisi", nil)
	}
	if in.Email == "" {
		return nil, httpapi.BadRequest("validation_error", "Email wajib diisi", nil)
	}
	if !isValidStaffRole(in.Role) {
		return nil, httpapi.BadRequest("validation_error", "Role tidak valid", nil)
	}
	idTrim := strings.TrimSpace(id)
	if err := s.ensureAtMostOneOwner(ctx, in.Role, &idTrim); err != nil {
		return nil, err
	}
	var passHash *string
	if in.Password != nil && strings.TrimSpace(*in.Password) != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(*in.Password)), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		s := string(hash)
		passHash = &s
	}
	out, err := s.repo.Update(ctx, idTrim, repository.UpdateEmployeeParams{
		Name:         in.Name,
		Email:        in.Email,
		Role:         in.Role,
		IsActive:     in.IsActive,
		PasswordHash: passHash,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.NotFound("Anggota tim tidak ditemukan")
		}
		return nil, err
	}
	return out, nil
}

func (s *EmployeeService) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, strings.TrimSpace(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return httpapi.NotFound("Anggota tim tidak ditemukan")
		}
		return err
	}
	return nil
}

func (s *EmployeeService) Performance(ctx context.Context, start, end *time.Time, onlyEmployeeID *string) ([]model.EmployeePerformanceRow, error) {
	return s.repo.Performance(ctx, start, end, onlyEmployeeID)
}

func (s *EmployeeService) ensureAtMostOneOwner(ctx context.Context, desiredRole string, editingEmployeeID *string) error {
	if strings.ToLower(strings.TrimSpace(desiredRole)) != "owner" {
		return nil
	}
	n, err := s.repo.CountEmployeesWithRole(ctx, "owner", editingEmployeeID)
	if err != nil {
		return err
	}
	if n > 0 {
		return httpapi.BadRequest("validation_error", "Hanya boleh ada satu akun Owner.", nil)
	}
	return nil
}

func isValidStaffRole(role string) bool {
	switch role {
	case "owner", "admin", "cashier", "employee":
		return true
	default:
		return false
	}
}

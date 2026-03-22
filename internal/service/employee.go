package service

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

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
			return nil, httpapi.NotFound("Karyawan tidak ditemukan")
		}
		return nil, err
	}
	return out, nil
}

func (s *EmployeeService) Create(ctx context.Context, p repository.CreateEmployeeParams) (*model.Employee, error) {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return nil, httpapi.BadRequest("validation_error", "Nama wajib diisi", nil)
	}
	return s.repo.Create(ctx, p)
}

func (s *EmployeeService) Update(ctx context.Context, id string, p repository.UpdateEmployeeParams) (*model.Employee, error) {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return nil, httpapi.BadRequest("validation_error", "Nama wajib diisi", nil)
	}
	out, err := s.repo.Update(ctx, id, p)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.NotFound("Karyawan tidak ditemukan")
		}
		return nil, err
	}
	return out, nil
}

func (s *EmployeeService) Performance(ctx context.Context, start, end *time.Time) ([]model.EmployeePerformanceRow, error) {
	return s.repo.Performance(ctx, start, end)
}

package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type ServiceTypeService struct {
	repo repository.ServiceTypeRepository
}

func NewServiceTypeService(repo repository.ServiceTypeRepository) *ServiceTypeService {
	return &ServiceTypeService{repo: repo}
}

func (s *ServiceTypeService) List(ctx context.Context, onlyActive *bool) ([]model.ServiceType, error) {
	return s.repo.List(ctx, onlyActive)
}

func (s *ServiceTypeService) Get(ctx context.Context, id string) (*model.ServiceType, error) {
	out, err := s.repo.Get(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.NotFound("Tipe layanan tidak ditemukan")
		}
		return nil, err
	}
	return out, nil
}

func (s *ServiceTypeService) Create(ctx context.Context, p repository.CreateServiceTypeParams) (*model.ServiceType, error) {
	p.Name = strings.TrimSpace(p.Name)
	p.Unit = strings.TrimSpace(p.Unit)
	if p.Name == "" || p.Unit == "" {
		return nil, httpapi.BadRequest("validation_error", "Nama dan satuan wajib diisi", nil)
	}
	if strings.TrimSpace(p.DefaultPrice) == "" {
		return nil, httpapi.BadRequest("validation_error", "Harga wajib diisi", nil)
	}
	return s.repo.Create(ctx, p)
}

func (s *ServiceTypeService) Update(ctx context.Context, id string, p repository.UpdateServiceTypeParams) (*model.ServiceType, error) {
	p.Name = strings.TrimSpace(p.Name)
	p.Unit = strings.TrimSpace(p.Unit)
	if p.Name == "" || p.Unit == "" {
		return nil, httpapi.BadRequest("validation_error", "Nama dan satuan wajib diisi", nil)
	}
	if strings.TrimSpace(p.DefaultPrice) == "" {
		return nil, httpapi.BadRequest("validation_error", "Harga wajib diisi", nil)
	}
	out, err := s.repo.Update(ctx, id, p)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.NotFound("Tipe layanan tidak ditemukan")
		}
		return nil, err
	}
	return out, nil
}

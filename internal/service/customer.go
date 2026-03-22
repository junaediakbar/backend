package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type CustomerService struct {
	repo repository.CustomerRepository
}

func NewCustomerService(repo repository.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) List(ctx context.Context, q string, page, pageSize int) (model.Paged[model.Customer], error) {
	return s.repo.List(ctx, q, page, pageSize)
}

func (s *CustomerService) Get(ctx context.Context, id string) (*model.Customer, error) {
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.NotFound("Pelanggan tidak ditemukan")
		}
		return nil, err
	}
	return c, nil
}

func (s *CustomerService) RecentOrders(ctx context.Context, customerID string, limit int) ([]model.CustomerOrderSummary, error) {
	_, err := s.Get(ctx, customerID)
	if err != nil {
		return nil, err
	}
	return s.repo.RecentOrders(ctx, customerID, limit)
}

func (s *CustomerService) Create(ctx context.Context, p repository.CreateCustomerParams) (*model.Customer, error) {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return nil, httpapi.BadRequest("validation_error", "Nama wajib diisi", nil)
	}
	return s.repo.Create(ctx, p)
}

func (s *CustomerService) Update(ctx context.Context, id string, p repository.UpdateCustomerParams) (*model.Customer, error) {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return nil, httpapi.BadRequest("validation_error", "Nama wajib diisi", nil)
	}
	c, err := s.repo.Update(ctx, id, p)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.NotFound("Pelanggan tidak ditemukan")
		}
		return nil, err
	}
	return c, nil
}

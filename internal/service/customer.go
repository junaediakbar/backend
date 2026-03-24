package service

import (
	"context"
	"errors"
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
	if err := validateCustomerCoordinates(p.Latitude, p.Longitude); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, p)
}

func (s *CustomerService) Update(ctx context.Context, id string, p repository.UpdateCustomerParams) (*model.Customer, error) {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return nil, httpapi.BadRequest("validation_error", "Nama wajib diisi", nil)
	}
	if err := validateCustomerCoordinates(p.Latitude, p.Longitude); err != nil {
		return nil, err
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

func validateCustomerCoordinates(lat, lng *float64) error {
	if (lat == nil) != (lng == nil) {
		return httpapi.BadRequest("validation_error", "Latitude dan longitude harus diisi berpasangan", nil)
	}
	if lat != nil {
		if *lat < -90 || *lat > 90 {
			return httpapi.BadRequest("validation_error", "Latitude tidak valid", nil)
		}
	}
	if lng != nil {
		if *lng < -180 || *lng > 180 {
			return httpapi.BadRequest("validation_error", "Longitude tidak valid", nil)
		}
	}
	return nil
}

func (s *CustomerService) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return httpapi.BadRequest("validation_error", "ID tidak valid", nil)
	}
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return httpapi.NotFound("Pelanggan tidak ditemukan")
		}
		if errors.Is(err, repository.ErrCustomerHasOrders) {
			return httpapi.Conflict("Tidak dapat menghapus pelanggan yang masih memiliki nota.")
		}
		if errors.Is(err, repository.ErrCustomerHasDeliveryStops) {
			return httpapi.Conflict("Pelanggan masih terdaftar dalam rute pengiriman. Hapus stop dari rute terlebih dahulu.")
		}
		return err
	}
	return nil
}

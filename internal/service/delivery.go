package service

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
	"laundry-backend/internal/util"
)

type DeliveryService struct {
	repo repository.DeliveryRepository
}

func NewDeliveryService(repo repository.DeliveryRepository) *DeliveryService {
	return &DeliveryService{repo: repo}
}

func (s *DeliveryService) ListPlans(ctx context.Context, limit int) ([]model.DeliveryPlanListItem, error) {
	return s.repo.ListPlans(ctx, limit)
}

func (s *DeliveryService) GetPlan(ctx context.Context, id string) (*model.DeliveryPlanDetail, error) {
	out, err := s.repo.GetPlan(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.NotFound("Rencana tidak ditemukan")
		}
		return nil, err
	}
	return out, nil
}

type CreatePlanInput struct {
	Name         string
	PlannedDate  time.Time
	StartAddress *string
	StartLat     float64
	StartLng     float64
	EndAddress   *string
	EndLat       float64
	EndLng       float64
	Stops        []repository.CreateStopParams
}

func (s *DeliveryService) CreatePlan(ctx context.Context, in CreatePlanInput) (*model.DeliveryPlanDetail, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, httpapi.BadRequest("validation_error", "Nama wajib diisi", nil)
	}
	if in.PlannedDate.IsZero() {
		return nil, httpapi.BadRequest("validation_error", "Tanggal wajib diisi", nil)
	}
	if len(in.Stops) == 0 {
		return nil, httpapi.BadRequest("validation_error", "Minimal 1 stop", nil)
	}

	stops := make([]repository.CreateStopParams, 0, len(in.Stops))
	for _, s0 := range in.Stops {
		if strings.TrimSpace(s0.CustomerID) == "" || s0.Sequence <= 0 {
			continue
		}
		if strings.TrimSpace(s0.DistanceKm) == "" {
			s0.DistanceKm = util.Money2(0)
		}
		stops = append(stops, repository.CreateStopParams{
			CustomerID: s0.CustomerID,
			Sequence:   s0.Sequence,
			DistanceKm: s0.DistanceKm,
		})
	}
	if len(stops) == 0 {
		return nil, httpapi.BadRequest("validation_error", "Stop tidak valid", nil)
	}

	return s.repo.CreatePlan(ctx, repository.CreatePlanParams{
		Name:         in.Name,
		PlannedDate:  in.PlannedDate,
		StartAddress: in.StartAddress,
		StartLat:     in.StartLat,
		StartLng:     in.StartLng,
		EndAddress:   in.EndAddress,
		EndLat:       in.EndLat,
		EndLng:       in.EndLng,
		Stops:        stops,
	})
}

func (s *DeliveryService) DeletePlan(ctx context.Context, id string) error {
	err := s.repo.DeletePlan(ctx, strings.TrimSpace(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return httpapi.NotFound("Rencana tidak ditemukan")
		}
		return err
	}
	return nil
}

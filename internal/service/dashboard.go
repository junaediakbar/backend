package service

import (
	"context"
	"time"

	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type DashboardService struct {
	repo repository.DashboardRepository
}

func NewDashboardService(repo repository.DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) Summary(ctx context.Context, start, end *time.Time) (*model.DashboardSummary, error) {
	return s.repo.Summary(ctx, start, end)
}

func (s *DashboardService) RevenueSeries(ctx context.Context, start, end time.Time) ([]model.DashboardDailyRow, error) {
	return s.repo.RevenueSeries(ctx, start, end)
}

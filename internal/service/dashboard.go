package service

import (
	"context"

	"laundry-backend/internal/model"
	"laundry-backend/internal/repository"
)

type DashboardService struct {
	repo repository.DashboardRepository
}

func NewDashboardService(repo repository.DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) Summary(ctx context.Context) (*model.DashboardSummary, error) {
	return s.repo.Summary(ctx)
}

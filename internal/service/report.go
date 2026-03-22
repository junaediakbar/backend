package service

import (
	"context"
	"time"

	"laundry-backend/internal/repository"
)

type ReportService struct {
	repo repository.ReportRepository
}

func NewReportService(repo repository.ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) OrdersCSV(ctx context.Context, start, end *time.Time) ([]byte, string, error) {
	return s.repo.OrdersCSV(ctx, start, end)
}

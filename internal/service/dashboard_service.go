package service

import (
	"context"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
)

// DashboardService aggregates ASM metrics for an organization.
type DashboardService struct {
	repos *repository.Repositories
}

// Summary builds the full dashboard summary for an organization.
func (s *DashboardService) Summary(ctx context.Context, orgID string) (*models.DashboardSummary, error) {
	total, err := s.repos.Dashboard.TotalAssets(ctx, orgID)
	if err != nil {
		return nil, err
	}
	byType, err := s.repos.Dashboard.AssetsByType(ctx, orgID)
	if err != nil {
		return nil, err
	}
	recent, err := s.repos.Dashboard.RecentChanges(ctx, orgID, 10)
	if err != nil {
		return nil, err
	}
	projects, err := s.repos.Dashboard.ProjectStats(ctx, orgID)
	if err != nil {
		return nil, err
	}
	teams, err := s.repos.Dashboard.TeamStats(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if byType == nil {
		byType = []models.AssetTypeCount{}
	}
	if recent == nil {
		recent = []models.Asset{}
	}
	return &models.DashboardSummary{
		TotalAssets:   total,
		AssetsByType:  byType,
		RecentChanges: recent,
		ProjectStats:  projects,
		TeamStats:     teams,
	}, nil
}

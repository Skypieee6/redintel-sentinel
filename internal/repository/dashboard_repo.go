package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

// DashboardRepository runs aggregate queries for dashboards.
type DashboardRepository struct{ pool *pgxpool.Pool }

// TotalAssets returns the asset count for an organization.
func (r *DashboardRepository) TotalAssets(ctx context.Context, orgID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM assets WHERE org_id = $1`, orgID).Scan(&n)
	return n, mapError(err)
}

// AssetsByType returns per-type asset counts for an organization.
func (r *DashboardRepository) AssetsByType(ctx context.Context, orgID string) ([]models.AssetTypeCount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT type, count(*) FROM assets WHERE org_id = $1 GROUP BY type ORDER BY count(*) DESC`, orgID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []models.AssetTypeCount
	for rows.Next() {
		var c models.AssetTypeCount
		if err := rows.Scan(&c.Type, &c.Count); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// RecentChanges returns the most recently updated assets for an organization.
func (r *DashboardRepository) RecentChanges(ctx context.Context, orgID string, limit int) ([]models.Asset, error) {
	assetRepo := &AssetRepository{pool: r.pool}
	assets, _, err := assetRepo.List(ctx, AssetFilter{OrgID: orgID, Limit: limit})
	return assets, err
}

// ProjectStats returns project counts and per-project asset tallies.
func (r *DashboardRepository) ProjectStats(ctx context.Context, orgID string) (models.ProjectStatistics, error) {
	var s models.ProjectStatistics
	rows, err := r.pool.Query(ctx,
		`SELECT p.id, p.name, p.status, count(a.id)
		 FROM projects p LEFT JOIN assets a ON a.project_id = p.id
		 WHERE p.org_id = $1
		 GROUP BY p.id, p.name, p.status
		 ORDER BY count(a.id) DESC`, orgID)
	if err != nil {
		return s, mapError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var pc models.ProjectAssetCount
		if err := rows.Scan(&pc.ProjectID, &pc.Name, &pc.Status, &pc.Assets); err != nil {
			return s, err
		}
		s.Total++
		if pc.Status == "archived" {
			s.Archived++
		} else {
			s.Active++
		}
		s.ByProject = append(s.ByProject, pc)
	}
	return s, rows.Err()
}

// TeamStats returns team, member and pending-invitation counts.
func (r *DashboardRepository) TeamStats(ctx context.Context, orgID string) (models.TeamStatistics, error) {
	var s models.TeamStatistics
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM teams WHERE org_id = $1),
			(SELECT count(*) FROM memberships WHERE org_id = $1),
			(SELECT count(*) FROM invitations WHERE org_id = $1 AND status = 'pending')`,
		orgID).Scan(&s.Teams, &s.Members, &s.PendingInvites)
	return s, mapError(err)
}

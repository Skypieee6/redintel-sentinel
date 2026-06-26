package service

import (
	"context"
	"strings"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
)

// AssetPage is a paginated asset listing.
type AssetPage struct {
	Assets []models.Asset `json:"assets"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// AssetService manages attack-surface assets within a project.
type AssetService struct {
	repos *repository.Repositories
	audit *AuditService
}

// canManageProject reuses project access rules for asset writes.
func (s *AssetService) canManageProject(ctx context.Context, orgID, projectID, userID string, orgRole models.Role) (bool, error) {
	p, err := s.repos.Projects.Get(ctx, orgID, projectID)
	if err != nil {
		return false, mapRepoErr(err, "project")
	}
	if orgRole.AtLeast(models.RoleAnalyst) {
		return true, nil
	}
	if p.OwnerID == userID {
		return true, nil
	}
	members, err := s.repos.Projects.ListMembers(ctx, projectID)
	if err != nil {
		return false, err
	}
	for _, m := range members {
		if m.UserID == userID && m.Role.AtLeast(models.RoleAnalyst) {
			return true, nil
		}
	}
	return false, nil
}

// Create adds an asset after validation and access checks.
func (s *AssetService) Create(ctx context.Context, actorID, orgID, projectID string, orgRole models.Role, a *models.Asset, ip string) (*models.Asset, error) {
	ok, err := s.canManageProject(ctx, orgID, projectID, actorID, orgRole)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, wrap(ErrForbidden, "analyst role or above is required to add assets")
	}
	if !a.Type.Valid() {
		return nil, wrap(ErrValidation, "invalid asset type %q", a.Type)
	}
	if strings.TrimSpace(a.Value) == "" {
		return nil, wrap(ErrValidation, "asset value is required")
	}
	a.OrgID = orgID
	a.ProjectID = projectID
	created, err := s.repos.Assets.Create(ctx, a)
	if err != nil {
		if err == repository.ErrConflict {
			return nil, wrap(ErrConflict, "this asset already exists in the project")
		}
		return nil, err
	}
	s.audit.Record(ctx, "asset.created", actorID, orgID, "asset", created.ID, ip,
		map[string]any{"type": string(created.Type), "value": created.Value})
	return created, nil
}

// Get returns an asset.
func (s *AssetService) Get(ctx context.Context, projectID, id string) (*models.Asset, error) {
	a, err := s.repos.Assets.Get(ctx, projectID, id)
	if err != nil {
		return nil, mapRepoErr(err, "asset")
	}
	return a, nil
}

// List returns a paginated, filtered asset listing.
func (s *AssetService) List(ctx context.Context, f repository.AssetFilter) (*AssetPage, error) {
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	f.Limit = limit
	assets, total, err := s.repos.Assets.List(ctx, f)
	if err != nil {
		return nil, err
	}
	if assets == nil {
		assets = []models.Asset{}
	}
	return &AssetPage{Assets: assets, Total: total, Limit: limit, Offset: f.Offset}, nil
}

// Update modifies an asset after access checks.
func (s *AssetService) Update(ctx context.Context, actorID, orgID, projectID, id string, orgRole models.Role, value string, tags []string, attributes map[string]any, status, ip string) (*models.Asset, error) {
	ok, err := s.canManageProject(ctx, orgID, projectID, actorID, orgRole)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, wrap(ErrForbidden, "analyst role or above is required to modify assets")
	}
	existing, err := s.repos.Assets.Get(ctx, projectID, id)
	if err != nil {
		return nil, mapRepoErr(err, "asset")
	}
	if value == "" {
		value = existing.Value
	}
	if status == "" {
		status = existing.Status
	}
	if tags == nil {
		tags = existing.Tags
	}
	if attributes == nil {
		attributes = existing.Attributes
	}
	updated, err := s.repos.Assets.Update(ctx, projectID, id, value, tags, attributes, status)
	if err != nil {
		return nil, mapRepoErr(err, "asset")
	}
	s.audit.Record(ctx, "asset.updated", actorID, orgID, "asset", id, ip, nil)
	return updated, nil
}

// Delete removes an asset after access checks.
func (s *AssetService) Delete(ctx context.Context, actorID, orgID, projectID, id string, orgRole models.Role, ip string) error {
	ok, err := s.canManageProject(ctx, orgID, projectID, actorID, orgRole)
	if err != nil {
		return err
	}
	if !ok {
		return wrap(ErrForbidden, "analyst role or above is required to delete assets")
	}
	if err := s.repos.Assets.Delete(ctx, projectID, id); err != nil {
		return mapRepoErr(err, "asset")
	}
	s.audit.Record(ctx, "asset.deleted", actorID, orgID, "asset", id, ip, nil)
	return nil
}

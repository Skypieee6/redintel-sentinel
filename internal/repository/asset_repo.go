package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

// AssetRepository persists attack-surface assets.
type AssetRepository struct{ pool *pgxpool.Pool }

const assetColumns = "id, org_id, project_id, type, value, tags, attributes, status, first_seen, last_seen, created_at, updated_at"

func scanAsset(row interface{ Scan(...any) error }) (*models.Asset, error) {
	var a models.Asset
	if err := row.Scan(&a.ID, &a.OrgID, &a.ProjectID, &a.Type, &a.Value, &a.Tags,
		&a.Attributes, &a.Status, &a.FirstSeen, &a.LastSeen, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, mapError(err)
	}
	return &a, nil
}

// Create inserts an asset.
func (r *AssetRepository) Create(ctx context.Context, a *models.Asset) (*models.Asset, error) {
	attrs := a.Attributes
	if attrs == nil {
		attrs = map[string]any{}
	}
	tags := a.Tags
	if tags == nil {
		tags = []string{}
	}
	return scanAsset(r.pool.QueryRow(ctx,
		`INSERT INTO assets (org_id, project_id, type, value, tags, attributes)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+assetColumns,
		a.OrgID, a.ProjectID, string(a.Type), a.Value, tags, attrs))
}

// Get returns an asset scoped to its project.
func (r *AssetRepository) Get(ctx context.Context, projectID, id string) (*models.Asset, error) {
	return scanAsset(r.pool.QueryRow(ctx,
		`SELECT `+assetColumns+` FROM assets WHERE id = $1 AND project_id = $2`, id, projectID))
}

// Update modifies mutable asset fields and bumps last_seen.
func (r *AssetRepository) Update(ctx context.Context, projectID, id string, value string, tags []string, attributes map[string]any, status string) (*models.Asset, error) {
	if tags == nil {
		tags = []string{}
	}
	if attributes == nil {
		attributes = map[string]any{}
	}
	return scanAsset(r.pool.QueryRow(ctx,
		`UPDATE assets SET value = $3, tags = $4, attributes = $5, status = $6,
		   last_seen = now(), updated_at = now()
		 WHERE id = $1 AND project_id = $2 RETURNING `+assetColumns,
		id, projectID, value, tags, attributes, status))
}

// Delete removes an asset.
func (r *AssetRepository) Delete(ctx context.Context, projectID, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM assets WHERE id = $1 AND project_id = $2`, id, projectID)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// AssetFilter narrows an asset listing.
type AssetFilter struct {
	ProjectID string
	OrgID     string
	Type      string
	Query     string
	Tag       string
	Status    string
	Limit     int
	Offset    int
}

func (f AssetFilter) where() (string, []any) {
	conds := []string{}
	args := []any{}
	add := func(cond string, val any) {
		args = append(args, val)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}
	if f.ProjectID != "" {
		add("project_id = $%d", f.ProjectID)
	}
	if f.OrgID != "" {
		add("org_id = $%d", f.OrgID)
	}
	if f.Type != "" {
		add("type = $%d", f.Type)
	}
	if f.Status != "" {
		add("status = $%d", f.Status)
	}
	if f.Query != "" {
		add("value ILIKE '%%' || $%d || '%%'", f.Query)
	}
	if f.Tag != "" {
		add("$%d = ANY(tags)", f.Tag)
	}
	if len(conds) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}

// List returns matching assets plus the total count for pagination.
func (r *AssetRepository) List(ctx context.Context, f AssetFilter) ([]models.Asset, int, error) {
	where, args := f.where()

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM assets`+where, args...).Scan(&total); err != nil {
		return nil, 0, mapError(err)
	}

	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	listArgs := append(args, limit, f.Offset)
	q := fmt.Sprintf(`SELECT %s FROM assets%s ORDER BY updated_at DESC LIMIT $%d OFFSET $%d`,
		assetColumns, where, len(listArgs)-1, len(listArgs))

	rows, err := r.pool.Query(ctx, q, listArgs...)
	if err != nil {
		return nil, 0, mapError(err)
	}
	defer rows.Close()
	var out []models.Asset
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *a)
	}
	return out, total, rows.Err()
}

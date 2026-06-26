package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

// DiscoveryRepository persists passive discovery jobs and their results.
type DiscoveryRepository struct{ pool *pgxpool.Pool }

const discoveryJobColumns = `id, org_id, project_id, input_type, input_value, sources,
	status, error, assets_found, assets_created, COALESCE(created_by::text, ''),
	started_at, completed_at, created_at, updated_at`

func scanDiscoveryJob(row interface{ Scan(...any) error }) (*models.DiscoveryJob, error) {
	var j models.DiscoveryJob
	if err := row.Scan(&j.ID, &j.OrgID, &j.ProjectID, &j.InputType, &j.InputValue, &j.Sources,
		&j.Status, &j.Error, &j.AssetsFound, &j.AssetsCreated, &j.CreatedBy,
		&j.StartedAt, &j.CompletedAt, &j.CreatedAt, &j.UpdatedAt); err != nil {
		return nil, mapError(err)
	}
	if j.Sources == nil {
		j.Sources = []string{}
	}
	return &j, nil
}

// CreateJob inserts a new discovery job in the pending state.
func (r *DiscoveryRepository) CreateJob(ctx context.Context, j *models.DiscoveryJob) (*models.DiscoveryJob, error) {
	sources := j.Sources
	if sources == nil {
		sources = []string{}
	}
	return scanDiscoveryJob(r.pool.QueryRow(ctx,
		`INSERT INTO discovery_jobs (org_id, project_id, input_type, input_value, sources, created_by)
		 VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::uuid) RETURNING `+discoveryJobColumns,
		j.OrgID, j.ProjectID, string(j.InputType), j.InputValue, sources, j.CreatedBy))
}

// GetJob returns a discovery job scoped to its org and project.
func (r *DiscoveryRepository) GetJob(ctx context.Context, orgID, projectID, id string) (*models.DiscoveryJob, error) {
	return scanDiscoveryJob(r.pool.QueryRow(ctx,
		`SELECT `+discoveryJobColumns+` FROM discovery_jobs
		 WHERE id = $1 AND org_id = $2 AND project_id = $3`, id, orgID, projectID))
}

// MarkRunning transitions a job to the running state.
func (r *DiscoveryRepository) MarkRunning(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE discovery_jobs SET status = 'running', started_at = now(), updated_at = now()
		 WHERE id = $1`, id)
	return mapError(err)
}

// MarkCompleted records a successful run and its asset counts.
func (r *DiscoveryRepository) MarkCompleted(ctx context.Context, id string, found, created int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE discovery_jobs SET status = 'completed', assets_found = $2, assets_created = $3,
		   completed_at = now(), updated_at = now()
		 WHERE id = $1`, id, found, created)
	return mapError(err)
}

// MarkFailed records a failed run with its error message.
func (r *DiscoveryRepository) MarkFailed(ctx context.Context, id, reason string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE discovery_jobs SET status = 'failed', error = $2,
		   completed_at = now(), updated_at = now()
		 WHERE id = $1`, id, reason)
	return mapError(err)
}

// ListJobs returns paginated discovery jobs for a project (most recent first)
// plus the total count.
func (r *DiscoveryRepository) ListJobs(ctx context.Context, orgID, projectID string, limit, offset int) ([]models.DiscoveryJob, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT count(*) FROM discovery_jobs WHERE org_id = $1 AND project_id = $2`,
		orgID, projectID).Scan(&total); err != nil {
		return nil, 0, mapError(err)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+discoveryJobColumns+` FROM discovery_jobs
		 WHERE org_id = $1 AND project_id = $2
		 ORDER BY created_at DESC LIMIT $3 OFFSET $4`, orgID, projectID, limit, offset)
	if err != nil {
		return nil, 0, mapError(err)
	}
	defer rows.Close()
	var out []models.DiscoveryJob
	for rows.Next() {
		j, err := scanDiscoveryJob(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *j)
	}
	return out, total, rows.Err()
}

// AddResults batch-inserts a job's findings.
func (r *DiscoveryRepository) AddResults(ctx context.Context, results []*models.DiscoveryResult) error {
	for _, res := range results {
		attrs := res.Attributes
		if attrs == nil {
			attrs = map[string]any{}
		}
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO discovery_results (job_id, asset_id, type, value, source, attributes, is_new)
			 VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, $6, $7)`,
			res.JobID, res.AssetID, string(res.Type), res.Value, res.Source, attrs, res.IsNew); err != nil {
			return mapError(err)
		}
	}
	return nil
}

// ListResults returns the findings of a discovery job.
func (r *DiscoveryRepository) ListResults(ctx context.Context, jobID string) ([]models.DiscoveryResult, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, job_id, COALESCE(asset_id::text, ''), type, value, source, attributes, is_new, created_at
		 FROM discovery_results WHERE job_id = $1 ORDER BY type, value`, jobID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []models.DiscoveryResult
	for rows.Next() {
		var res models.DiscoveryResult
		if err := rows.Scan(&res.ID, &res.JobID, &res.AssetID, &res.Type, &res.Value,
			&res.Source, &res.Attributes, &res.IsNew, &res.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	return out, rows.Err()
}

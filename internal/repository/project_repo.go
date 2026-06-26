package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

// ProjectRepository persists projects and explicit project membership.
type ProjectRepository struct{ pool *pgxpool.Pool }

const projColumns = "id, org_id, name, description, COALESCE(owner_id::text, ''), status, created_at, updated_at"

func scanProject(row interface{ Scan(...any) error }) (*models.Project, error) {
	var p models.Project
	if err := row.Scan(&p.ID, &p.OrgID, &p.Name, &p.Description, &p.OwnerID, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, mapError(err)
	}
	return &p, nil
}

// Create inserts a project.
func (r *ProjectRepository) Create(ctx context.Context, orgID, name, description, ownerID string) (*models.Project, error) {
	return scanProject(r.pool.QueryRow(ctx,
		`INSERT INTO projects (org_id, name, description, owner_id) VALUES ($1, $2, $3, $4) RETURNING `+projColumns,
		orgID, name, description, ownerID))
}

// Get returns a project by id within an org.
func (r *ProjectRepository) Get(ctx context.Context, orgID, id string) (*models.Project, error) {
	return scanProject(r.pool.QueryRow(ctx,
		`SELECT `+projColumns+` FROM projects WHERE id = $1 AND org_id = $2`, id, orgID))
}

// List returns the projects of an org.
func (r *ProjectRepository) List(ctx context.Context, orgID string) ([]models.Project, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+projColumns+` FROM projects WHERE org_id = $1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []models.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// Update changes mutable project fields.
func (r *ProjectRepository) Update(ctx context.Context, orgID, id, name, description, status string) (*models.Project, error) {
	return scanProject(r.pool.QueryRow(ctx,
		`UPDATE projects SET name = $3, description = $4, status = $5, updated_at = now()
		 WHERE id = $1 AND org_id = $2 RETURNING `+projColumns, id, orgID, name, description, status))
}

// SetStatus updates a project's status (e.g. active/archived).
func (r *ProjectRepository) SetStatus(ctx context.Context, orgID, id, status string) (*models.Project, error) {
	return scanProject(r.pool.QueryRow(ctx,
		`UPDATE projects SET status = $3, updated_at = now()
		 WHERE id = $1 AND org_id = $2 RETURNING `+projColumns, id, orgID, status))
}

// Delete removes a project.
func (r *ProjectRepository) Delete(ctx context.Context, orgID, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1 AND org_id = $2`, id, orgID)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// AddMember grants a user explicit access to a project.
func (r *ProjectRepository) AddMember(ctx context.Context, projectID, userID string, role models.Role) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT (project_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		projectID, userID, string(role))
	return mapError(err)
}

// RemoveMember revokes a user's explicit project access.
func (r *ProjectRepository) RemoveMember(ctx context.Context, projectID, userID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`, projectID, userID)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListMembers returns explicit members of a project.
func (r *ProjectRepository) ListMembers(ctx context.Context, projectID string) ([]models.ProjectMember, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT pm.project_id, pm.user_id, u.email, pm.role, pm.created_at
		 FROM project_members pm JOIN users u ON u.id = pm.user_id
		 WHERE pm.project_id = $1 ORDER BY pm.created_at`, projectID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []models.ProjectMember
	for rows.Next() {
		var m models.ProjectMember
		if err := rows.Scan(&m.ProjectID, &m.UserID, &m.Email, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

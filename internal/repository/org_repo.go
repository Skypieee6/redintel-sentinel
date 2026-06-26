package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

// OrgRepository persists organizations.
type OrgRepository struct{ pool *pgxpool.Pool }

const orgColumns = "id, name, slug, COALESCE(created_by::text, ''), created_at, updated_at"

func scanOrg(row interface{ Scan(...any) error }) (*models.Organization, error) {
	var o models.Organization
	err := row.Scan(&o.ID, &o.Name, &o.Slug, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, mapError(err)
	}
	return &o, nil
}

// Create inserts an organization.
func (r *OrgRepository) Create(ctx context.Context, name, slug, createdBy string) (*models.Organization, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO organizations (name, slug, created_by) VALUES ($1, $2, $3) RETURNING `+orgColumns,
		name, slug, createdBy)
	return scanOrg(row)
}

// GetByID returns an organization by id.
func (r *OrgRepository) GetByID(ctx context.Context, id string) (*models.Organization, error) {
	return scanOrg(r.pool.QueryRow(ctx, `SELECT `+orgColumns+` FROM organizations WHERE id = $1`, id))
}

// Update changes an organization's name.
func (r *OrgRepository) Update(ctx context.Context, id, name string) (*models.Organization, error) {
	return scanOrg(r.pool.QueryRow(ctx,
		`UPDATE organizations SET name = $2, updated_at = now() WHERE id = $1 RETURNING `+orgColumns, id, name))
}

// Delete removes an organization.
func (r *OrgRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListForUser returns organizations the user is a member of.
func (r *OrgRepository) ListForUser(ctx context.Context, userID string) ([]models.Organization, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT o.id, o.name, o.slug, COALESCE(o.created_by::text, ''), o.created_at, o.updated_at
		 FROM organizations o
		 JOIN memberships m ON m.org_id = o.id
		 WHERE m.user_id = $1 ORDER BY o.created_at DESC`, userID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []models.Organization
	for rows.Next() {
		o, err := scanOrg(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *o)
	}
	return out, rows.Err()
}

// MembershipRepository persists organization memberships.
type MembershipRepository struct{ pool *pgxpool.Pool }

// Upsert creates or updates a membership role.
func (r *MembershipRepository) Upsert(ctx context.Context, orgID, userID string, role models.Role) (*models.Membership, error) {
	var m models.Membership
	err := r.pool.QueryRow(ctx,
		`INSERT INTO memberships (org_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT (org_id, user_id) DO UPDATE SET role = EXCLUDED.role, updated_at = now()
		 RETURNING id, org_id, user_id, role, created_at, updated_at`,
		orgID, userID, string(role)).
		Scan(&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, mapError(err)
	}
	return &m, nil
}

// Get returns the membership of a user in an org.
func (r *MembershipRepository) Get(ctx context.Context, orgID, userID string) (*models.Membership, error) {
	var m models.Membership
	err := r.pool.QueryRow(ctx,
		`SELECT id, org_id, user_id, role, created_at, updated_at
		 FROM memberships WHERE org_id = $1 AND user_id = $2`, orgID, userID).
		Scan(&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, mapError(err)
	}
	return &m, nil
}

// List returns memberships of an org joined with user emails.
func (r *MembershipRepository) List(ctx context.Context, orgID string) ([]models.Membership, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT m.id, m.org_id, m.user_id, u.email, m.role, m.created_at, m.updated_at
		 FROM memberships m JOIN users u ON u.id = m.user_id
		 WHERE m.org_id = $1 ORDER BY m.created_at`, orgID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []models.Membership
	for rows.Next() {
		var m models.Membership
		if err := rows.Scan(&m.ID, &m.OrgID, &m.UserID, &m.Email, &m.Role, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// CountAdmins returns the number of admins in an org.
func (r *MembershipRepository) CountAdmins(ctx context.Context, orgID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT count(*) FROM memberships WHERE org_id = $1 AND role = 'admin'`, orgID).Scan(&n)
	return n, mapError(err)
}

// Delete removes a membership.
func (r *MembershipRepository) Delete(ctx context.Context, orgID, userID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM memberships WHERE org_id = $1 AND user_id = $2`, orgID, userID)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

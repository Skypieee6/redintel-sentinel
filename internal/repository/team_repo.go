package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

// TeamRepository persists teams and team membership.
type TeamRepository struct{ pool *pgxpool.Pool }

const teamColumns = "id, org_id, name, description, created_at, updated_at"

func scanTeam(row interface{ Scan(...any) error }) (*models.Team, error) {
	var t models.Team
	if err := row.Scan(&t.ID, &t.OrgID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, mapError(err)
	}
	return &t, nil
}

// Create inserts a team.
func (r *TeamRepository) Create(ctx context.Context, orgID, name, description string) (*models.Team, error) {
	return scanTeam(r.pool.QueryRow(ctx,
		`INSERT INTO teams (org_id, name, description) VALUES ($1, $2, $3) RETURNING `+teamColumns,
		orgID, name, description))
}

// Get returns a team by id within an org.
func (r *TeamRepository) Get(ctx context.Context, orgID, id string) (*models.Team, error) {
	return scanTeam(r.pool.QueryRow(ctx,
		`SELECT `+teamColumns+` FROM teams WHERE id = $1 AND org_id = $2`, id, orgID))
}

// List returns the teams of an org.
func (r *TeamRepository) List(ctx context.Context, orgID string) ([]models.Team, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+teamColumns+` FROM teams WHERE org_id = $1 ORDER BY name`, orgID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []models.Team
	for rows.Next() {
		t, err := scanTeam(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

// Update changes a team's name and description.
func (r *TeamRepository) Update(ctx context.Context, orgID, id, name, description string) (*models.Team, error) {
	return scanTeam(r.pool.QueryRow(ctx,
		`UPDATE teams SET name = $3, description = $4, updated_at = now()
		 WHERE id = $1 AND org_id = $2 RETURNING `+teamColumns, id, orgID, name, description))
}

// Delete removes a team.
func (r *TeamRepository) Delete(ctx context.Context, orgID, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM teams WHERE id = $1 AND org_id = $2`, id, orgID)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// AddMember adds a user to a team.
func (r *TeamRepository) AddMember(ctx context.Context, teamID, userID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`, teamID, userID)
	return mapError(err)
}

// RemoveMember removes a user from a team.
func (r *TeamRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, userID)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListMembers returns the members of a team with emails.
func (r *TeamRepository) ListMembers(ctx context.Context, teamID string) ([]models.Membership, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.email, COALESCE(m.role, 'viewer'), tm.created_at
		 FROM team_members tm
		 JOIN users u ON u.id = tm.user_id
		 LEFT JOIN teams t ON t.id = tm.team_id
		 LEFT JOIN memberships m ON m.user_id = tm.user_id AND m.org_id = t.org_id
		 WHERE tm.team_id = $1 ORDER BY tm.created_at`, teamID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []models.Membership
	for rows.Next() {
		var m models.Membership
		if err := rows.Scan(&m.UserID, &m.Email, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

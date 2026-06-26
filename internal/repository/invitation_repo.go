package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

// InvitationRepository persists organization invitations.
type InvitationRepository struct{ pool *pgxpool.Pool }

const invColumns = "id, org_id, email, role, status, COALESCE(invited_by::text, ''), expires_at, created_at"

func scanInvitation(row interface{ Scan(...any) error }) (*models.Invitation, error) {
	var i models.Invitation
	if err := row.Scan(&i.ID, &i.OrgID, &i.Email, &i.Role, &i.Status, &i.InvitedBy, &i.ExpiresAt, &i.CreatedAt); err != nil {
		return nil, mapError(err)
	}
	return &i, nil
}

// Create inserts an invitation.
func (r *InvitationRepository) Create(ctx context.Context, orgID, email string, role models.Role, tokenHash, invitedBy string, expiresAt time.Time) (*models.Invitation, error) {
	return scanInvitation(r.pool.QueryRow(ctx,
		`INSERT INTO invitations (org_id, email, role, token_hash, invited_by, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+invColumns,
		orgID, email, string(role), tokenHash, invitedBy, expiresAt))
}

// List returns invitations for an org.
func (r *InvitationRepository) List(ctx context.Context, orgID string) ([]models.Invitation, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+invColumns+` FROM invitations WHERE org_id = $1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []models.Invitation
	for rows.Next() {
		i, err := scanInvitation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *i)
	}
	return out, rows.Err()
}

// GetPendingByToken returns a pending, non-expired invitation by token hash.
func (r *InvitationRepository) GetPendingByToken(ctx context.Context, tokenHash string) (*models.Invitation, error) {
	return scanInvitation(r.pool.QueryRow(ctx,
		`SELECT `+invColumns+` FROM invitations
		 WHERE token_hash = $1 AND status = 'pending' AND expires_at > now()`, tokenHash))
}

// MarkAccepted sets an invitation status to accepted.
func (r *InvitationRepository) MarkAccepted(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE invitations SET status = 'accepted' WHERE id = $1`, id)
	return mapError(err)
}

// Revoke marks an invitation revoked within an org.
func (r *InvitationRepository) Revoke(ctx context.Context, orgID, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE invitations SET status = 'revoked' WHERE id = $1 AND org_id = $2 AND status = 'pending'`, id, orgID)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

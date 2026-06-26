package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

// AuditRepository persists audit log entries.
type AuditRepository struct{ pool *pgxpool.Pool }

// Create inserts an audit log entry.
func (r *AuditRepository) Create(ctx context.Context, e *models.AuditLog) error {
	meta := e.Metadata
	if meta == nil {
		meta = map[string]any{}
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO audit_logs (org_id, actor_id, action, target_type, target_id, metadata, ip)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		e.OrgID, e.ActorID, e.Action, e.TargetType, e.TargetID, meta, e.IP)
	return mapError(err)
}

// AuditFilter narrows an audit log query.
type AuditFilter struct {
	OrgID  string
	Action string
	Limit  int
	Offset int
}

// List returns audit entries matching the filter (most recent first).
func (r *AuditRepository) List(ctx context.Context, f AuditFilter) ([]models.AuditLog, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, org_id, actor_id, action, target_type, target_id, metadata, ip, created_at
		 FROM audit_logs
		 WHERE ($1 = '' OR org_id = $1::uuid)
		   AND ($2 = '' OR action = $2)
		 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		f.OrgID, f.Action, f.Limit, f.Offset)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []models.AuditLog
	for rows.Next() {
		var e models.AuditLog
		if err := rows.Scan(&e.ID, &e.OrgID, &e.ActorID, &e.Action,
			&e.TargetType, &e.TargetID, &e.Metadata, &e.IP, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
)

// AuditService records and queries audit events.
type AuditService struct {
	repo *repository.AuditRepository
	log  *zap.Logger
}

// Record persists an audit event. Failures are logged, never propagated, so an
// audit write never breaks the primary operation.
func (s *AuditService) Record(ctx context.Context, action, actorID, orgID, targetType, targetID, ip string, metadata map[string]any) {
	e := &models.AuditLog{
		OrgID:      strPtr(orgID),
		ActorID:    strPtr(actorID),
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Metadata:   metadata,
		IP:         ip,
	}
	if err := s.repo.Create(ctx, e); err != nil {
		s.log.Warn("failed to write audit log", zap.String("action", action), zap.Error(err))
	}
}

// List returns audit events matching the filter.
func (s *AuditService) List(ctx context.Context, orgID, action string, limit, offset int) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.List(ctx, repository.AuditFilter{OrgID: orgID, Action: action, Limit: limit, Offset: offset})
}

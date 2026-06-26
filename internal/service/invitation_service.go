package service

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/Skypieee6/redintel-sentinel/internal/auth"
	"github.com/Skypieee6/redintel-sentinel/internal/config"
	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
)

// InvitationService manages organization invitations.
type InvitationService struct {
	repos *repository.Repositories
	cfg   config.AuthConfig
	audit *AuditService
	log   *zap.Logger
}

// Create issues an invitation and returns it with the raw token (shown once).
func (s *InvitationService) Create(ctx context.Context, actorID, orgID, email string, role models.Role, ip string) (*models.Invitation, error) {
	email = normalizeEmail(email)
	if email == "" {
		return nil, wrap(ErrValidation, "email is required")
	}
	if !role.Valid() {
		return nil, wrap(ErrValidation, "invalid role")
	}
	raw, err := auth.GenerateOpaqueToken()
	if err != nil {
		return nil, err
	}
	inv, err := s.repos.Invitations.Create(ctx, orgID, email, role,
		auth.HashToken(raw), actorID, time.Now().Add(s.cfg.InvitationTTL))
	if err != nil {
		return nil, mapRepoErr(err, "invitation")
	}
	inv.Token = raw
	s.audit.Record(ctx, "invitation.created", actorID, orgID, "invitation", inv.ID, ip, map[string]any{"email": email, "role": string(role)})
	s.log.Info("organization invitation issued",
		zap.String("org_id", orgID), zap.String("email", email), zap.String("invite_token", raw))
	return inv, nil
}

// List returns an org's invitations.
func (s *InvitationService) List(ctx context.Context, orgID string) ([]models.Invitation, error) {
	return s.repos.Invitations.List(ctx, orgID)
}

// Accept consumes an invitation token, joining the user to the org.
func (s *InvitationService) Accept(ctx context.Context, token, userID, userEmail, ip string) (*models.Membership, error) {
	inv, err := s.repos.Invitations.GetPendingByToken(ctx, auth.HashToken(token))
	if err != nil {
		return nil, wrap(ErrInvalidCredentials, "invalid or expired invitation")
	}
	if !strings.EqualFold(inv.Email, userEmail) {
		return nil, wrap(ErrForbidden, "this invitation was issued to a different email")
	}
	m, err := s.repos.Memberships.Upsert(ctx, inv.OrgID, userID, inv.Role)
	if err != nil {
		return nil, err
	}
	if err := s.repos.Invitations.MarkAccepted(ctx, inv.ID); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "invitation.accepted", userID, inv.OrgID, "invitation", inv.ID, ip, nil)
	return m, nil
}

// Revoke cancels a pending invitation.
func (s *InvitationService) Revoke(ctx context.Context, actorID, orgID, id, ip string) error {
	if err := s.repos.Invitations.Revoke(ctx, orgID, id); err != nil {
		return mapRepoErr(err, "invitation")
	}
	s.audit.Record(ctx, "invitation.revoked", actorID, orgID, "invitation", id, ip, nil)
	return nil
}

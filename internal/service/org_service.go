package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
)

// OrgService manages organizations and their memberships.
type OrgService struct {
	repos *repository.Repositories
	audit *AuditService
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	s := slugRe.ReplaceAllString(strings.ToLower(strings.TrimSpace(name)), "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "org"
	}
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return s + "-" + hex.EncodeToString(b)
}

// Create makes a new org and assigns the creator as admin.
func (s *OrgService) Create(ctx context.Context, userID, name, ip string) (*models.Organization, error) {
	if strings.TrimSpace(name) == "" {
		return nil, wrap(ErrValidation, "organization name is required")
	}
	org, err := s.repos.Orgs.Create(ctx, name, slugify(name), userID)
	if err != nil {
		return nil, mapRepoErr(err, "organization")
	}
	if _, err := s.repos.Memberships.Upsert(ctx, org.ID, userID, models.RoleAdmin); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "org.created", userID, org.ID, "organization", org.ID, ip, map[string]any{"name": name})
	return org, nil
}

// ListForUser returns orgs the user belongs to.
func (s *OrgService) ListForUser(ctx context.Context, userID string) ([]models.Organization, error) {
	return s.repos.Orgs.ListForUser(ctx, userID)
}

// Get returns an org by id.
func (s *OrgService) Get(ctx context.Context, id string) (*models.Organization, error) {
	o, err := s.repos.Orgs.GetByID(ctx, id)
	if err != nil {
		return nil, mapRepoErr(err, "organization")
	}
	return o, nil
}

// Update renames an org.
func (s *OrgService) Update(ctx context.Context, userID, id, name, ip string) (*models.Organization, error) {
	if strings.TrimSpace(name) == "" {
		return nil, wrap(ErrValidation, "organization name is required")
	}
	o, err := s.repos.Orgs.Update(ctx, id, name)
	if err != nil {
		return nil, mapRepoErr(err, "organization")
	}
	s.audit.Record(ctx, "org.updated", userID, id, "organization", id, ip, nil)
	return o, nil
}

// Delete removes an org.
func (s *OrgService) Delete(ctx context.Context, userID, id, ip string) error {
	if err := s.repos.Orgs.Delete(ctx, id); err != nil {
		return mapRepoErr(err, "organization")
	}
	s.audit.Record(ctx, "org.deleted", userID, id, "organization", id, ip, nil)
	return nil
}

// Membership returns a user's membership in an org (or ErrNotFound).
func (s *OrgService) Membership(ctx context.Context, orgID, userID string) (*models.Membership, error) {
	m, err := s.repos.Memberships.Get(ctx, orgID, userID)
	if err != nil {
		return nil, mapRepoErr(err, "membership")
	}
	return m, nil
}

// ListMembers returns the members of an org.
func (s *OrgService) ListMembers(ctx context.Context, orgID string) ([]models.Membership, error) {
	return s.repos.Memberships.List(ctx, orgID)
}

// SetMemberRole updates the role of an existing member by email.
func (s *OrgService) SetMemberRole(ctx context.Context, actorID, orgID, email string, role models.Role, ip string) (*models.Membership, error) {
	if !role.Valid() {
		return nil, wrap(ErrValidation, "invalid role")
	}
	u, err := s.repos.Users.GetByEmail(ctx, email)
	if err != nil {
		return nil, wrap(ErrNotFound, "no user with email %s", email)
	}
	m, err := s.repos.Memberships.Upsert(ctx, orgID, u.ID, role)
	if err != nil {
		return nil, err
	}
	m.Email = u.Email
	s.audit.Record(ctx, "org.member_role_set", actorID, orgID, "membership", u.ID, ip, map[string]any{"role": string(role)})
	return m, nil
}

// RemoveMember removes a user from an org, refusing to remove the last admin.
func (s *OrgService) RemoveMember(ctx context.Context, actorID, orgID, targetUserID, ip string) error {
	m, err := s.repos.Memberships.Get(ctx, orgID, targetUserID)
	if err != nil {
		return mapRepoErr(err, "membership")
	}
	if m.Role == models.RoleAdmin {
		n, err := s.repos.Memberships.CountAdmins(ctx, orgID)
		if err != nil {
			return err
		}
		if n <= 1 {
			return wrap(ErrValidation, "cannot remove the last admin of the organization")
		}
	}
	if err := s.repos.Memberships.Delete(ctx, orgID, targetUserID); err != nil {
		return mapRepoErr(err, "membership")
	}
	s.audit.Record(ctx, "org.member_removed", actorID, orgID, "membership", targetUserID, ip, nil)
	return nil
}

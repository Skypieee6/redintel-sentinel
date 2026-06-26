package service

import (
	"context"
	"strings"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
)

// TeamService manages teams within an organization.
type TeamService struct {
	repos *repository.Repositories
	audit *AuditService
}

// Create makes a team.
func (s *TeamService) Create(ctx context.Context, actorID, orgID, name, description, ip string) (*models.Team, error) {
	if strings.TrimSpace(name) == "" {
		return nil, wrap(ErrValidation, "team name is required")
	}
	t, err := s.repos.Teams.Create(ctx, orgID, name, description)
	if err != nil {
		return nil, mapRepoErr(err, "team")
	}
	s.audit.Record(ctx, "team.created", actorID, orgID, "team", t.ID, ip, map[string]any{"name": name})
	return t, nil
}

// List returns an org's teams.
func (s *TeamService) List(ctx context.Context, orgID string) ([]models.Team, error) {
	return s.repos.Teams.List(ctx, orgID)
}

// Get returns a team.
func (s *TeamService) Get(ctx context.Context, orgID, id string) (*models.Team, error) {
	t, err := s.repos.Teams.Get(ctx, orgID, id)
	if err != nil {
		return nil, mapRepoErr(err, "team")
	}
	return t, nil
}

// Update changes a team.
func (s *TeamService) Update(ctx context.Context, actorID, orgID, id, name, description, ip string) (*models.Team, error) {
	if strings.TrimSpace(name) == "" {
		return nil, wrap(ErrValidation, "team name is required")
	}
	t, err := s.repos.Teams.Update(ctx, orgID, id, name, description)
	if err != nil {
		return nil, mapRepoErr(err, "team")
	}
	s.audit.Record(ctx, "team.updated", actorID, orgID, "team", id, ip, nil)
	return t, nil
}

// Delete removes a team.
func (s *TeamService) Delete(ctx context.Context, actorID, orgID, id, ip string) error {
	if err := s.repos.Teams.Delete(ctx, orgID, id); err != nil {
		return mapRepoErr(err, "team")
	}
	s.audit.Record(ctx, "team.deleted", actorID, orgID, "team", id, ip, nil)
	return nil
}

// AddMember adds an org member (by email) to a team.
func (s *TeamService) AddMember(ctx context.Context, actorID, orgID, teamID, email, ip string) error {
	if _, err := s.repos.Teams.Get(ctx, orgID, teamID); err != nil {
		return mapRepoErr(err, "team")
	}
	u, err := s.repos.Users.GetByEmail(ctx, email)
	if err != nil {
		return wrap(ErrNotFound, "no user with email %s", email)
	}
	if _, err := s.repos.Memberships.Get(ctx, orgID, u.ID); err != nil {
		return wrap(ErrValidation, "user is not a member of this organization")
	}
	if err := s.repos.Teams.AddMember(ctx, teamID, u.ID); err != nil {
		return mapRepoErr(err, "team member")
	}
	s.audit.Record(ctx, "team.member_added", actorID, orgID, "team", teamID, ip, map[string]any{"user_id": u.ID})
	return nil
}

// RemoveMember removes a user from a team.
func (s *TeamService) RemoveMember(ctx context.Context, actorID, orgID, teamID, userID, ip string) error {
	if _, err := s.repos.Teams.Get(ctx, orgID, teamID); err != nil {
		return mapRepoErr(err, "team")
	}
	if err := s.repos.Teams.RemoveMember(ctx, teamID, userID); err != nil {
		return mapRepoErr(err, "team member")
	}
	s.audit.Record(ctx, "team.member_removed", actorID, orgID, "team", teamID, ip, map[string]any{"user_id": userID})
	return nil
}

// ListMembers lists a team's members.
func (s *TeamService) ListMembers(ctx context.Context, orgID, teamID string) ([]models.Membership, error) {
	if _, err := s.repos.Teams.Get(ctx, orgID, teamID); err != nil {
		return nil, mapRepoErr(err, "team")
	}
	return s.repos.Teams.ListMembers(ctx, teamID)
}

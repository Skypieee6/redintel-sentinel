package service

import (
	"context"
	"strings"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
)

// ProjectService manages projects and their access controls.
type ProjectService struct {
	repos *repository.Repositories
	audit *AuditService
}

// canManage reports whether a user may modify a project: org managers/admins,
// the project owner, or an explicit project member with manager+ role.
func (s *ProjectService) canManage(ctx context.Context, p *models.Project, userID string, orgRole models.Role) bool {
	if orgRole.AtLeast(models.RoleManager) {
		return true
	}
	if p.OwnerID == userID {
		return true
	}
	members, err := s.repos.Projects.ListMembers(ctx, p.ID)
	if err != nil {
		return false
	}
	for _, m := range members {
		if m.UserID == userID && m.Role.AtLeast(models.RoleManager) {
			return true
		}
	}
	return false
}

// Create makes a project; requires org role manager or above.
func (s *ProjectService) Create(ctx context.Context, actorID, orgID string, orgRole models.Role, name, description, ip string) (*models.Project, error) {
	if !orgRole.AtLeast(models.RoleManager) {
		return nil, wrap(ErrForbidden, "manager role or above is required to create projects")
	}
	if strings.TrimSpace(name) == "" {
		return nil, wrap(ErrValidation, "project name is required")
	}
	p, err := s.repos.Projects.Create(ctx, orgID, name, description, actorID)
	if err != nil {
		return nil, mapRepoErr(err, "project")
	}
	_ = s.repos.Projects.AddMember(ctx, p.ID, actorID, models.RoleAdmin)
	s.audit.Record(ctx, "project.created", actorID, orgID, "project", p.ID, ip, map[string]any{"name": name})
	return p, nil
}

// List returns an org's projects.
func (s *ProjectService) List(ctx context.Context, orgID string) ([]models.Project, error) {
	return s.repos.Projects.List(ctx, orgID)
}

// Get returns a project.
func (s *ProjectService) Get(ctx context.Context, orgID, id string) (*models.Project, error) {
	p, err := s.repos.Projects.Get(ctx, orgID, id)
	if err != nil {
		return nil, mapRepoErr(err, "project")
	}
	return p, nil
}

// Update modifies a project after access checks.
func (s *ProjectService) Update(ctx context.Context, actorID, orgID, id string, orgRole models.Role, name, description, status, ip string) (*models.Project, error) {
	p, err := s.repos.Projects.Get(ctx, orgID, id)
	if err != nil {
		return nil, mapRepoErr(err, "project")
	}
	if !s.canManage(ctx, p, actorID, orgRole) {
		return nil, wrap(ErrForbidden, "you do not have permission to modify this project")
	}
	if strings.TrimSpace(name) == "" {
		name = p.Name
	}
	if status == "" {
		status = p.Status
	}
	updated, err := s.repos.Projects.Update(ctx, orgID, id, name, description, status)
	if err != nil {
		return nil, mapRepoErr(err, "project")
	}
	s.audit.Record(ctx, "project.updated", actorID, orgID, "project", id, ip, nil)
	return updated, nil
}

// Delete removes a project; requires org admin or project ownership.
func (s *ProjectService) Delete(ctx context.Context, actorID, orgID, id string, orgRole models.Role, ip string) error {
	p, err := s.repos.Projects.Get(ctx, orgID, id)
	if err != nil {
		return mapRepoErr(err, "project")
	}
	if !orgRole.AtLeast(models.RoleAdmin) && p.OwnerID != actorID {
		return wrap(ErrForbidden, "only an org admin or the project owner may delete this project")
	}
	if err := s.repos.Projects.Delete(ctx, orgID, id); err != nil {
		return mapRepoErr(err, "project")
	}
	s.audit.Record(ctx, "project.deleted", actorID, orgID, "project", id, ip, nil)
	return nil
}

// AddMember grants a user access to a project.
func (s *ProjectService) AddMember(ctx context.Context, actorID, orgID, projectID string, orgRole models.Role, email string, role models.Role, ip string) error {
	p, err := s.repos.Projects.Get(ctx, orgID, projectID)
	if err != nil {
		return mapRepoErr(err, "project")
	}
	if !s.canManage(ctx, p, actorID, orgRole) {
		return wrap(ErrForbidden, "you do not have permission to manage this project's members")
	}
	if !role.Valid() {
		return wrap(ErrValidation, "invalid role")
	}
	u, err := s.repos.Users.GetByEmail(ctx, email)
	if err != nil {
		return wrap(ErrNotFound, "no user with email %s", email)
	}
	if _, err := s.repos.Memberships.Get(ctx, orgID, u.ID); err != nil {
		return wrap(ErrValidation, "user is not a member of this organization")
	}
	if err := s.repos.Projects.AddMember(ctx, projectID, u.ID, role); err != nil {
		return mapRepoErr(err, "project member")
	}
	s.audit.Record(ctx, "project.member_added", actorID, orgID, "project", projectID, ip, map[string]any{"user_id": u.ID, "role": string(role)})
	return nil
}

// RemoveMember revokes a user's project access.
func (s *ProjectService) RemoveMember(ctx context.Context, actorID, orgID, projectID string, orgRole models.Role, userID, ip string) error {
	p, err := s.repos.Projects.Get(ctx, orgID, projectID)
	if err != nil {
		return mapRepoErr(err, "project")
	}
	if !s.canManage(ctx, p, actorID, orgRole) {
		return wrap(ErrForbidden, "you do not have permission to manage this project's members")
	}
	if err := s.repos.Projects.RemoveMember(ctx, projectID, userID); err != nil {
		return mapRepoErr(err, "project member")
	}
	s.audit.Record(ctx, "project.member_removed", actorID, orgID, "project", projectID, ip, map[string]any{"user_id": userID})
	return nil
}

// SetArchived archives or restores a project; requires manager+ or ownership.
func (s *ProjectService) SetArchived(ctx context.Context, actorID, orgID, id string, orgRole models.Role, archived bool, ip string) (*models.Project, error) {
	p, err := s.repos.Projects.Get(ctx, orgID, id)
	if err != nil {
		return nil, mapRepoErr(err, "project")
	}
	if !s.canManage(ctx, p, actorID, orgRole) {
		return nil, wrap(ErrForbidden, "you do not have permission to modify this project")
	}
	status := "active"
	action := "project.unarchived"
	if archived {
		status, action = "archived", "project.archived"
	}
	updated, err := s.repos.Projects.SetStatus(ctx, orgID, id, status)
	if err != nil {
		return nil, mapRepoErr(err, "project")
	}
	s.audit.Record(ctx, action, actorID, orgID, "project", id, ip, nil)
	return updated, nil
}

// ListMembers lists explicit members of a project.
func (s *ProjectService) ListMembers(ctx context.Context, orgID, projectID string) ([]models.ProjectMember, error) {
	if _, err := s.repos.Projects.Get(ctx, orgID, projectID); err != nil {
		return nil, mapRepoErr(err, "project")
	}
	return s.repos.Projects.ListMembers(ctx, projectID)
}

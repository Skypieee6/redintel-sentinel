package service

import (
	"go.uber.org/zap"

	"github.com/Skypieee6/redintel-sentinel/internal/auth"
	"github.com/Skypieee6/redintel-sentinel/internal/cache"
	"github.com/Skypieee6/redintel-sentinel/internal/config"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
)

// Services aggregates all application services.
type Services struct {
	Auth       *AuthService
	User       *UserService
	Org        *OrgService
	Team       *TeamService
	Invitation *InvitationService
	Project    *ProjectService
	Audit      *AuditService
}

// New constructs the service set.
func New(repos *repository.Repositories, jwt *auth.JWTManager, redis *cache.Redis, cfg config.AuthConfig, log *zap.Logger) *Services {
	audit := &AuditService{repo: repos.Audit, log: log}
	return &Services{
		Auth:       &AuthService{repos: repos, jwt: jwt, redis: redis, cfg: cfg, audit: audit, log: log},
		User:       &UserService{repos: repos},
		Org:        &OrgService{repos: repos, audit: audit},
		Team:       &TeamService{repos: repos, audit: audit},
		Invitation: &InvitationService{repos: repos, cfg: cfg, audit: audit, log: log},
		Project:    &ProjectService{repos: repos, audit: audit},
		Audit:      audit,
	}
}

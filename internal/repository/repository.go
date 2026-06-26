// Package repository implements the PostgreSQL persistence layer.
package repository

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors translated from database driver errors.
var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource already exists")
)

// mapError normalizes pgx errors into repository sentinel errors.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return err
}

// Repositories aggregates all concrete repositories.
type Repositories struct {
	Users       *UserRepository
	Tokens      *TokenRepository
	APIKeys     *APIKeyRepository
	Orgs        *OrgRepository
	Memberships *MembershipRepository
	Teams       *TeamRepository
	Invitations *InvitationRepository
	Projects    *ProjectRepository
	Audit       *AuditRepository
}

// New builds the repository set bound to a pgx pool.
func New(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		Users:       &UserRepository{pool: pool},
		Tokens:      &TokenRepository{pool: pool},
		APIKeys:     &APIKeyRepository{pool: pool},
		Orgs:        &OrgRepository{pool: pool},
		Memberships: &MembershipRepository{pool: pool},
		Teams:       &TeamRepository{pool: pool},
		Invitations: &InvitationRepository{pool: pool},
		Projects:    &ProjectRepository{pool: pool},
		Audit:       &AuditRepository{pool: pool},
	}
}

package service

import (
	"context"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
)

// UserService exposes user account queries.
type UserService struct{ repos *repository.Repositories }

// GetByID returns a user by id.
func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
	u, err := s.repos.Users.GetByID(ctx, id)
	if err != nil {
		return nil, mapRepoErr(err, "user")
	}
	return u, nil
}

// List returns users (superadmin only, enforced at handler).
func (s *UserService) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repos.Users.List(ctx, limit, offset)
}

// UpdateProfile updates the caller's own profile.
func (s *UserService) UpdateProfile(ctx context.Context, id, fullName string) (*models.User, error) {
	u, err := s.repos.Users.UpdateProfile(ctx, id, fullName)
	if err != nil {
		return nil, mapRepoErr(err, "user")
	}
	return u, nil
}

// mapRepoErr converts repository sentinel errors into service errors.
func mapRepoErr(err error, resource string) error {
	switch err {
	case repository.ErrNotFound:
		return wrap(ErrNotFound, "%s not found", resource)
	case repository.ErrConflict:
		return wrap(ErrConflict, "%s already exists", resource)
	default:
		return err
	}
}

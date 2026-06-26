package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

// UserRepository persists user accounts.
type UserRepository struct{ pool *pgxpool.Pool }

const userColumns = "id, email, password_hash, full_name, is_active, is_superadmin, created_at, updated_at"

func scanUser(row interface {
	Scan(dest ...any) error
}) (*models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName,
		&u.IsActive, &u.IsSuperadmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, mapError(err)
	}
	return &u, nil
}

// Create inserts a new user and returns the stored record.
func (r *UserRepository) Create(ctx context.Context, u *models.User) (*models.User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, full_name, is_active, is_superadmin)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+userColumns,
		u.Email, u.PasswordHash, u.FullName, u.IsActive, u.IsSuperadmin)
	return scanUser(row)
}

// GetByID returns a user by id.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id)
	return scanUser(row)
}

// GetByEmail returns a user by email (case-insensitive).
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE lower(email) = lower($1)`, email)
	return scanUser(row)
}

// UpdatePassword sets a new password hash for a user.
func (r *UserRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`, id, passwordHash)
	return mapError(err)
}

// UpdateProfile updates mutable profile fields.
func (r *UserRepository) UpdateProfile(ctx context.Context, id, fullName string) (*models.User, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE users SET full_name = $2, updated_at = now() WHERE id = $1 RETURNING `+userColumns,
		id, fullName)
	return scanUser(row)
}

// List returns users ordered by creation time.
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+userColumns+` FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []models.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

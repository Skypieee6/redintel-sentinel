package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

// APIKeyRepository persists API keys (hashed).
type APIKeyRepository struct{ pool *pgxpool.Pool }

// Create stores a new API key record.
func (r *APIKeyRepository) Create(ctx context.Context, userID, name, prefix, keyHash string, expiresAt *time.Time) (*models.APIKey, error) {
	var k models.APIKey
	err := r.pool.QueryRow(ctx,
		`INSERT INTO api_keys (user_id, name, prefix, key_hash, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, name, prefix, last_used_at, expires_at, revoked, created_at`,
		userID, name, prefix, keyHash, expiresAt).
		Scan(&k.ID, &k.UserID, &k.Name, &k.Prefix, &k.LastUsedAt, &k.ExpiresAt, &k.Revoked, &k.CreatedAt)
	if err != nil {
		return nil, mapError(err)
	}
	return &k, nil
}

// ResolveUser returns the active user id owning a key by its hash, and touches
// last_used_at. Returns ErrNotFound for unknown/expired/revoked keys.
func (r *APIKeyRepository) ResolveUser(ctx context.Context, keyHash string) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx,
		`UPDATE api_keys SET last_used_at = now()
		 WHERE key_hash = $1 AND revoked = FALSE
		   AND (expires_at IS NULL OR expires_at > now())
		 RETURNING user_id`, keyHash).Scan(&userID)
	return userID, mapError(err)
}

// ListByUser returns a user's API keys.
func (r *APIKeyRepository) ListByUser(ctx context.Context, userID string) ([]models.APIKey, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, prefix, last_used_at, expires_at, revoked, created_at
		 FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []models.APIKey
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.Prefix,
			&k.LastUsedAt, &k.ExpiresAt, &k.Revoked, &k.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// Revoke marks a user's API key revoked.
func (r *APIKeyRepository) Revoke(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE api_keys SET revoked = TRUE WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

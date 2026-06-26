package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TokenRepository persists refresh and password-reset tokens (hashed).
type TokenRepository struct{ pool *pgxpool.Pool }

// CreateRefreshToken stores a hashed refresh token.
func (r *TokenRepository) CreateRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt)
	return mapError(err)
}

// GetRefreshToken returns the owning user id if the token is valid (not revoked
// and not expired), else ErrNotFound.
func (r *TokenRepository) GetRefreshToken(ctx context.Context, tokenHash string) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx,
		`SELECT user_id FROM refresh_tokens
		 WHERE token_hash = $1 AND revoked = FALSE AND expires_at > now()`,
		tokenHash).Scan(&userID)
	return userID, mapError(err)
}

// RevokeRefreshToken revokes a single refresh token by hash.
func (r *TokenRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1`, tokenHash)
	return mapError(err)
}

// RevokeAllRefreshTokens revokes every refresh token for a user.
func (r *TokenRepository) RevokeAllRefreshTokens(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1`, userID)
	return mapError(err)
}

// CreatePasswordReset stores a hashed password-reset token.
func (r *TokenRepository) CreatePasswordReset(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt)
	return mapError(err)
}

// ConsumePasswordReset validates and marks a reset token used, returning the
// owning user id.
func (r *TokenRepository) ConsumePasswordReset(ctx context.Context, tokenHash string) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx,
		`UPDATE password_reset_tokens SET used = TRUE
		 WHERE token_hash = $1 AND used = FALSE AND expires_at > now()
		 RETURNING user_id`, tokenHash).Scan(&userID)
	return userID, mapError(err)
}

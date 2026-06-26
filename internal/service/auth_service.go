package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/Skypieee6/redintel-sentinel/internal/auth"
	"github.com/Skypieee6/redintel-sentinel/internal/cache"
	"github.com/Skypieee6/redintel-sentinel/internal/config"
	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
)

const (
	maxLoginAttempts = 5
	lockoutWindow    = 15 * time.Minute
)

// TokenPair is an issued access + refresh token pair.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// AuthService handles registration, login, tokens, password reset and API keys.
type AuthService struct {
	repos *repository.Repositories
	jwt   *auth.JWTManager
	redis *cache.Redis
	cfg   config.AuthConfig
	audit *AuditService
	log   *zap.Logger
}

func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func (s *AuthService) issueTokens(ctx context.Context, u *models.User) (*TokenPair, error) {
	access, expiresAt, err := s.jwt.GenerateAccessToken(u.ID, u.Email)
	if err != nil {
		return nil, err
	}
	refreshRaw, err := auth.GenerateOpaqueToken()
	if err != nil {
		return nil, err
	}
	if err := s.repos.Tokens.CreateRefreshToken(ctx, u.ID, auth.HashToken(refreshRaw),
		time.Now().Add(s.cfg.RefreshTokenTTL)); err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refreshRaw,
		TokenType:    "Bearer",
		ExpiresAt:    expiresAt,
	}, nil
}

// Register creates a new user account and issues tokens.
func (s *AuthService) Register(ctx context.Context, email, password, fullName, ip string) (*models.User, *TokenPair, error) {
	email = normalizeEmail(email)
	if email == "" || len(password) < 8 {
		return nil, nil, wrap(ErrValidation, "email is required and password must be at least 8 characters")
	}
	hash, err := auth.HashPassword(password, s.cfg.BcryptCost)
	if err != nil {
		return nil, nil, err
	}
	u, err := s.repos.Users.Create(ctx, &models.User{
		Email: email, PasswordHash: hash, FullName: fullName, IsActive: true,
	})
	if err != nil {
		if err == repository.ErrConflict {
			return nil, nil, wrap(ErrConflict, "an account with this email already exists")
		}
		return nil, nil, err
	}
	tokens, err := s.issueTokens(ctx, u)
	if err != nil {
		return nil, nil, err
	}
	s.audit.Record(ctx, "auth.register", u.ID, "", "user", u.ID, ip, nil)
	return u, tokens, nil
}

// Login authenticates a user and issues tokens, with brute-force protection.
func (s *AuthService) Login(ctx context.Context, email, password, ip string) (*models.User, *TokenPair, error) {
	email = normalizeEmail(email)
	if s.isLockedOut(ctx, email) {
		return nil, nil, wrap(ErrRateLimited, "too many failed attempts, try again later")
	}
	u, err := s.repos.Users.GetByEmail(ctx, email)
	if err != nil || !auth.VerifyPassword(u.PasswordHash, password) {
		s.recordFailure(ctx, email)
		s.audit.Record(ctx, "auth.login_failed", "", "", "user", email, ip, nil)
		return nil, nil, wrap(ErrInvalidCredentials, "invalid email or password")
	}
	if !u.IsActive {
		return nil, nil, wrap(ErrInactive, "this account has been deactivated")
	}
	s.clearFailures(ctx, email)
	tokens, err := s.issueTokens(ctx, u)
	if err != nil {
		return nil, nil, err
	}
	s.audit.Record(ctx, "auth.login", u.ID, "", "user", u.ID, ip, nil)
	return u, tokens, nil
}

// Refresh rotates a refresh token, returning a new token pair.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*models.User, *TokenPair, error) {
	hash := auth.HashToken(refreshToken)
	userID, err := s.repos.Tokens.GetRefreshToken(ctx, hash)
	if err != nil {
		return nil, nil, wrap(ErrInvalidCredentials, "invalid or expired refresh token")
	}
	u, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, wrap(ErrInvalidCredentials, "invalid refresh token")
	}
	if !u.IsActive {
		return nil, nil, wrap(ErrInactive, "this account has been deactivated")
	}
	// Rotate: revoke the presented token before issuing a new pair.
	_ = s.repos.Tokens.RevokeRefreshToken(ctx, hash)
	tokens, err := s.issueTokens(ctx, u)
	if err != nil {
		return nil, nil, err
	}
	return u, tokens, nil
}

// Logout revokes the presented refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshToken, actorID, ip string) error {
	if refreshToken != "" {
		_ = s.repos.Tokens.RevokeRefreshToken(ctx, auth.HashToken(refreshToken))
	}
	s.audit.Record(ctx, "auth.logout", actorID, "", "user", actorID, ip, nil)
	return nil
}

// ForgotPassword creates a reset token. The raw token is returned so the caller
// can deliver it (here it is logged); responses never reveal account existence.
func (s *AuthService) ForgotPassword(ctx context.Context, email, ip string) {
	email = normalizeEmail(email)
	u, err := s.repos.Users.GetByEmail(ctx, email)
	if err != nil {
		return
	}
	raw, err := auth.GenerateOpaqueToken()
	if err != nil {
		return
	}
	if err := s.repos.Tokens.CreatePasswordReset(ctx, u.ID, auth.HashToken(raw),
		time.Now().Add(s.cfg.PasswordResetTTL)); err != nil {
		return
	}
	s.audit.Record(ctx, "auth.password_reset_requested", u.ID, "", "user", u.ID, ip, nil)
	// In production this would be emailed. For now it is logged for operators.
	s.log.Info("password reset token issued",
		zap.String("email", email), zap.String("reset_token", raw))
}

// ResetPassword consumes a reset token and sets a new password.
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword, ip string) error {
	if len(newPassword) < 8 {
		return wrap(ErrValidation, "password must be at least 8 characters")
	}
	userID, err := s.repos.Tokens.ConsumePasswordReset(ctx, auth.HashToken(token))
	if err != nil {
		return wrap(ErrInvalidCredentials, "invalid or expired reset token")
	}
	hash, err := auth.HashPassword(newPassword, s.cfg.BcryptCost)
	if err != nil {
		return err
	}
	if err := s.repos.Users.UpdatePassword(ctx, userID, hash); err != nil {
		return err
	}
	_ = s.repos.Tokens.RevokeAllRefreshTokens(ctx, userID)
	s.audit.Record(ctx, "auth.password_reset", userID, "", "user", userID, ip, nil)
	return nil
}

// ChangePassword changes the authenticated user's password.
func (s *AuthService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword, ip string) error {
	if len(newPassword) < 8 {
		return wrap(ErrValidation, "password must be at least 8 characters")
	}
	u, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return wrap(ErrNotFound, "user not found")
	}
	if !auth.VerifyPassword(u.PasswordHash, oldPassword) {
		return wrap(ErrInvalidCredentials, "current password is incorrect")
	}
	hash, err := auth.HashPassword(newPassword, s.cfg.BcryptCost)
	if err != nil {
		return err
	}
	if err := s.repos.Users.UpdatePassword(ctx, userID, hash); err != nil {
		return err
	}
	_ = s.repos.Tokens.RevokeAllRefreshTokens(ctx, userID)
	s.audit.Record(ctx, "auth.password_changed", userID, "", "user", userID, ip, nil)
	return nil
}

// CreateAPIKey issues a new API key for the user. The plaintext is returned once.
func (s *AuthService) CreateAPIKey(ctx context.Context, userID, name string, expiresAt *time.Time, ip string) (*models.APIKey, error) {
	if strings.TrimSpace(name) == "" {
		return nil, wrap(ErrValidation, "api key name is required")
	}
	raw, prefix, hash, err := auth.GenerateAPIKey()
	if err != nil {
		return nil, err
	}
	k, err := s.repos.APIKeys.Create(ctx, userID, name, prefix, hash, expiresAt)
	if err != nil {
		return nil, err
	}
	k.Secret = raw
	s.audit.Record(ctx, "auth.api_key_created", userID, "", "api_key", k.ID, ip, nil)
	return k, nil
}

// ListAPIKeys lists the user's API keys.
func (s *AuthService) ListAPIKeys(ctx context.Context, userID string) ([]models.APIKey, error) {
	return s.repos.APIKeys.ListByUser(ctx, userID)
}

// RevokeAPIKey revokes an API key owned by the user.
func (s *AuthService) RevokeAPIKey(ctx context.Context, userID, id, ip string) error {
	if err := s.repos.APIKeys.Revoke(ctx, userID, id); err != nil {
		return mapRepoErr(err, "api key")
	}
	s.audit.Record(ctx, "auth.api_key_revoked", userID, "", "api_key", id, ip, nil)
	return nil
}

// --- brute force protection (Redis-backed, best effort) ---

func (s *AuthService) attemptsKey(email string) string {
	return fmt.Sprintf("login_attempts:%s", email)
}

func (s *AuthService) isLockedOut(ctx context.Context, email string) bool {
	if s.redis == nil {
		return false
	}
	n, err := s.redis.Client.Get(ctx, s.attemptsKey(email)).Int()
	if err != nil {
		return false
	}
	return n >= maxLoginAttempts
}

func (s *AuthService) recordFailure(ctx context.Context, email string) {
	if s.redis == nil {
		return
	}
	key := s.attemptsKey(email)
	if n, err := s.redis.Client.Incr(ctx, key).Result(); err == nil && n == 1 {
		s.redis.Client.Expire(ctx, key, lockoutWindow)
	}
}

func (s *AuthService) clearFailures(ctx context.Context, email string) {
	if s.redis == nil {
		return
	}
	s.redis.Client.Del(ctx, s.attemptsKey(email))
}

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/internal/auth"
	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

// Context keys for values stored by the auth/RBAC middleware.
const (
	ctxUserKey    = "auth_user"
	ctxOrgRoleKey = "org_role"
	ctxOrgIDKey   = "org_id"
)

// Authenticate resolves the caller from a Bearer access token or an X-API-Key
// header and stores the *models.User in the context. Rejects unauthenticated
// or inactive users.
func Authenticate(jwt *auth.JWTManager, repos *repository.Repositories) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user *models.User

		if apiKey := strings.TrimSpace(c.GetHeader("X-API-Key")); apiKey != "" {
			userID, err := repos.APIKeys.ResolveUser(c.Request.Context(), auth.HashToken(apiKey))
			if err == nil {
				user, _ = repos.Users.GetByID(c.Request.Context(), userID)
			}
		} else if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			claims, err := jwt.ParseAccessToken(strings.TrimPrefix(h, "Bearer "))
			if err == nil {
				user, _ = repos.Users.GetByID(c.Request.Context(), claims.Subject)
			}
		}

		if user == nil {
			response.Error(c, http.StatusUnauthorized, "unauthorized", "authentication required")
			c.Abort()
			return
		}
		if !user.IsActive {
			response.Error(c, http.StatusForbidden, "inactive", "account is inactive")
			c.Abort()
			return
		}
		c.Set(ctxUserKey, user)
		c.Next()
	}
}

// CurrentUser returns the authenticated user from the context.
func CurrentUser(c *gin.Context) (*models.User, bool) {
	v, ok := c.Get(ctxUserKey)
	if !ok {
		return nil, false
	}
	u, ok := v.(*models.User)
	return u, ok
}

// membershipResolver is the subset of OrgService the RBAC middleware needs.
type membershipResolver interface {
	Membership(ctx context.Context, orgID, userID string) (*models.Membership, error)
}

// OrgContext loads the caller's role within the :orgID path organization and
// stores it in the context. Superadmins are granted admin role implicitly.
func OrgContext(orgs membershipResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "unauthorized", "authentication required")
			c.Abort()
			return
		}
		orgID := c.Param("orgID")
		c.Set(ctxOrgIDKey, orgID)

		if user.IsSuperadmin {
			c.Set(ctxOrgRoleKey, models.RoleAdmin)
			c.Next()
			return
		}
		m, err := orgs.Membership(c.Request.Context(), orgID, user.ID)
		if err != nil {
			response.Error(c, http.StatusForbidden, "forbidden", "you are not a member of this organization")
			c.Abort()
			return
		}
		c.Set(ctxOrgRoleKey, m.Role)
		c.Next()
	}
}

// CurrentOrgRole returns the caller's role in the path organization.
func CurrentOrgRole(c *gin.Context) models.Role {
	if v, ok := c.Get(ctxOrgRoleKey); ok {
		if r, ok := v.(models.Role); ok {
			return r
		}
	}
	return ""
}

// RequireOrgRole aborts unless the caller's org role is at least min.
func RequireOrgRole(min models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !CurrentOrgRole(c).AtLeast(min) {
			response.Error(c, http.StatusForbidden, "forbidden", "insufficient role for this action")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireSuperadmin aborts unless the caller is a platform superadmin.
func RequireSuperadmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok || !user.IsSuperadmin {
			response.Error(c, http.StatusForbidden, "forbidden", "superadmin privileges required")
			c.Abort()
			return
		}
		c.Next()
	}
}

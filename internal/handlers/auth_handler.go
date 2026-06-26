package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type forgotRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type resetRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type apiKeyRequest struct {
	Name      string     `json:"name" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type acceptInviteRequest struct {
	Token string `json:"token" binding:"required"`
}

// Register handles POST /auth/register.
func (a *API) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	user, tokens, err := a.svc.Auth.Register(c.Request.Context(), req.Email, req.Password, req.FullName, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, gin.H{"user": user, "tokens": tokens})
}

// Login handles POST /auth/login.
func (a *API) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	user, tokens, err := a.svc.Auth.Login(c.Request.Context(), req.Email, req.Password, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, gin.H{"user": user, "tokens": tokens})
}

// Refresh handles POST /auth/refresh.
func (a *API) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	user, tokens, err := a.svc.Auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, gin.H{"user": user, "tokens": tokens})
}

// Logout handles POST /auth/logout.
func (a *API) Logout(c *gin.Context) {
	var req refreshRequest
	_ = c.ShouldBindJSON(&req)
	_ = a.svc.Auth.Logout(c.Request.Context(), req.RefreshToken, a.user(c).ID, c.ClientIP())
	response.OK(c, gin.H{"message": "logged out"})
}

// Me handles GET /auth/me.
func (a *API) Me(c *gin.Context) {
	response.OK(c, a.user(c))
}

// ForgotPassword handles POST /auth/forgot-password.
func (a *API) ForgotPassword(c *gin.Context) {
	var req forgotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	a.svc.Auth.ForgotPassword(c.Request.Context(), req.Email, c.ClientIP())
	// Always return 200 to avoid leaking which emails are registered.
	response.OK(c, gin.H{"message": "if the account exists, a reset link has been sent"})
}

// ResetPassword handles POST /auth/reset-password.
func (a *API) ResetPassword(c *gin.Context) {
	var req resetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := a.svc.Auth.ResetPassword(c.Request.Context(), req.Token, req.NewPassword, c.ClientIP()); err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "password has been reset"})
}

// ChangePassword handles POST /auth/change-password.
func (a *API) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := a.svc.Auth.ChangePassword(c.Request.Context(), a.user(c).ID, req.OldPassword, req.NewPassword, c.ClientIP()); err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "password changed"})
}

// CreateAPIKey handles POST /auth/api-keys.
func (a *API) CreateAPIKey(c *gin.Context) {
	var req apiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	k, err := a.svc.Auth.CreateAPIKey(c.Request.Context(), a.user(c).ID, req.Name, req.ExpiresAt, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, k)
}

// ListAPIKeys handles GET /auth/api-keys.
func (a *API) ListAPIKeys(c *gin.Context) {
	keys, err := a.svc.Auth.ListAPIKeys(c.Request.Context(), a.user(c).ID)
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, keys)
}

// RevokeAPIKey handles DELETE /auth/api-keys/:id.
func (a *API) RevokeAPIKey(c *gin.Context) {
	if err := a.svc.Auth.RevokeAPIKey(c.Request.Context(), a.user(c).ID, c.Param("id"), c.ClientIP()); err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "api key revoked"})
}

// AcceptInvitation handles POST /invitations/accept.
func (a *API) AcceptInvitation(c *gin.Context) {
	var req acceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	u := a.user(c)
	m, err := a.svc.Invitation.Accept(c.Request.Context(), req.Token, u.ID, u.Email, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, m)
}

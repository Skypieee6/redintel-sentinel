package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

type invitationRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

// CreateInvitation handles POST /orgs/:orgID/invitations.
func (a *API) CreateInvitation(c *gin.Context) {
	var req invitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	inv, err := a.svc.Invitation.Create(c.Request.Context(), a.user(c).ID, c.Param("orgID"), req.Email, models.Role(req.Role), c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, inv)
}

// ListInvitations handles GET /orgs/:orgID/invitations.
func (a *API) ListInvitations(c *gin.Context) {
	invs, err := a.svc.Invitation.List(c.Request.Context(), c.Param("orgID"))
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, invs)
}

// RevokeInvitation handles DELETE /orgs/:orgID/invitations/:inviteID.
func (a *API) RevokeInvitation(c *gin.Context) {
	if err := a.svc.Invitation.Revoke(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("inviteID"), c.ClientIP()); err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "invitation revoked"})
}

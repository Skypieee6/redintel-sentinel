package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

type orgRequest struct {
	Name string `json:"name" binding:"required"`
}

type memberRoleRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

// CreateOrg handles POST /orgs.
func (a *API) CreateOrg(c *gin.Context) {
	var req orgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	org, err := a.svc.Org.Create(c.Request.Context(), a.user(c).ID, req.Name, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, org)
}

// ListOrgs handles GET /orgs.
func (a *API) ListOrgs(c *gin.Context) {
	orgs, err := a.svc.Org.ListForUser(c.Request.Context(), a.user(c).ID)
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, orgs)
}

// GetOrg handles GET /orgs/:orgID.
func (a *API) GetOrg(c *gin.Context) {
	org, err := a.svc.Org.Get(c.Request.Context(), c.Param("orgID"))
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, org)
}

// UpdateOrg handles PUT /orgs/:orgID.
func (a *API) UpdateOrg(c *gin.Context) {
	var req orgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	org, err := a.svc.Org.Update(c.Request.Context(), a.user(c).ID, c.Param("orgID"), req.Name, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, org)
}

// DeleteOrg handles DELETE /orgs/:orgID.
func (a *API) DeleteOrg(c *gin.Context) {
	if err := a.svc.Org.Delete(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.ClientIP()); err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "organization deleted"})
}

// ListMembers handles GET /orgs/:orgID/members.
func (a *API) ListMembers(c *gin.Context) {
	members, err := a.svc.Org.ListMembers(c.Request.Context(), c.Param("orgID"))
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, members)
}

// SetMemberRole handles PUT /orgs/:orgID/members.
func (a *API) SetMemberRole(c *gin.Context) {
	var req memberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	m, err := a.svc.Org.SetMemberRole(c.Request.Context(), a.user(c).ID, c.Param("orgID"), req.Email, models.Role(req.Role), c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, m)
}

// RemoveMember handles DELETE /orgs/:orgID/members/:userID.
func (a *API) RemoveMember(c *gin.Context) {
	if err := a.svc.Org.RemoveMember(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("userID"), c.ClientIP()); err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "member removed"})
}

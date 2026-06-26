package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

type teamRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type teamMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// CreateTeam handles POST /orgs/:orgID/teams.
func (a *API) CreateTeam(c *gin.Context) {
	var req teamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	t, err := a.svc.Team.Create(c.Request.Context(), a.user(c).ID, c.Param("orgID"), req.Name, req.Description, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, t)
}

// ListTeams handles GET /orgs/:orgID/teams.
func (a *API) ListTeams(c *gin.Context) {
	teams, err := a.svc.Team.List(c.Request.Context(), c.Param("orgID"))
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, teams)
}

// GetTeam handles GET /orgs/:orgID/teams/:teamID.
func (a *API) GetTeam(c *gin.Context) {
	t, err := a.svc.Team.Get(c.Request.Context(), c.Param("orgID"), c.Param("teamID"))
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, t)
}

// UpdateTeam handles PUT /orgs/:orgID/teams/:teamID.
func (a *API) UpdateTeam(c *gin.Context) {
	var req teamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	t, err := a.svc.Team.Update(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("teamID"), req.Name, req.Description, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, t)
}

// DeleteTeam handles DELETE /orgs/:orgID/teams/:teamID.
func (a *API) DeleteTeam(c *gin.Context) {
	if err := a.svc.Team.Delete(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("teamID"), c.ClientIP()); err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "team deleted"})
}

// AddTeamMember handles POST /orgs/:orgID/teams/:teamID/members.
func (a *API) AddTeamMember(c *gin.Context) {
	var req teamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := a.svc.Team.AddMember(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("teamID"), req.Email, c.ClientIP()); err != nil {
		a.fail(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, gin.H{"message": "member added"})
}

// RemoveTeamMember handles DELETE /orgs/:orgID/teams/:teamID/members/:userID.
func (a *API) RemoveTeamMember(c *gin.Context) {
	if err := a.svc.Team.RemoveMember(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("teamID"), c.Param("userID"), c.ClientIP()); err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "member removed"})
}

// ListTeamMembers handles GET /orgs/:orgID/teams/:teamID/members.
func (a *API) ListTeamMembers(c *gin.Context) {
	members, err := a.svc.Team.ListMembers(c.Request.Context(), c.Param("orgID"), c.Param("teamID"))
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, members)
}

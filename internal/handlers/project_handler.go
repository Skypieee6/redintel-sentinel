package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

type projectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type projectMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

// CreateProject handles POST /orgs/:orgID/projects.
func (a *API) CreateProject(c *gin.Context) {
	var req projectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	p, err := a.svc.Project.Create(c.Request.Context(), a.user(c).ID, c.Param("orgID"), a.orgRole(c), req.Name, req.Description, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, p)
}

// ListProjects handles GET /orgs/:orgID/projects.
func (a *API) ListProjects(c *gin.Context) {
	projects, err := a.svc.Project.List(c.Request.Context(), c.Param("orgID"))
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, projects)
}

// GetProject handles GET /orgs/:orgID/projects/:projectID.
func (a *API) GetProject(c *gin.Context) {
	p, err := a.svc.Project.Get(c.Request.Context(), c.Param("orgID"), c.Param("projectID"))
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, p)
}

// UpdateProject handles PUT /orgs/:orgID/projects/:projectID.
func (a *API) UpdateProject(c *gin.Context) {
	var req projectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	p, err := a.svc.Project.Update(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("projectID"), a.orgRole(c), req.Name, req.Description, req.Status, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, p)
}

// DeleteProject handles DELETE /orgs/:orgID/projects/:projectID.
func (a *API) DeleteProject(c *gin.Context) {
	if err := a.svc.Project.Delete(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("projectID"), a.orgRole(c), c.ClientIP()); err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "project deleted"})
}

// AddProjectMember handles POST /orgs/:orgID/projects/:projectID/members.
func (a *API) AddProjectMember(c *gin.Context) {
	var req projectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := a.svc.Project.AddMember(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("projectID"), a.orgRole(c), req.Email, models.Role(req.Role), c.ClientIP()); err != nil {
		a.fail(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, gin.H{"message": "member added"})
}

// RemoveProjectMember handles DELETE /orgs/:orgID/projects/:projectID/members/:userID.
func (a *API) RemoveProjectMember(c *gin.Context) {
	if err := a.svc.Project.RemoveMember(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("projectID"), a.orgRole(c), c.Param("userID"), c.ClientIP()); err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "member removed"})
}

// ListProjectMembers handles GET /orgs/:orgID/projects/:projectID/members.
func (a *API) ListProjectMembers(c *gin.Context) {
	members, err := a.svc.Project.ListMembers(c.Request.Context(), c.Param("orgID"), c.Param("projectID"))
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, members)
}

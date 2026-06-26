package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

type startDiscoveryRequest struct {
	InputType  string `json:"input_type" binding:"required"`
	InputValue string `json:"input_value" binding:"required"`
}

// StartDiscovery handles POST /orgs/:orgID/projects/:projectID/discovery.
func (a *API) StartDiscovery(c *gin.Context) {
	var req startDiscoveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	job, err := a.svc.Discovery.Start(
		c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("projectID"), a.orgRole(c),
		models.AssetType(req.InputType), req.InputValue, c.ClientIP(),
	)
	if err != nil {
		a.fail(c, err)
		return
	}
	response.JSON(c, http.StatusAccepted, job)
}

// ListDiscoveryJobs handles GET /orgs/:orgID/projects/:projectID/discovery with
// pagination via limit and offset query params.
func (a *API) ListDiscoveryJobs(c *gin.Context) {
	limit, offset := parseLimitOffset(c)
	page, err := a.svc.Discovery.ListJobs(c.Request.Context(), c.Param("orgID"), c.Param("projectID"), limit, offset)
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, page)
}

// GetDiscoveryJob handles GET /orgs/:orgID/projects/:projectID/discovery/:jobID
// and includes the job's findings.
func (a *API) GetDiscoveryJob(c *gin.Context) {
	job, err := a.svc.Discovery.Get(c.Request.Context(), c.Param("orgID"), c.Param("projectID"), c.Param("jobID"))
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, job)
}

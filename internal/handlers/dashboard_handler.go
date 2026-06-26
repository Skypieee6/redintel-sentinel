package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/internal/service"
	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

// Dashboard handles GET /orgs/:orgID/dashboard.
func (a *API) Dashboard(c *gin.Context) {
	summary, err := a.svc.Dashboard.Summary(c.Request.Context(), c.Param("orgID"))
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, summary)
}

// ArchiveProject handles POST /orgs/:orgID/projects/:projectID/archive.
func (a *API) ArchiveProject(c *gin.Context) {
	p, err := a.svc.Project.SetArchived(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("projectID"), a.orgRole(c), true, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, p)
}

// UnarchiveProject handles POST /orgs/:orgID/projects/:projectID/unarchive.
func (a *API) UnarchiveProject(c *gin.Context) {
	p, err := a.svc.Project.SetArchived(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("projectID"), a.orgRole(c), false, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, p)
}

// GenerateReport handles GET /orgs/:orgID/projects/:projectID/report?format=json|csv|markdown|html.
func (a *API) GenerateReport(c *gin.Context) {
	format := service.ReportFormat(c.DefaultQuery("format", "json"))
	report, err := a.svc.Report.Generate(c.Request.Context(), c.Param("orgID"), c.Param("projectID"), format)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+report.Filename)
	c.Data(http.StatusOK, report.ContentType, report.Body)
}

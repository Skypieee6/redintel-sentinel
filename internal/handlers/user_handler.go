package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

type updateProfileRequest struct {
	FullName string `json:"full_name" binding:"required"`
}

func parseLimitOffset(c *gin.Context) (int, int) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// UpdateMe handles PUT /auth/me.
func (a *API) UpdateMe(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	u, err := a.svc.User.UpdateProfile(c.Request.Context(), a.user(c).ID, req.FullName)
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, u)
}

// ListUsers handles GET /admin/users (superadmin only).
func (a *API) ListUsers(c *gin.Context) {
	limit, offset := parseLimitOffset(c)
	users, err := a.svc.User.List(c.Request.Context(), limit, offset)
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, users)
}

// ListOrgAuditLogs handles GET /orgs/:orgID/audit-logs.
func (a *API) ListOrgAuditLogs(c *gin.Context) {
	limit, offset := parseLimitOffset(c)
	logs, err := a.svc.Audit.List(c.Request.Context(), c.Param("orgID"), c.Query("action"), limit, offset)
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, logs)
}

// ListAllAuditLogs handles GET /admin/audit-logs (superadmin only).
func (a *API) ListAllAuditLogs(c *gin.Context) {
	limit, offset := parseLimitOffset(c)
	logs, err := a.svc.Audit.List(c.Request.Context(), c.Query("org_id"), c.Query("action"), limit, offset)
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, logs)
}

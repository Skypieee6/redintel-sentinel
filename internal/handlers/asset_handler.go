package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

type assetRequest struct {
	Type       string         `json:"type" binding:"required"`
	Value      string         `json:"value" binding:"required"`
	Tags       []string       `json:"tags"`
	Attributes map[string]any `json:"attributes"`
	Status     string         `json:"status"`
}

// CreateAsset handles POST /orgs/:orgID/projects/:projectID/assets.
func (a *API) CreateAsset(c *gin.Context) {
	var req assetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	asset := &models.Asset{
		Type:       models.AssetType(req.Type),
		Value:      req.Value,
		Tags:       req.Tags,
		Attributes: req.Attributes,
	}
	created, err := a.svc.Asset.Create(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("projectID"), a.orgRole(c), asset, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, created)
}

// ListAssets handles GET /orgs/:orgID/projects/:projectID/assets with filtering,
// search and pagination via query params: type, q, tag, status, limit, offset.
func (a *API) ListAssets(c *gin.Context) {
	limit, offset := parseLimitOffset(c)
	page, err := a.svc.Asset.List(c.Request.Context(), repository.AssetFilter{
		OrgID:     c.Param("orgID"),
		ProjectID: c.Param("projectID"),
		Type:      c.Query("type"),
		Query:     c.Query("q"),
		Tag:       c.Query("tag"),
		Status:    c.Query("status"),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, page)
}

// GetAsset handles GET /orgs/:orgID/projects/:projectID/assets/:assetID.
func (a *API) GetAsset(c *gin.Context) {
	asset, err := a.svc.Asset.Get(c.Request.Context(), c.Param("projectID"), c.Param("assetID"))
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, asset)
}

// UpdateAsset handles PUT /orgs/:orgID/projects/:projectID/assets/:assetID.
func (a *API) UpdateAsset(c *gin.Context) {
	var req assetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	updated, err := a.svc.Asset.Update(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("projectID"), c.Param("assetID"), a.orgRole(c), req.Value, req.Tags, req.Attributes, req.Status, c.ClientIP())
	if err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, updated)
}

// DeleteAsset handles DELETE /orgs/:orgID/projects/:projectID/assets/:assetID.
func (a *API) DeleteAsset(c *gin.Context) {
	if err := a.svc.Asset.Delete(c.Request.Context(), a.user(c).ID, c.Param("orgID"), c.Param("projectID"), c.Param("assetID"), a.orgRole(c), c.ClientIP()); err != nil {
		a.fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "asset deleted"})
}

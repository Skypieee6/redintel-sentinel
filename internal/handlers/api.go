package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/internal/middleware"
	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/service"
	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

// API bundles the HTTP handlers backed by the service layer.
type API struct {
	svc *service.Services
}

// NewAPI constructs the API handler set.
func NewAPI(svc *service.Services) *API { return &API{svc: svc} }

// fail maps a service error to an HTTP response.
func (a *API) fail(c *gin.Context, err error) {
	var status int
	code := "error"
	switch {
	case errors.Is(err, service.ErrValidation):
		status, code = http.StatusBadRequest, "validation_error"
	case errors.Is(err, service.ErrInvalidCredentials):
		status, code = http.StatusUnauthorized, "invalid_credentials"
	case errors.Is(err, service.ErrInactive):
		status, code = http.StatusForbidden, "inactive"
	case errors.Is(err, service.ErrForbidden):
		status, code = http.StatusForbidden, "forbidden"
	case errors.Is(err, service.ErrNotFound):
		status, code = http.StatusNotFound, "not_found"
	case errors.Is(err, service.ErrConflict):
		status, code = http.StatusConflict, "conflict"
	case errors.Is(err, service.ErrRateLimited):
		status, code = http.StatusTooManyRequests, "rate_limited"
	default:
		response.Error(c, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
		return
	}
	response.Error(c, status, code, err.Error())
}

// user returns the authenticated user; it never returns nil after the auth
// middleware has run.
func (a *API) user(c *gin.Context) *models.User {
	u, _ := middleware.CurrentUser(c)
	return u
}

func (a *API) orgRole(c *gin.Context) models.Role { return middleware.CurrentOrgRole(c) }

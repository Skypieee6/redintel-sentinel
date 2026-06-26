package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/internal/version"
	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

// VersionHandler serves build/version metadata.
type VersionHandler struct{}

// NewVersionHandler constructs a VersionHandler.
func NewVersionHandler() *VersionHandler { return &VersionHandler{} }

// Version returns build metadata for the running binary.
func (h *VersionHandler) Version(c *gin.Context) {
	response.OK(c, version.Get())
}

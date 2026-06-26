// Package handlers implements HTTP request handlers.
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/internal/cache"
	"github.com/Skypieee6/redintel-sentinel/internal/database"
	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

// HealthHandler serves liveness and readiness probes.
type HealthHandler struct {
	db    *database.Postgres
	redis *cache.Redis
}

// NewHealthHandler constructs a HealthHandler. db and redis may be nil; in that
// case the corresponding dependency is reported as "unconfigured".
func NewHealthHandler(db *database.Postgres, redis *cache.Redis) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// Health is a liveness probe. It returns 200 as long as the process is running
// and able to serve traffic; it does not check downstream dependencies.
func (h *HealthHandler) Health(c *gin.Context) {
	response.JSON(c, http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready is a readiness probe. It verifies that critical downstream dependencies
// (PostgreSQL, Redis) are reachable and returns 503 if any are not.
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	checks := map[string]string{}
	ready := true

	switch {
	case h.db == nil:
		checks["postgres"] = "unconfigured"
	case h.db.Ping(ctx) != nil:
		checks["postgres"] = "down"
		ready = false
	default:
		checks["postgres"] = "ok"
	}

	switch {
	case h.redis == nil:
		checks["redis"] = "unconfigured"
	case h.redis.Ping(ctx) != nil:
		checks["redis"] = "down"
		ready = false
	default:
		checks["redis"] = "ok"
	}

	status := http.StatusOK
	overall := "ready"
	if !ready {
		status = http.StatusServiceUnavailable
		overall = "not_ready"
	}

	response.JSON(c, status, gin.H{
		"status": overall,
		"checks": checks,
	})
}

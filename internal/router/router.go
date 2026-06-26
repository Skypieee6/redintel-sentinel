// Package router wires HTTP routes onto a Gin engine.
package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Skypieee6/redintel-sentinel/internal/cache"
	"github.com/Skypieee6/redintel-sentinel/internal/config"
	"github.com/Skypieee6/redintel-sentinel/internal/database"
	"github.com/Skypieee6/redintel-sentinel/internal/handlers"
	"github.com/Skypieee6/redintel-sentinel/internal/middleware"
)

// Dependencies bundles everything the router needs to build the engine.
type Dependencies struct {
	Config *config.Config
	Logger *zap.Logger
	DB     *database.Postgres
	Redis  *cache.Redis
}

// New constructs a fully configured Gin engine.
func New(deps Dependencies) *gin.Engine {
	if deps.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.RedirectTrailingSlash = false

	// Global middleware, ordered: correlation ID -> logging -> recovery.
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Logger(deps.Logger))
	engine.Use(middleware.Recovery(deps.Logger))

	health := handlers.NewHealthHandler(deps.DB, deps.Redis)
	ver := handlers.NewVersionHandler()

	// Operational / probe endpoints (root level for orchestrators like k8s).
	engine.GET("/health", health.Health)
	engine.GET("/healthz", health.Health)
	engine.GET("/ready", health.Ready)
	engine.GET("/readyz", health.Ready)
	engine.GET("/version", ver.Version)

	// Versioned API surface. Feature modules (assets, projects, scans, ...)
	// will register their routes under this group in later phases.
	v1 := engine.Group("/api/v1")
	{
		v1.GET("/health", health.Health)
		v1.GET("/ready", health.Ready)
		v1.GET("/version", ver.Version)
	}

	return engine
}

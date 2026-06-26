// Package router wires HTTP routes onto a Gin engine.
package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Skypieee6/redintel-sentinel/internal/auth"
	"github.com/Skypieee6/redintel-sentinel/internal/cache"
	"github.com/Skypieee6/redintel-sentinel/internal/config"
	"github.com/Skypieee6/redintel-sentinel/internal/database"
	"github.com/Skypieee6/redintel-sentinel/internal/handlers"
	"github.com/Skypieee6/redintel-sentinel/internal/middleware"
	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
	"github.com/Skypieee6/redintel-sentinel/internal/service"
)

// Dependencies bundles everything the router needs to build the engine.
type Dependencies struct {
	Config   *config.Config
	Logger   *zap.Logger
	DB       *database.Postgres
	Redis    *cache.Redis
	Repos    *repository.Repositories
	Services *service.Services
	JWT      *auth.JWTManager
}

// New constructs a fully configured Gin engine.
func New(deps Dependencies) *gin.Engine {
	if deps.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.RedirectTrailingSlash = false
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Logger(deps.Logger))
	engine.Use(middleware.Recovery(deps.Logger))

	health := handlers.NewHealthHandler(deps.DB, deps.Redis)
	ver := handlers.NewVersionHandler()
	api := handlers.NewAPI(deps.Services)
	authn := middleware.Authenticate(deps.JWT, deps.Repos)

	// Probes.
	engine.GET("/health", health.Health)
	engine.GET("/healthz", health.Health)
	engine.GET("/ready", health.Ready)
	engine.GET("/readyz", health.Ready)
	engine.GET("/version", ver.Version)

	// API docs.
	engine.GET("/docs", handlers.SwaggerUI)
	engine.StaticFile("/api/v1/openapi.yaml", "configs/openapi.yaml")

	v1 := engine.Group("/api/v1")
	v1.GET("/health", health.Health)
	v1.GET("/version", ver.Version)

	registerAuthRoutes(v1, api, authn)
	registerOrgRoutes(v1, api, authn, deps.Services.Org)
	registerAdminRoutes(v1, api, authn)

	return engine
}

func registerAuthRoutes(v1 *gin.RouterGroup, api *handlers.API, authn gin.HandlerFunc) {
	a := v1.Group("/auth")
	a.POST("/register", api.Register)
	a.POST("/login", api.Login)
	a.POST("/refresh", api.Refresh)
	a.POST("/forgot-password", api.ForgotPassword)
	a.POST("/reset-password", api.ResetPassword)

	sec := a.Group("")
	sec.Use(authn)
	sec.GET("/me", api.Me)
	sec.PUT("/me", api.UpdateMe)
	sec.POST("/logout", api.Logout)
	sec.POST("/change-password", api.ChangePassword)
	sec.GET("/api-keys", api.ListAPIKeys)
	sec.POST("/api-keys", api.CreateAPIKey)
	sec.DELETE("/api-keys/:id", api.RevokeAPIKey)

	inv := v1.Group("/invitations")
	inv.Use(authn)
	inv.POST("/accept", api.AcceptInvitation)
}

func registerOrgRoutes(v1 *gin.RouterGroup, api *handlers.API, authn gin.HandlerFunc, orgs *service.OrgService) {
	orgsGroup := v1.Group("/orgs")
	orgsGroup.Use(authn)
	orgsGroup.GET("", api.ListOrgs)
	orgsGroup.POST("", api.CreateOrg)

	// Scoped to a single org; OrgContext loads the caller's role.
	o := orgsGroup.Group("/:orgID")
	o.Use(middleware.OrgContext(orgs))

	o.GET("", middleware.RequireOrgRole(models.RoleViewer), api.GetOrg)
	o.PUT("", middleware.RequireOrgRole(models.RoleAdmin), api.UpdateOrg)
	o.DELETE("", middleware.RequireOrgRole(models.RoleAdmin), api.DeleteOrg)

	o.GET("/members", middleware.RequireOrgRole(models.RoleViewer), api.ListMembers)
	o.PUT("/members", middleware.RequireOrgRole(models.RoleAdmin), api.SetMemberRole)
	o.DELETE("/members/:userID", middleware.RequireOrgRole(models.RoleAdmin), api.RemoveMember)

	o.GET("/teams", middleware.RequireOrgRole(models.RoleViewer), api.ListTeams)
	o.POST("/teams", middleware.RequireOrgRole(models.RoleManager), api.CreateTeam)
	o.GET("/teams/:teamID", middleware.RequireOrgRole(models.RoleViewer), api.GetTeam)
	o.PUT("/teams/:teamID", middleware.RequireOrgRole(models.RoleManager), api.UpdateTeam)
	o.DELETE("/teams/:teamID", middleware.RequireOrgRole(models.RoleManager), api.DeleteTeam)
	o.GET("/teams/:teamID/members", middleware.RequireOrgRole(models.RoleViewer), api.ListTeamMembers)
	o.POST("/teams/:teamID/members", middleware.RequireOrgRole(models.RoleManager), api.AddTeamMember)
	o.DELETE("/teams/:teamID/members/:userID", middleware.RequireOrgRole(models.RoleManager), api.RemoveTeamMember)

	o.GET("/invitations", middleware.RequireOrgRole(models.RoleManager), api.ListInvitations)
	o.POST("/invitations", middleware.RequireOrgRole(models.RoleAdmin), api.CreateInvitation)
	o.DELETE("/invitations/:inviteID", middleware.RequireOrgRole(models.RoleAdmin), api.RevokeInvitation)

	o.GET("/projects", middleware.RequireOrgRole(models.RoleViewer), api.ListProjects)
	o.POST("/projects", middleware.RequireOrgRole(models.RoleManager), api.CreateProject)
	o.GET("/projects/:projectID", middleware.RequireOrgRole(models.RoleViewer), api.GetProject)
	o.PUT("/projects/:projectID", middleware.RequireOrgRole(models.RoleAnalyst), api.UpdateProject)
	o.DELETE("/projects/:projectID", middleware.RequireOrgRole(models.RoleAnalyst), api.DeleteProject)
	o.GET("/projects/:projectID/members", middleware.RequireOrgRole(models.RoleViewer), api.ListProjectMembers)
	o.POST("/projects/:projectID/members", middleware.RequireOrgRole(models.RoleAnalyst), api.AddProjectMember)
	o.DELETE("/projects/:projectID/members/:userID", middleware.RequireOrgRole(models.RoleAnalyst), api.RemoveProjectMember)

	o.POST("/projects/:projectID/archive", middleware.RequireOrgRole(models.RoleAnalyst), api.ArchiveProject)
	o.POST("/projects/:projectID/unarchive", middleware.RequireOrgRole(models.RoleAnalyst), api.UnarchiveProject)
	o.GET("/projects/:projectID/report", middleware.RequireOrgRole(models.RoleViewer), api.GenerateReport)

	// Asset inventory (Phase 5).
	o.GET("/projects/:projectID/assets", middleware.RequireOrgRole(models.RoleViewer), api.ListAssets)
	o.POST("/projects/:projectID/assets", middleware.RequireOrgRole(models.RoleAnalyst), api.CreateAsset)
	o.GET("/projects/:projectID/assets/:assetID", middleware.RequireOrgRole(models.RoleViewer), api.GetAsset)
	o.PUT("/projects/:projectID/assets/:assetID", middleware.RequireOrgRole(models.RoleAnalyst), api.UpdateAsset)
	o.DELETE("/projects/:projectID/assets/:assetID", middleware.RequireOrgRole(models.RoleAnalyst), api.DeleteAsset)

	// Passive asset discovery (Phase 1).
	o.GET("/projects/:projectID/discovery", middleware.RequireOrgRole(models.RoleViewer), api.ListDiscoveryJobs)
	o.POST("/projects/:projectID/discovery", middleware.RequireOrgRole(models.RoleAnalyst), api.StartDiscovery)
	o.GET("/projects/:projectID/discovery/:jobID", middleware.RequireOrgRole(models.RoleViewer), api.GetDiscoveryJob)

	// Dashboard (Phase 6).
	o.GET("/dashboard", middleware.RequireOrgRole(models.RoleViewer), api.Dashboard)

	o.GET("/audit-logs", middleware.RequireOrgRole(models.RoleManager), api.ListOrgAuditLogs)
}

func registerAdminRoutes(v1 *gin.RouterGroup, api *handlers.API, authn gin.HandlerFunc) {
	admin := v1.Group("/admin")
	admin.Use(authn, middleware.RequireSuperadmin())
	admin.GET("/users", api.ListUsers)
	admin.GET("/audit-logs", api.ListAllAuditLogs)
}

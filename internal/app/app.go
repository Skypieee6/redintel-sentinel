// Package app wires together configuration, logging, datastores, services and
// the HTTP server into a runnable application.
package app

import (
	"context"

	"go.uber.org/zap"

	"github.com/Skypieee6/redintel-sentinel/internal/auth"
	"github.com/Skypieee6/redintel-sentinel/internal/cache"
	"github.com/Skypieee6/redintel-sentinel/internal/config"
	"github.com/Skypieee6/redintel-sentinel/internal/database"
	"github.com/Skypieee6/redintel-sentinel/internal/logger"
	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
	"github.com/Skypieee6/redintel-sentinel/internal/router"
	"github.com/Skypieee6/redintel-sentinel/internal/server"
	"github.com/Skypieee6/redintel-sentinel/internal/service"
	"github.com/Skypieee6/redintel-sentinel/internal/version"
)

// App holds long-lived application dependencies.
type App struct {
	cfg    *config.Config
	log    *zap.Logger
	db     *database.Postgres
	redis  *cache.Redis
	server *server.Server
}

// New builds the application: it loads config, initializes the logger, connects
// to PostgreSQL and Redis, wires the service layer and constructs the HTTP
// server.
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	log, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		return nil, err
	}

	log.Info("initializing application",
		zap.String("name", cfg.App.Name),
		zap.String("environment", cfg.App.Environment),
		zap.String("version", version.Get().Version),
	)

	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		log.Error("failed to connect to postgres", zap.Error(err))
		return nil, err
	}
	log.Info("connected to postgres", zap.String("host", cfg.Database.Host))

	redis, err := cache.New(ctx, cfg.Redis)
	if err != nil {
		log.Error("failed to connect to redis", zap.Error(err))
		db.Close()
		return nil, err
	}
	log.Info("connected to redis", zap.String("addr", cfg.Redis.Addr()))

	repos := repository.New(db.Pool)
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenTTL)
	services := service.New(repos, jwtManager, redis, cfg.Auth, log)

	if err := seedAdmin(ctx, repos, cfg.Auth, log); err != nil {
		log.Warn("admin seeding failed", zap.Error(err))
	}

	engine := router.New(router.Dependencies{
		Config:   cfg,
		Logger:   log,
		DB:       db,
		Redis:    redis,
		Repos:    repos,
		Services: services,
		JWT:      jwtManager,
	})

	srv := server.New(cfg.Server, engine, log)

	return &App{cfg: cfg, log: log, db: db, redis: redis, server: srv}, nil
}

// seedAdmin ensures a platform superadmin exists, creating or repairing it.
func seedAdmin(ctx context.Context, repos *repository.Repositories, cfg config.AuthConfig, log *zap.Logger) error {
	if cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		return nil
	}
	existing, err := repos.Users.GetByEmail(ctx, cfg.AdminEmail)
	if err == nil && existing != nil {
		if !existing.IsSuperadmin {
			log.Info("existing admin user is not superadmin; leaving as-is", zap.String("email", cfg.AdminEmail))
		}
		return nil
	}
	hash, err := auth.HashPassword(cfg.AdminPassword, cfg.BcryptCost)
	if err != nil {
		return err
	}
	if _, err := repos.Users.Create(ctx, &models.User{
		Email:        cfg.AdminEmail,
		PasswordHash: hash,
		FullName:     "Platform Administrator",
		IsActive:     true,
		IsSuperadmin: true,
	}); err != nil {
		if err == repository.ErrConflict {
			return nil
		}
		return err
	}
	log.Info("seeded platform superadmin", zap.String("email", cfg.AdminEmail))
	return nil
}

// Run starts the HTTP server and blocks until ctx is canceled.
func (a *App) Run(ctx context.Context) error { return a.server.Run(ctx) }

// Close releases all held resources.
func (a *App) Close() {
	if a.redis != nil {
		_ = a.redis.Close()
	}
	if a.db != nil {
		a.db.Close()
	}
	if a.log != nil {
		_ = a.log.Sync()
	}
}

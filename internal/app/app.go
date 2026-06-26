// Package app wires together configuration, logging, datastores and the HTTP
// server into a runnable application.
package app

import (
	"context"

	"go.uber.org/zap"

	"github.com/Skypieee6/redintel-sentinel/internal/cache"
	"github.com/Skypieee6/redintel-sentinel/internal/config"
	"github.com/Skypieee6/redintel-sentinel/internal/database"
	"github.com/Skypieee6/redintel-sentinel/internal/logger"
	"github.com/Skypieee6/redintel-sentinel/internal/router"
	"github.com/Skypieee6/redintel-sentinel/internal/server"
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
// to PostgreSQL and Redis, and constructs the HTTP server.
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

	engine := router.New(router.Dependencies{
		Config: cfg,
		Logger: log,
		DB:     db,
		Redis:  redis,
	})

	srv := server.New(cfg.Server, engine, log)

	return &App{
		cfg:    cfg,
		log:    log,
		db:     db,
		redis:  redis,
		server: srv,
	}, nil
}

// Run starts the HTTP server and blocks until ctx is canceled.
func (a *App) Run(ctx context.Context) error {
	return a.server.Run(ctx)
}

// Close releases all held resources. Safe to call multiple times.
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

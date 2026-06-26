// Package server runs the HTTP server with graceful shutdown.
package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/Skypieee6/redintel-sentinel/internal/config"
)

// Server wraps an *http.Server with lifecycle management.
type Server struct {
	httpServer      *http.Server
	log             *zap.Logger
	shutdownTimeout time.Duration
}

// New creates a Server bound according to the supplied configuration.
func New(cfg config.ServerConfig, handler http.Handler, log *zap.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Addr(),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		log:             log,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

// Run starts the server and blocks until ctx is canceled, at which point it
// performs a graceful shutdown bounded by the configured shutdown timeout.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.log.Info("http server listening", zap.String("addr", s.httpServer.Addr))
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.log.Info("shutdown signal received, draining connections",
			zap.Duration("timeout", s.shutdownTimeout))

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.log.Error("graceful shutdown failed, forcing close", zap.Error(err))
			_ = s.httpServer.Close()
			return err
		}
		s.log.Info("http server stopped cleanly")
		return nil
	}
}

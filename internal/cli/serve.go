package cli

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/Skypieee6/redintel-sentinel/internal/app"
	"github.com/Skypieee6/redintel-sentinel/internal/config"
)

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP API server",
		Long:  "Start the RedIntel Sentinel HTTP API server with graceful shutdown.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Trap SIGINT/SIGTERM and cancel the context to trigger a graceful
			// shutdown of the HTTP server.
			ctx, stop := signal.NotifyContext(context.Background(),
				syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			application, err := app.New(ctx, cfg)
			if err != nil {
				return err
			}
			defer application.Close()

			return application.Run(ctx)
		},
	}
}

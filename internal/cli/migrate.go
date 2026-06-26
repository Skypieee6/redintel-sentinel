package cli

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"

	"github.com/Skypieee6/redintel-sentinel/internal/config"
)

func newMigrateCmd() *cobra.Command {
	var migrationsPath string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Manage database schema migrations",
		Long:  "Apply or roll back PostgreSQL schema migrations using golang-migrate.",
	}
	cmd.PersistentFlags().StringVar(&migrationsPath, "path", "migrations",
		"filesystem path containing migration files")

	newMigrator := func() (*migrate.Migrate, error) {
		cfg, err := config.Load()
		if err != nil {
			return nil, err
		}
		// golang-migrate expects the pgx5 scheme.
		dsn := "pgx5://" + cfg.Database.DSN()[len("postgres://"):]
		m, err := migrate.New("file://"+migrationsPath, dsn)
		if err != nil {
			return nil, fmt.Errorf("init migrator: %w", err)
		}
		return m, nil
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "Apply all pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := newMigrator()
			if err != nil {
				return err
			}
			defer m.Close()
			if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}
			fmt.Println("migrations applied")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "down",
		Short: "Roll back the most recent migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := newMigrator()
			if err != nil {
				return err
			}
			defer m.Close()
			if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}
			fmt.Println("rolled back one migration")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the current migration version",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := newMigrator()
			if err != nil {
				return err
			}
			defer m.Close()
			v, dirty, err := m.Version()
			if err != nil {
				return err
			}
			fmt.Printf("version=%d dirty=%t\n", v, dirty)
			return nil
		},
	})

	return cmd
}

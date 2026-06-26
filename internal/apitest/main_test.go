package apitest

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/Skypieee6/redintel-sentinel/internal/config"
)

// TestMain applies database migrations before the integration suite runs so the
// schema exists in any environment (including CI, which provisions a fresh
// Postgres without a separate migrate step). It is best-effort: if the database
// is unreachable the individual tests skip themselves via setup().
func TestMain(m *testing.M) {
	applyMigrations()
	os.Exit(m.Run())
}

func applyMigrations() {
	cfg, err := config.Load()
	if err != nil {
		return
	}
	dsn := "pgx5://" + cfg.Database.DSN()[len("postgres://"):]

	_, thisFile, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations")

	mg, err := migrate.New("file://"+migrationsDir, dsn)
	if err != nil {
		return
	}
	defer mg.Close()
	if err := mg.Up(); err != nil && err != migrate.ErrNoChange {
		return
	}
}

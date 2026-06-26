package config

import "testing"

func TestServerAddr(t *testing.T) {
	s := ServerConfig{Host: "0.0.0.0", Port: 8080}
	if got := s.Addr(); got != "0.0.0.0:8080" {
		t.Fatalf("Addr() = %q, want %q", got, "0.0.0.0:8080")
	}
}

func TestDatabaseDSN(t *testing.T) {
	d := DatabaseConfig{
		Host: "db", Port: 5432, User: "postgres", Password: "secret",
		Name: "redintel", SSLMode: "disable",
	}
	want := "postgres://postgres:secret@db:5432/redintel?sslmode=disable"
	if got := d.DSN(); got != want {
		t.Fatalf("DSN() = %q, want %q", got, want)
	}
}

func TestRedisAddr(t *testing.T) {
	r := RedisConfig{Host: "redis", Port: 6379}
	if got := r.Addr(); got != "redis:6379" {
		t.Fatalf("Addr() = %q, want %q", got, "redis:6379")
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("REDINTEL_DATABASE_HOST", "localhost")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("default server port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.App.Name == "" {
		t.Error("app name should not be empty")
	}
	if cfg.IsProduction() {
		t.Error("default environment should not be production")
	}
}

func TestValidateRejectsBadPort(t *testing.T) {
	c := &Config{}
	c.Server.Port = 70000
	c.Database.Host = "h"
	c.Database.Name = "n"
	c.Log.Level = "info"
	if err := c.validate(); err == nil {
		t.Fatal("expected validation error for out-of-range port")
	}
}

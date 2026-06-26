// Package config loads and validates application configuration.
//
// Configuration is layered, with later sources overriding earlier ones:
//  1. Hard-coded defaults (defined in this package).
//  2. An optional YAML config file (config.yaml / configs/config.yaml).
//  3. Environment variables, prefixed with REDINTEL_ (e.g. REDINTEL_SERVER_PORT).
//
// A .env file in the working directory is loaded automatically (best-effort)
// to ease local development.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config is the root application configuration.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Log      LogConfig      `mapstructure:"log"`
}

// AuthConfig holds authentication and token settings.
type AuthConfig struct {
	JWTSecret        string        `mapstructure:"jwt_secret"`
	AccessTokenTTL   time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL  time.Duration `mapstructure:"refresh_token_ttl"`
	PasswordResetTTL time.Duration `mapstructure:"password_reset_ttl"`
	InvitationTTL    time.Duration `mapstructure:"invitation_ttl"`
	BcryptCost       int           `mapstructure:"bcrypt_cost"`
	AdminEmail       string        `mapstructure:"admin_email"`
	AdminPassword    string        `mapstructure:"admin_password"`
}

// AppConfig holds top-level application metadata.
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// Addr returns the host:port the server should bind to.
func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxConns        int32         `mapstructure:"max_conns"`
	MinConns        int32         `mapstructure:"min_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout"`
}

// DSN returns a PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Addr returns the host:port for the Redis client.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// LogConfig holds logging settings.
type LogConfig struct {
	// Level is one of: debug, info, warn, error.
	Level string `mapstructure:"level"`
	// Format is one of: json, console.
	Format string `mapstructure:"format"`
}

// Load reads configuration from defaults, an optional config file and the
// environment, then validates the result.
func Load() (*Config, error) {
	// Best-effort load of a local .env file; ignored if not present.
	_ = godotenv.Load()

	v := viper.New()
	setDefaults(v)

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")
	v.AddConfigPath("/etc/redintel")

	v.SetEnvPrefix("REDINTEL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		// A missing config file is acceptable; anything else is fatal.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// IsProduction reports whether the app runs in a production-like environment.
func (c *Config) IsProduction() bool {
	env := strings.ToLower(c.App.Environment)
	return env == "production" || env == "prod"
}

func (c *Config) validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port %d out of range", c.Server.Port)
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database.name is required")
	}
	switch strings.ToLower(c.Log.Level) {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("log.level %q is invalid", c.Log.Level)
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required")
	}
	if c.IsProduction() && strings.HasPrefix(c.Auth.JWTSecret, "dev-insecure") {
		return fmt.Errorf("auth.jwt_secret must be set to a strong value in production")
	}
	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "redintel-sentinel")
	v.SetDefault("app.environment", "development")

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "15s")
	v.SetDefault("server.write_timeout", "15s")
	v.SetDefault("server.idle_timeout", "60s")
	v.SetDefault("server.shutdown_timeout", "15s")

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "redintel")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_conns", 10)
	v.SetDefault("database.min_conns", 2)
	v.SetDefault("database.max_conn_lifetime", "1h")
	v.SetDefault("database.connect_timeout", "5s")

	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	v.SetDefault("auth.jwt_secret", "dev-insecure-change-me-please-0123456789abcdef")
	v.SetDefault("auth.access_token_ttl", "15m")
	v.SetDefault("auth.refresh_token_ttl", "168h")
	v.SetDefault("auth.password_reset_ttl", "1h")
	v.SetDefault("auth.invitation_ttl", "168h")
	v.SetDefault("auth.bcrypt_cost", 12)
	v.SetDefault("auth.admin_email", "admin@redintel.local")
	v.SetDefault("auth.admin_password", "ChangeMe123!")
}

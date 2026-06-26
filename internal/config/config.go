package config

import "os"

// Config stores application configuration.
type Config struct {
	Port string
	Environment string
}

// Load returns configuration from environment variables with sensible defaults.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	return Config{
		Port: port,
		Environment: env,
	}
}

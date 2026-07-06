package config

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/caarlos0/env/v11"
)

// Environment enumerates the supported runtime environments. Parsing external
// input into a small enum keeps the "development"/"production" strings out of
// call sites and lets the type system reject anything else at the boundary.
type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

// Config holds environment-driven settings (Twelve-Factor). Add the fields each
// entry point needs — e.g. CLI/agent projects can drop Port, and ADK agents add
// GOOGLE_CLOUD_* fields for Vertex AI.
type Config struct {
	Environment Environment `env:"ENVIRONMENT" envDefault:"development"`
	Port        int         `env:"PORT" envDefault:"8080"` // web server listen port
}

// Load parses configuration from the environment and validates it, failing fast
// so misconfiguration surfaces at startup rather than mid-request.
func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	switch cfg.Environment {
	case Development, Production:
	default:
		return Config{}, fmt.Errorf("invalid ENVIRONMENT %q (want %q or %q)", cfg.Environment, Development, Production)
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return Config{}, fmt.Errorf("invalid PORT %d (want 1-65535)", cfg.Port)
	}
	return cfg, nil
}

// NewHandler builds the slog handler for the environment: human-readable text at
// debug level in development, structured JSON at info level in production.
func (c Config) NewHandler(w io.Writer) slog.Handler {
	if c.Environment == Production {
		return slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	return slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug})
}

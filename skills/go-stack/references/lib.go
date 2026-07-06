package <slug>

import (
	"context"
	"fmt"
	"log/slog"
)

// Config represents the library configuration.
type Config struct {
	ConfigPath string `json:"config_path,omitzero"`
}

// Client is the entry point for the library business logic.
type Client struct {
	logger *slog.Logger
}

// NewClient initializes a new Client.
func NewClient(logger *slog.Logger) *Client {
	return &Client{logger: logger}
}

// DoSomething executes the library business logic.
func (c *Client) DoSomething(ctx context.Context, cfg Config) error {
	c.logger.InfoContext(ctx, "doing something in library", "configPath", cfg.ConfigPath)
	if cfg.ConfigPath != "" {
		return fmt.Errorf("config files are not implemented yet: %s", cfg.ConfigPath)
	}
	return nil
}

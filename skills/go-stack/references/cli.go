package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	"<import_path>"
	"<import_path>/config"
)

func main() {
	// Load configuration and initialize structured logging (fail fast on error).
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load configuration", "error", err)
		os.Exit(1)
	}
	logger := slog.New(cfg.NewHandler(os.Stderr))
	slog.SetDefault(logger)

	cmd := &cli.Command{
		Name:    "<slug>",
		Usage:   "A modern Go CLI application",
		Version: "0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "path to configuration file",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			logger.Info("Starting <slug>", "version", cmd.Version)

			libCfg := <slug>.Config{
				ConfigPath: cmd.String("config"),
			}

			client := <slug>.NewClient(logger)
			return client.DoSomething(ctx, libCfg)
		},
	}

	// Cancel the context on Ctrl+C / SIGTERM so long-running actions can stop cleanly.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cmd.Run(ctx, os.Args); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}
}

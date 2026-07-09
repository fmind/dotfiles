// Package dot implements the unified CLI utility to manage dotfiles, local clusters, and workspaces.
package dot

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

// GlobalState wraps dependencies shared between subcommands.
type GlobalState struct {
	Config     *Config
	Logger     *slog.Logger
	Runner     Runner
	Browser    Browser
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	ConfigPath string // resolved path of the config file (loaded or default location)
}

// NewApp constructs the main urfave/cli/v3 command for dot.
func NewApp() *cli.Command {
	state := &GlobalState{
		Config:  DefaultConfig(), // Start with default config, overwrite in Before
		Logger:  slog.Default(),
		Browser: OSBrowser{},
		Stdin:   os.Stdin,
		Stdout:  colorWriter(os.Stdout),
		Stderr:  colorWriter(os.Stderr),
	}
	state.Runner = NewStandardRunner(state.Stdin, state.Stdout, state.Stderr)

	return &cli.Command{
		Name:    "dot",
		Usage:   "Unified CLI utility to manage dotfiles, local clusters, and workspaces",
		Version: Version,
		// Enable the hidden --generate-shell-completion flag that dot.fish drives for
		// dynamic self-completion. The auto-added command is renamed so it can't clash
		// with dot's own `completion` command (which generates completions for external tools).
		EnableShellCompletion:      true,
		ShellCompletionCommandName: "__completion",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to the configuration file",
				Sources: cli.EnvVars("DOT_CONFIG_PATH"),
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Usage:   "Enable verbose debug logging",
				Sources: cli.EnvVars("DOT_VERBOSE"),
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			level := slog.LevelInfo
			if cmd.Bool("verbose") {
				level = slog.LevelDebug
			}
			handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
			logger := slog.New(handler)
			slog.SetDefault(logger)
			state.Logger = logger

			configPath := cmd.String("config")
			resolved, _, perr := ConfigFilePath(configPath)
			if perr == nil {
				state.ConfigPath = resolved
			}
			cfg, err := LoadConfig(configPath)
			if err != nil {
				// A missing file is never fatal: `config init`/`edit` scaffold it and every
				// command falls back to defaults. A malformed/unreadable config, however, IS
				// fatal — strict decoding exists to fail loudly rather than silently revert to
				// defaults (which would, e.g., make `pull` operate on the wrong directories).
				// The `config` group is exempt so its edit/init/validate commands stay reachable
				// to repair the very file that failed to parse.
				fatal := !errors.Is(err, os.ErrNotExist)
				if sub := cmd.Args().First(); sub == "config" || sub == "f" {
					fatal = false
				}
				if fatal {
					return ctx, err
				}
				state.Logger.Warn("Falling back to default configuration", "path", resolved, "error", err)
			}
			state.Config = cfg
			return ctx, nil
		},
		Commands: []*cli.Command{
			NewVerifyCmd(state),
			NewPullCmd(state),
			NewCommitCmd(state),
			NewClusterCmd(state),
			NewLoginCmd(state),
			NewSetupCmd(state),
			NewCompletionCmd(state),
			NewPrCmd(state),
			NewReleaseCmd(state),
			NewStatusCmd(state),
			NewChezmoiCmd(state),
			NewConfigCmd(state),
			NewVersionCmd(state),
			NewAgentCmd(state),
		},
	}
}

package dot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

// NewConfigCmd constructs the top-level config command group. Every other command
// consumes ~/.config/dot.yaml; this group makes that file discoverable: inspect the
// effective configuration, locate it, scaffold a documented starter, edit, or validate.
func NewConfigCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "config",
		Aliases: []string{"f"},
		Usage:   "Inspect, scaffold, edit, and validate the dot configuration file",
		Commands: []*cli.Command{
			NewConfigShowCmd(state),
			NewConfigPathCmd(state),
			NewConfigInitCmd(state),
			NewConfigEditCmd(state),
			NewConfigValidateCmd(state),
		},
	}
}

// NewConfigShowCmd prints the effective configuration as YAML.
func NewConfigShowCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "show",
		Aliases: []string{"s"},
		Usage:   "Print the effective configuration (defaults merged with the file) as YAML",
		Action: func(_ context.Context, _ *cli.Command) error {
			return RunConfigShow(state)
		},
	}
}

// RunConfigShow marshals the loaded configuration back to YAML for inspection.
func RunConfigShow(state *GlobalState) error {
	out, err := yaml.Marshal(state.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	_, _ = fmt.Fprint(state.Stdout, string(out))
	return nil
}

// NewConfigPathCmd prints the resolved configuration file path.
func NewConfigPathCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "path",
		Aliases: []string{"p"},
		Usage:   "Print the resolved configuration file path",
		Action: func(_ context.Context, _ *cli.Command) error {
			_, _ = fmt.Fprintln(state.Stdout, state.ConfigPath)
			return nil
		},
	}
}

// NewConfigInitCmd scaffolds a config file populated with the built-in defaults.
func NewConfigInitCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "init",
		Aliases: []string{"i"},
		Usage:   "Write a starter configuration file populated with the built-in defaults",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Overwrite an existing configuration file",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			return RunConfigInit(state, cmd.Bool("force"))
		},
	}
}

// RunConfigInit writes DefaultConfig() as YAML to the resolved path, refusing to
// clobber an existing file unless force is set.
func RunConfigInit(state *GlobalState, force bool) error {
	path := state.ConfigPath
	if path == "" {
		return errors.New("could not resolve configuration path")
	}
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("config file already exists at %s (use --force to overwrite)", path)
	}

	out, err := yaml.Marshal(DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	_, _ = fmt.Fprintf(state.Stdout, "%s\n", green("✓ Wrote default configuration to "+path))
	return nil
}

// NewConfigEditCmd opens the configuration file in $EDITOR.
func NewConfigEditCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "edit",
		Aliases: []string{"e"},
		Usage:   "Open the configuration file in $EDITOR (scaffolds it first if missing)",
		Action: func(ctx context.Context, _ *cli.Command) error {
			return RunConfigEdit(ctx, state)
		},
	}
}

// RunConfigEdit ensures a config file exists, then opens it in the user's editor.
func RunConfigEdit(ctx context.Context, state *GlobalState) error {
	path := state.ConfigPath
	if path == "" {
		return errors.New("could not resolve configuration path")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := RunConfigInit(state, false); err != nil {
			return err
		}
	}

	// EDITOR may carry flags (e.g. "code -w"); split so the binary and args are separate.
	// Fields also drops a whitespace-only EDITOR, which would otherwise slip past cmp.Or.
	fields := strings.Fields(os.Getenv("EDITOR"))
	if len(fields) == 0 {
		fields = []string{"vi"}
	}
	editor := fields[0]
	if _, err := state.Runner.LookPath(editor); err != nil {
		return fmt.Errorf("editor %q not found in PATH", editor)
	}
	args := append(fields[1:], path)
	return state.Runner.RunInteractive(ctx, "", editor, args...)
}

// NewConfigValidateCmd checks that the configuration file parses under strict decoding.
func NewConfigValidateCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "validate",
		Aliases: []string{"v"},
		Usage:   "Validate that the configuration file parses (strict, unknown keys rejected)",
		Action: func(_ context.Context, _ *cli.Command) error {
			return RunConfigValidate(state)
		},
	}
}

// RunConfigValidate reports whether the resolved config file parses cleanly. A missing
// file is not an error: the built-in defaults simply apply.
func RunConfigValidate(state *GlobalState) error {
	path := state.ConfigPath
	if path == "" {
		return errors.New("could not resolve configuration path")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_, _ = fmt.Fprintf(state.Stdout, "%s No config file at %s; built-in defaults are in effect.\n", skipIcon, path)
		return nil
	}
	if _, err := LoadConfig(path); err != nil {
		_, _ = fmt.Fprintf(state.Stdout, "%s %v\n", failIcon, err)
		return err
	}
	_, _ = fmt.Fprintf(state.Stdout, "%s\n", green("✓ Configuration at "+path+" is valid."))
	return nil
}

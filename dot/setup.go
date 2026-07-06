package dot

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

// NewSetupCmd constructs the top-level setup command group.
func NewSetupCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "setup",
		Aliases: []string{"u"},
		Usage:   "Setup wrappers for external services and environments",
		Commands: []*cli.Command{
			NewSetupWorkspaceCmd(state),
		},
	}
}

// NewSetupWorkspaceCmd constructs the workspace setup subcommand.
func NewSetupWorkspaceCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:      "workspace",
		Aliases:   []string{"w"},
		Usage:     "Configure GCP project APIs and associate with gws CLI",
		ArgsUsage: "[PROJECT_ID]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			projectID := cmd.Args().First()
			return RunSetupWorkspace(ctx, state, projectID)
		},
	}
}

// RunSetupWorkspace configures Google Workspace APIs for the specified GCP project.
// If projectID is empty, it falls back to the GWS_PROJECT environment variable and
// returns an error when neither is provided.
func RunSetupWorkspace(ctx context.Context, state *GlobalState, projectID string) error {
	if _, err := state.Runner.LookPath("gws"); err != nil {
		return ErrGwsNotInstalled
	}
	if _, err := state.Runner.LookPath("gcloud"); err != nil {
		return ErrGcloudNotInstalled
	}

	if projectID == "" {
		projectID = os.Getenv(EnvGWSProject)
	}

	if projectID == "" {
		return fmt.Errorf("provide a project ID as an argument or set the %s environment variable", EnvGWSProject)
	}

	apis := state.Config.Setup.WorkspaceAPIs
	if len(apis) == 0 {
		return errors.New("no Google Workspace APIs configured to enable")
	}

	_, _ = fmt.Fprintf(state.Stdout, "gws: enabling Workspace APIs on project '%s'...\n", projectID)
	enableArgs := append([]string{"services", "enable"}, apis...)
	enableArgs = append(enableArgs, "--project", projectID, "--quiet")
	err := state.Runner.RunInteractive(ctx, "", "gcloud", enableArgs...)
	if err != nil {
		return fmt.Errorf("failed to enable gcloud services: %w", err)
	}

	_, _ = fmt.Fprintf(state.Stdout, "gws: configuring project '%s'...\n", projectID)
	return state.Runner.RunInteractive(ctx, "", "gws", "auth", "setup", "--project", projectID)
}

// SetupConfig represents the workspace setup configuration.
type SetupConfig struct {
	WorkspaceAPIs []string `yaml:"workspace_apis"`
}

func defaultSetupConfig() SetupConfig {
	return SetupConfig{
		WorkspaceAPIs: []string{
			"calendar-json.googleapis.com",
			"chat.googleapis.com",
			"docs.googleapis.com",
			"drive.googleapis.com",
			"forms.googleapis.com",
			"gmail.googleapis.com",
			"keep.googleapis.com",
			"meet.googleapis.com",
			"people.googleapis.com",
			"script.googleapis.com",
			"sheets.googleapis.com",
			"slides.googleapis.com",
			"tasks.googleapis.com",
		},
	}
}

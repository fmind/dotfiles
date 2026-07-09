package dot

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

// NewLoginCmd constructs the top-level login command group.
func NewLoginCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "login",
		Aliases: []string{"l"},
		Usage:   "Authentication wrappers for external service CLI tools",
		Commands: []*cli.Command{
			NewLoginGithubCmd(state),
			NewLoginWorkspaceCmd(state),
			NewLoginGcpCmd(state),
			NewLoginClaspCmd(state),
		},
	}
}

// NewLoginGithubCmd constructs the github login subcommand.
func NewLoginGithubCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "github",
		Aliases: []string{"g"},
		Usage:   "Interactive OAuth login for github.com via gh CLI",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunLoginGithub(ctx, state)
		},
	}
}

// NewLoginWorkspaceCmd constructs the workspace login subcommand.
func NewLoginWorkspaceCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "workspace",
		Aliases: []string{"w"},
		Usage:   "Interactive OAuth login for Google Workspace and GCP via gws CLI",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunLoginWorkspace(ctx, state)
		},
	}
}

// NewLoginGcpCmd constructs the GCP login subcommand.
func NewLoginGcpCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "gcp",
		Aliases: []string{"c"},
		Usage:   "Interactive login for Google Cloud SDK (gcloud) and Application Default Credentials (ADC)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunLoginGcp(ctx, state)
		},
	}
}

// NewLoginClaspCmd constructs the clasp login subcommand.
func NewLoginClaspCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "clasp",
		Aliases: []string{"a"},
		Usage:   "Interactive OAuth login for Google Apps Script via clasp CLI",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunLoginClasp(ctx, state)
		},
	}
}

// RunLoginGithub triggers interactive github authentication.
func RunLoginGithub(ctx context.Context, state *GlobalState) error {
	host := state.Config.Login.GithubHost
	scopes := strings.Join(state.Config.Login.GithubScopes, ",")

	if _, err := state.Runner.LookPath("gh"); err != nil {
		return ErrGhNotInstalled
	}

	_, err := state.Runner.Run(ctx, "", nil, "gh", "auth", "status", "--hostname", host)
	if err == nil {
		question := fmt.Sprintf("gh: already authenticated on %s. Re-authenticate? [y/N]: ", host)
		if !confirm(state.Stdin, state.Stdout, question) {
			_, _ = fmt.Fprintln(state.Stdout, "Canceled.")
			return nil
		}
	}

	_, _ = fmt.Fprintf(state.Stdout, "gh: requesting OAuth login for %s...\n", host)
	return state.Runner.RunInteractive(ctx, "", "gh", "auth", "login", "--hostname", host, "--scopes", scopes)
}

// RunLoginWorkspace triggers interactive Google Workspace/GCP authentication.
func RunLoginWorkspace(ctx context.Context, state *GlobalState) error {
	if _, err := state.Runner.LookPath("gws"); err != nil {
		return ErrGwsNotInstalled
	}

	_, _ = fmt.Fprintf(state.Stdout, "gws: requesting OAuth login (%d scopes)...\n", len(state.Config.Login.WorkspaceScopes))
	scopesCSV := strings.Join(state.Config.Login.WorkspaceScopes, ",")

	runner := state.Runner
	if _, ok := state.Runner.(*StandardRunner); ok {
		opener := &urlOpener{browser: state.Browser}
		runner = NewStandardRunner(
			state.Stdin,
			&urlOpenerWriter{w: state.Stdout, opener: opener},
			&urlOpenerWriter{w: state.Stderr, opener: opener},
		)
	}

	return runner.RunInteractive(ctx, "", "gws", "auth", "login", "--scopes", scopesCSV)
}

// RunLoginGcp triggers interactive GCP login and ADC credentials setup.
// Uses --update-adc to set both user and Application Default Credentials in a single auth flow.
func RunLoginGcp(ctx context.Context, state *GlobalState) error {
	if _, err := state.Runner.LookPath("gcloud"); err != nil {
		return ErrGcloudNotInstalled
	}

	_, _ = fmt.Fprintln(state.Stdout, "gcloud: authenticating user and Application Default Credentials (ADC)...")
	err := state.Runner.RunInteractive(ctx, "", "gcloud", "auth", "login", "--update-adc")
	if err != nil {
		return fmt.Errorf("gcloud login failed: %w", err)
	}

	_, _ = fmt.Fprintln(state.Stdout, "gcloud: credentials successfully updated.")
	return nil
}

// RunLoginClasp triggers interactive Google Apps Script authentication via clasp CLI.
func RunLoginClasp(ctx context.Context, state *GlobalState) error {
	if _, err := state.Runner.LookPath("clasp"); err != nil {
		return ErrClaspNotInstalled
	}

	home, err := os.UserHomeDir()
	if err == nil {
		claspJSON := filepath.Join(home, ".clasprc.json")
		if _, err := os.Stat(claspJSON); err == nil {
			question := "clasp: already authenticated. Re-authenticate? [y/N]: "
			if !confirm(state.Stdin, state.Stdout, question) {
				_, _ = fmt.Fprintln(state.Stdout, "Canceled.")
				return nil
			}
		}
	}

	_, _ = fmt.Fprintln(state.Stdout, "clasp: requesting Apps Script login...")
	return state.Runner.RunInteractive(ctx, "", "clasp", "login")
}

// confirm prompts the user on stdout with the message and reads response from stdin.
func confirm(stdin io.Reader, stdout io.Writer, msg string) bool {
	_, _ = fmt.Fprint(stdout, msg)
	reader := bufio.NewReader(stdin)
	ans, _ := reader.ReadString('\n')
	ans = strings.TrimSpace(strings.ToLower(ans))
	return ans == "y" || ans == "yes"
}

// LoginConfig represents the configuration for CLI authentication.
type LoginConfig struct {
	GithubHost      string   `yaml:"github_host"`
	GithubScopes    []string `yaml:"github_scopes"`
	WorkspaceScopes []string `yaml:"workspace_scopes"`
}

func defaultLoginConfig() LoginConfig {
	return LoginConfig{
		GithubHost: "github.com",
		GithubScopes: []string{
			"gist", "notifications", "read:org", "repo", "user",
		},
		WorkspaceScopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/user.emails.read",
			"https://www.googleapis.com/auth/cloud-platform",
			"https://www.googleapis.com/auth/calendar",
			"https://www.googleapis.com/auth/contacts",
			"https://www.googleapis.com/auth/directory.readonly",
			"https://www.googleapis.com/auth/documents",
			"https://www.googleapis.com/auth/drive",
			"https://www.googleapis.com/auth/forms.body",
			"https://www.googleapis.com/auth/forms.responses.readonly",
			"https://www.googleapis.com/auth/gmail.modify",
			"https://www.googleapis.com/auth/meetings.space.created",
			"https://www.googleapis.com/auth/meetings.space.readonly",
			"https://www.googleapis.com/auth/meetings.space.settings",
			"https://www.googleapis.com/auth/presentations",
			"https://www.googleapis.com/auth/spreadsheets",
			"https://www.googleapis.com/auth/tasks",
			"https://www.googleapis.com/auth/chat.spaces",
			"https://www.googleapis.com/auth/chat.messages",
			"https://www.googleapis.com/auth/chat.memberships",
			"https://www.googleapis.com/auth/script.projects",
			"https://www.googleapis.com/auth/script.deployments",
			"https://www.googleapis.com/auth/script.processes",
		},
	}
}

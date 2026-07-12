package dot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

// NewClusterCmd constructs the top-level cluster command group.
func NewClusterCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "cluster",
		Aliases: []string{"k"},
		Usage:   "Manage the local development k3d Kubernetes cluster",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "kubeconfig",
				Aliases: []string{"k"},
				Usage:   "Path to the isolated kubeconfig file",
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			kubeconfig := cmd.String("kubeconfig")
			if kubeconfig == "" && state.Config.Cluster.KubeconfigPath != "" {
				kubeconfig = state.Config.Cluster.KubeconfigPath
			}
			if kubeconfig != "" {
				kubeconfig = ExpandPath(kubeconfig)
				if err := os.Setenv("KUBECONFIG", kubeconfig); err != nil {
					return ctx, fmt.Errorf("failed to set KUBECONFIG: %w", err)
				}
			}
			return ctx, nil
		},
		Commands: []*cli.Command{
			NewClusterStartCmd(state),
			NewClusterStopCmd(state),
			NewClusterStatusCmd(state),
			NewClusterDeleteCmd(state),
			NewClusterNamespaceCmd(state),
		},
	}
}

// NewClusterStartCmd constructs the cluster start subcommand.
func NewClusterStartCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "start",
		Aliases: []string{"s"},
		Usage:   "Idempotently start or create the local cluster",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunClusterStart(ctx, state)
		},
	}
}

// NewClusterStopCmd constructs the cluster stop subcommand.
func NewClusterStopCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "stop",
		Aliases: []string{"x"},
		Usage:   "Stop the local cluster",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunClusterStop(ctx, state)
		},
	}
}

// NewClusterStatusCmd constructs the cluster status subcommand.
func NewClusterStatusCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "status",
		Aliases: []string{"t"},
		Usage:   "Check local cluster status",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunClusterStatus(ctx, state)
		},
	}
}

// NewClusterDeleteCmd constructs the cluster delete subcommand.
func NewClusterDeleteCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "delete",
		Aliases: []string{"d"},
		Usage:   "Delete the local cluster",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Skip the confirmation prompt",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RunClusterDelete(ctx, state, cmd.Bool("yes"))
		},
	}
}

// requireTools ensures each named CLI dependency is available in PATH.
func requireTools(state *GlobalState, tools ...string) error {
	for _, tool := range tools {
		if _, err := state.Runner.LookPath(tool); err != nil {
			return fmt.Errorf("%w: %s", ErrToolNotInstalled, tool)
		}
	}
	return nil
}

// RunClusterStart checks Docker daemon availability, verifies or creates the k3d cluster config, merges credentials, and waits for nodes to be ready.
func RunClusterStart(ctx context.Context, state *GlobalState) error {
	name := state.Config.Cluster.Name
	configPath := ExpandPath(state.Config.Cluster.ConfigPath)

	if err := requireTools(state, "docker", "k3d", "kubectl"); err != nil {
		return err
	}

	if _, err := state.Runner.Run(ctx, "", nil, "docker", "info"); err != nil {
		return errors.New("docker daemon is not running. please start docker")
	}

	listOut, err := state.Runner.Run(ctx, "", nil, "k3d", "cluster", "list", name)
	clusterExists := err == nil && listOut != ""

	if clusterExists {
		_, _ = fmt.Fprintf(state.Stdout, "Starting cluster '%s'...\n", name)
		_, err = state.Runner.Run(ctx, "", nil, "k3d", "cluster", "start", name)
		if err != nil {
			return fmt.Errorf("failed to start cluster: %w", err)
		}
	} else {
		if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
			return fmt.Errorf("config file not found at %s", configPath)
		}
		_, _ = fmt.Fprintf(state.Stdout, "Creating cluster '%s' using config %s...\n", name, configPath)
		err = state.Runner.RunInteractive(ctx, "", "k3d", "cluster", "create", name, "--config", configPath)
		if err != nil {
			return fmt.Errorf("failed to create cluster: %w", err)
		}
	}

	_, _ = fmt.Fprintln(state.Stdout, "Updating kubeconfig...")
	_, err = state.Runner.Run(ctx, "", nil, "k3d", "kubeconfig", "merge", name, "--kubeconfig-merge-default", "--kubeconfig-switch-context")
	if err != nil {
		return fmt.Errorf("failed to merge kubeconfig: %w", err)
	}

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		_, _ = fmt.Fprintf(state.Stdout, "Isolated kubeconfig updated. To use it in your terminal, run:\n  export KUBECONFIG=%s\n", kubeconfig)
	}

	_, _ = fmt.Fprintln(state.Stdout, "Waiting for nodes to be ready...")
	err = state.Runner.RunInteractive(ctx, "", "kubectl", "--request-timeout=15s", "wait", "--for=condition=Ready", "nodes", "--all", "--timeout=30s")
	if err != nil {
		_, _ = fmt.Fprintf(state.Stderr, "Warning: Nodes are not fully ready yet: %v. Services will continue reconciling in background.\n", err)
	}

	_, _ = fmt.Fprintf(state.Stdout, "Cluster context set to 'k3d-%s'. Services will reconcile in background.\n", name)
	return nil
}

// RunClusterStop stops the local development Kubernetes cluster.
func RunClusterStop(ctx context.Context, state *GlobalState) error {
	name := state.Config.Cluster.Name
	if err := requireTools(state, "k3d"); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(state.Stdout, "Stopping cluster '%s'...\n", name)
	return state.Runner.RunInteractive(ctx, "", "k3d", "cluster", "stop", name)
}

// RunClusterStatus queries k3d and kubectl to display development cluster node details.
func RunClusterStatus(ctx context.Context, state *GlobalState) error {
	name := state.Config.Cluster.Name
	if err := requireTools(state, "k3d", "kubectl"); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(state.Stdout, "=> k3d cluster list:")
	if err := state.Runner.RunInteractive(ctx, "", "k3d", "cluster", "list", name); err != nil {
		return fmt.Errorf("failed to list k3d clusters: %w", err)
	}

	_, _ = fmt.Fprintln(state.Stdout, "\n=> kubectl node status:")
	if err := state.Runner.RunInteractive(ctx, "", "kubectl", "get", "nodes", "-o", "wide"); err != nil {
		return fmt.Errorf("failed to get kubectl nodes: %w", err)
	}
	return nil
}

// RunClusterDelete tears down and removes the local development Kubernetes cluster.
// Because the cluster is shared across every local project, deletion is guarded by a
// confirmation prompt unless autoApprove is set.
func RunClusterDelete(ctx context.Context, state *GlobalState, autoApprove bool) error {
	name := state.Config.Cluster.Name
	if err := requireTools(state, "k3d"); err != nil {
		return err
	}

	if !autoApprove {
		question := fmt.Sprintf("Delete cluster '%s'? This destroys all namespaces and workloads. [y/N]: ", name)
		if !confirm(state.Stdin, state.Stdout, question) {
			_, _ = fmt.Fprintln(state.Stdout, "Canceled.")
			return nil
		}
	}

	_, _ = fmt.Fprintf(state.Stdout, "Deleting cluster '%s'...\n", name)
	return state.Runner.RunInteractive(ctx, "", "k3d", "cluster", "delete", name)
}

// NewClusterNamespaceCmd constructs the cluster namespace subcommand.
func NewClusterNamespaceCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:      "namespace",
		Aliases:   []string{"n", "ns"},
		Usage:     "Idempotently create and switch context to a Kubernetes namespace",
		ArgsUsage: "[NAMESPACE]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.Args().First()
			return RunClusterNamespace(ctx, state, name)
		},
	}
}

// RunClusterNamespace idempotently creates a namespace and switches the current context to it.
func RunClusterNamespace(ctx context.Context, state *GlobalState, name string) error {
	if name == "" {
		return errors.New("namespace name is required")
	}

	if err := requireTools(state, "kubectl"); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(state.Stdout, "Checking namespace '%s'...\n", name)
	// --ignore-not-found lets us tell a genuinely missing namespace (empty output, no
	// error) apart from a real failure (cluster unreachable, wrong context, unauthorized),
	// so we never misreport a transport/auth error as "not found" and then fail on create.
	out, err := state.Runner.Run(ctx, "", nil, "kubectl", "get", "namespace", name, "--ignore-not-found", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to query namespace '%s': %w", name, err)
	}
	if strings.TrimSpace(out) == "" {
		_, _ = fmt.Fprintf(state.Stdout, "Namespace '%s' not found. Creating...\n", name)
		if _, createErr := state.Runner.Run(ctx, "", nil, "kubectl", "create", "namespace", name); createErr != nil {
			return fmt.Errorf("failed to create namespace '%s': %w", name, createErr)
		}
		_, _ = fmt.Fprintf(state.Stdout, "Namespace '%s' created.\n", name)
	} else {
		_, _ = fmt.Fprintf(state.Stdout, "Namespace '%s' already exists.\n", name)
	}

	_, _ = fmt.Fprintf(state.Stdout, "Setting current context namespace to '%s'...\n", name)
	_, err = state.Runner.Run(ctx, "", nil, "kubectl", "config", "set-context", "--current", "--namespace", name)
	if err != nil {
		return fmt.Errorf("failed to set context namespace: %w", err)
	}
	_, _ = fmt.Fprintf(state.Stdout, "Current context namespace set to '%s'.\n", name)
	return nil
}

// ClusterConfig represents the configuration for local cluster management.
type ClusterConfig struct {
	Name           string `yaml:"name"`
	ConfigPath     string `yaml:"config_path"`
	KubeconfigPath string `yaml:"kubeconfig_path"`
}

func defaultClusterConfig() ClusterConfig {
	return ClusterConfig{
		Name:           "local",
		ConfigPath:     "~/.config/k3d/local.yaml",
		KubeconfigPath: "",
	}
}

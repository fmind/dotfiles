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
			&cli.StringFlag{
				Name:  "context",
				Usage: "Explicit context inside the managed cluster kubeconfig",
			},
		},
		Commands: []*cli.Command{
			NewClusterStartCmd(state),
			NewClusterStopCmd(state),
			NewClusterStatusCmd(state),
			NewClusterDiagnoseCmd(state),
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
			return RunClusterStartWithOptions(ctx, state, clusterTargetOptions(cmd))
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
			return RunClusterStopWithOptions(ctx, state, clusterTargetOptions(cmd))
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
			return RunClusterStatusWithOptions(ctx, state, clusterTargetOptions(cmd))
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
			return RunClusterDeleteWithOptions(ctx, state, cmd.Bool("yes"), clusterTargetOptions(cmd))
		},
	}
}

func clusterTargetOptions(cmd *cli.Command) ClusterTargetOptions {
	return ClusterTargetOptions{KubeconfigPath: cmd.String("kubeconfig"), Context: cmd.String("context")}
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

func requireDocker(ctx context.Context, state *GlobalState) error {
	if _, err := state.Runner.Run(ctx, "", nil, "docker", "info"); err != nil {
		// Keep the probe's own error: a permission-denied socket or a broken
		// context needs a different fix than a daemon that is simply stopped.
		return fmt.Errorf("docker daemon is not running or not accessible: %w", err)
	}
	return nil
}

func k3dClusterExists(ctx context.Context, state *GlobalState, name string) (bool, error) {
	// List every cluster instead of filtering by name: `k3d cluster list <name>`
	// exits non-zero when the name is unknown, which would make "does not exist"
	// indistinguishable from a real k3d failure and block first-time creation.
	listOut, err := state.Runner.Run(ctx, "", nil, "k3d", "cluster", "list", "--no-headers")
	if err != nil {
		return false, fmt.Errorf("failed to inspect cluster %q: %w", name, err)
	}
	for line := range strings.SplitSeq(strings.TrimSpace(listOut), "\n") {
		if fields := strings.Fields(line); len(fields) > 0 && fields[0] == name {
			return true, nil
		}
	}
	return false, nil
}

// RunClusterStart checks Docker, verifies or creates the named k3d cluster without
// changing the default kubeconfig, and waits for the isolated target to be ready.
func RunClusterStart(ctx context.Context, state *GlobalState) error {
	return RunClusterStartWithOptions(ctx, state, ClusterTargetOptions{})
}

func RunClusterStartWithOptions(ctx context.Context, state *GlobalState, options ClusterTargetOptions) error {
	name := state.Config.Cluster.Name
	configPath := ExpandPath(state.Config.Cluster.ConfigPath)

	if err := requireTools(state, "docker", "k3d", "kubectl"); err != nil {
		return err
	}

	if err := requireDocker(ctx, state); err != nil {
		return err
	}

	clusterExists, err := k3dClusterExists(ctx, state, name)
	if err != nil {
		return err
	}

	var kubeconfig string
	if clusterExists {
		if ensureErr := ensureClusterKubeconfig(ctx, state, options); ensureErr != nil {
			return ensureErr
		}
		kubeconfig, err = clusterKubeconfigPath(state.Config.Cluster, options)
		if err != nil {
			return err
		}
		target, resolveErr := resolveClusterTarget(ctx, state, options, false)
		if resolveErr != nil {
			return resolveErr
		}
		reportClusterTarget(state, target)
		_, _ = fmt.Fprintf(state.Stdout, "Starting cluster '%s'...\n", name)
		_, err = state.Runner.Run(ctx, "", nil, "k3d", "cluster", "start", name)
		if err != nil {
			return fmt.Errorf("failed to start cluster: %w", err)
		}
	} else {
		if options.Context != "" && options.Context != expectedClusterContext(name) {
			return fmt.Errorf("context override %q cannot target cluster %q before it exists", options.Context, name)
		}
		if _, statErr := os.Stat(configPath); statErr != nil {
			if os.IsNotExist(statErr) {
				return fmt.Errorf("config file not found at %s", configPath)
			}
			return fmt.Errorf("failed to inspect cluster config %s: %w", configPath, statErr)
		}
		target, targetErr := intendedClusterTarget(state, options)
		if targetErr != nil {
			return targetErr
		}
		reportClusterTarget(state, target)
		if existsNow, inspectErr := k3dClusterExists(ctx, state, name); inspectErr != nil {
			return inspectErr
		} else if existsNow {
			return fmt.Errorf("cluster %q appeared before creation; refusing an ambiguous target", name)
		}
		_, _ = fmt.Fprintf(state.Stdout, "Creating cluster '%s' using config %s...\n", name, configPath)
		err = state.Runner.RunInteractive(ctx, "", "k3d", "cluster", "create", name, "--config", configPath, "--kubeconfig-update-default=false", "--kubeconfig-switch-context=false")
		if err != nil {
			return fmt.Errorf("failed to create cluster: %w", err)
		}
		kubeconfig, err = refreshClusterKubeconfig(ctx, state, options)
		if err != nil {
			return err
		}
	}
	target, err := resolveClusterTarget(ctx, state, options, true)
	if err != nil {
		return err
	}
	reportClusterTarget(state, target)
	_, _ = fmt.Fprintf(state.Stdout, "Isolated kubeconfig updated. To use it in your terminal, run:\n  export KUBECONFIG=%s\n", kubeconfig)

	_, _ = fmt.Fprintln(state.Stdout, "Waiting for nodes to be ready...")
	err = state.Runner.RunInteractive(ctx, "", "kubectl", target.kubectlArgs("--request-timeout=15s", "wait", "--for=condition=Ready", "nodes", "--all", "--timeout=30s")...)
	if err != nil {
		_, _ = fmt.Fprintf(state.Stderr, "Warning: Nodes are not fully ready yet: %v. Services will continue reconciling in background.\n", err)
	}

	_, _ = fmt.Fprintf(state.Stdout, "Cluster context '%s' is isolated; the default kubeconfig was not changed. Services will reconcile in background.\n", target.Context)
	return nil
}

// RunClusterStop stops the local development Kubernetes cluster.
func RunClusterStop(ctx context.Context, state *GlobalState) error {
	return RunClusterStopWithOptions(ctx, state, ClusterTargetOptions{})
}

func RunClusterStopWithOptions(ctx context.Context, state *GlobalState, options ClusterTargetOptions) error {
	name := state.Config.Cluster.Name
	if err := requireTools(state, "docker", "k3d", "kubectl"); err != nil {
		return err
	}
	if err := requireDocker(ctx, state); err != nil {
		return err
	}
	if err := ensureClusterKubeconfig(ctx, state, options); err != nil {
		return err
	}
	target, err := resolveClusterTarget(ctx, state, options, true)
	if err != nil {
		return err
	}
	reportClusterTarget(state, target)
	_, _ = fmt.Fprintf(state.Stdout, "Stopping cluster '%s'...\n", name)
	return state.Runner.RunInteractive(ctx, "", "k3d", "cluster", "stop", name)
}

// RunClusterStatus queries k3d and kubectl to display development cluster node details.
func RunClusterStatus(ctx context.Context, state *GlobalState) error {
	return RunClusterStatusWithOptions(ctx, state, ClusterTargetOptions{})
}

func RunClusterStatusWithOptions(ctx context.Context, state *GlobalState, options ClusterTargetOptions) error {
	name := state.Config.Cluster.Name
	if err := requireTools(state, "k3d", "kubectl"); err != nil {
		return err
	}
	if err := ensureClusterKubeconfig(ctx, state, options); err != nil {
		return err
	}
	target, err := resolveClusterTarget(ctx, state, options, false)
	if err != nil {
		return err
	}
	reportClusterTarget(state, target)

	_, _ = fmt.Fprintln(state.Stdout, "=> k3d cluster list:")
	if err := state.Runner.RunInteractive(ctx, "", "k3d", "cluster", "list", name); err != nil {
		return fmt.Errorf("failed to list k3d clusters: %w", err)
	}

	_, _ = fmt.Fprintln(state.Stdout, "\n=> kubectl node status:")
	if err := state.Runner.RunInteractive(ctx, "", "kubectl", target.kubectlArgs("get", "nodes", "-o", "wide")...); err != nil {
		return fmt.Errorf("failed to get kubectl nodes: %w", err)
	}
	return nil
}

// RunClusterDeleteWithOptions tears down and removes the local development Kubernetes
// cluster. Because the cluster is shared across every local project, deletion is guarded
// by a confirmation prompt unless autoApprove is set.
func RunClusterDeleteWithOptions(ctx context.Context, state *GlobalState, autoApprove bool, options ClusterTargetOptions) error {
	name := state.Config.Cluster.Name
	if err := requireTools(state, "docker", "k3d", "kubectl"); err != nil {
		return err
	}

	if !autoApprove {
		question := fmt.Sprintf("Delete cluster '%s'? This destroys all namespaces and workloads. [y/N]: ", name)
		if !confirm(state.Stdin, state.Stdout, question) {
			_, _ = fmt.Fprintln(state.Stdout, "Canceled.")
			return nil
		}
	}
	if err := requireDocker(ctx, state); err != nil {
		return err
	}
	if err := ensureClusterKubeconfig(ctx, state, options); err != nil {
		return err
	}
	target, err := resolveClusterTarget(ctx, state, options, true)
	if err != nil {
		return err
	}
	reportClusterTarget(state, target)
	_, _ = fmt.Fprintf(state.Stdout, "Deleting cluster '%s'...\n", name)
	return state.Runner.RunInteractive(ctx, "", "k3d", "cluster", "delete", name)
}

// NewClusterNamespaceCmd constructs the cluster namespace subcommand.
func NewClusterNamespaceCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:      "namespace",
		Aliases:   []string{"n", "ns"},
		Usage:     "Idempotently create and switch context to a Kubernetes namespace",
		ArgsUsage: "<NAMESPACE>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.Args().First()
			return RunClusterNamespaceWithOptions(ctx, state, name, clusterTargetOptions(cmd))
		},
	}
}

// RunClusterNamespace idempotently creates a namespace and switches the current context to it.
func RunClusterNamespace(ctx context.Context, state *GlobalState, name string) error {
	return RunClusterNamespaceWithOptions(ctx, state, name, ClusterTargetOptions{})
}

func RunClusterNamespaceWithOptions(ctx context.Context, state *GlobalState, name string, options ClusterTargetOptions) error {
	if name == "" {
		return errors.New("namespace name is required")
	}

	if err := requireTools(state, "docker", "k3d", "kubectl"); err != nil {
		return err
	}
	if err := requireDocker(ctx, state); err != nil {
		return err
	}
	if err := ensureClusterKubeconfig(ctx, state, options); err != nil {
		return err
	}
	target, err := resolveClusterTarget(ctx, state, options, true)
	if err != nil {
		return err
	}
	reportClusterTarget(state, target)

	_, _ = fmt.Fprintf(state.Stdout, "Checking namespace '%s'...\n", name)
	// --ignore-not-found lets us tell a genuinely missing namespace (empty output, no
	// error) apart from a real failure (cluster unreachable, wrong context, unauthorized),
	// so we never misreport a transport/auth error as "not found" and then fail on create.
	out, err := state.Runner.Run(ctx, "", nil, "kubectl", target.kubectlArgs("get", "namespace", name, "--ignore-not-found", "-o", "name")...)
	if err != nil {
		return fmt.Errorf("failed to query namespace '%s': %w", name, err)
	}
	if strings.TrimSpace(out) == "" {
		verifiedTarget, verifyErr := resolveClusterTarget(ctx, state, options, true)
		if verifyErr != nil {
			return verifyErr
		}
		reportClusterTarget(state, verifiedTarget)
		_, _ = fmt.Fprintf(state.Stdout, "Namespace '%s' not found. Creating...\n", name)
		if _, createErr := state.Runner.Run(ctx, "", nil, "kubectl", verifiedTarget.kubectlArgs("create", "namespace", name)...); createErr != nil {
			return fmt.Errorf("failed to create namespace '%s': %w", name, createErr)
		}
		_, _ = fmt.Fprintf(state.Stdout, "Namespace '%s' created.\n", name)
	} else {
		_, _ = fmt.Fprintf(state.Stdout, "Namespace '%s' already exists.\n", name)
	}

	target, err = resolveClusterTarget(ctx, state, options, true)
	if err != nil {
		return err
	}
	reportClusterTarget(state, target)
	_, _ = fmt.Fprintf(state.Stdout, "Setting isolated context '%s' namespace to '%s'...\n", target.Context, name)
	_, err = state.Runner.Run(ctx, "", nil, "kubectl", "config", "set-context", target.Context, "--namespace", name, "--kubeconfig", target.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to set context namespace: %w", err)
	}
	_, _ = fmt.Fprintf(state.Stdout, "Isolated context '%s' namespace set to '%s'; the default kubeconfig was not changed.\n", target.Context, name)
	return nil
}

// ClusterConfig represents the configuration for local cluster management.
type ClusterConfig struct {
	Name           string                  `yaml:"name"`
	ConfigPath     string                  `yaml:"config_path"`
	KubeconfigPath string                  `yaml:"kubeconfig_path"`
	Diagnostics    ClusterDiagnosticConfig `yaml:"diagnostics"`
}

func defaultClusterConfig() ClusterConfig {
	return ClusterConfig{
		Name:           "local",
		ConfigPath:     "~/.config/k3d/local.yaml",
		KubeconfigPath: "",
		Diagnostics:    ClusterDiagnosticConfig{},
	}
}

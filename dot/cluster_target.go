package dot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ClusterTargetOptions struct {
	KubeconfigPath string
	Context        string
}

type clusterTarget struct {
	Name        string
	Kubeconfig  string
	Context     string
	Namespace   string
	Server      string
	Fingerprint string
}

type kubeconfigMetadata struct {
	CurrentContext string `yaml:"current-context"`
	Clusters       []struct {
		Name    string `yaml:"name"`
		Cluster struct {
			Server string `yaml:"server"`
		} `yaml:"cluster"`
	} `yaml:"clusters"`
	Contexts []struct {
		Name    string `yaml:"name"`
		Context struct {
			Cluster   string `yaml:"cluster"`
			Namespace string `yaml:"namespace"`
		} `yaml:"context"`
	} `yaml:"contexts"`
}

type kubeconfigContextMetadata struct {
	Context   string
	Namespace string
	Server    string
}

func expectedClusterContext(name string) string {
	return "k3d-" + name
}

func validClusterName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			continue
		}
		return false
	}
	return true
}

func clusterKubeconfigPath(config ClusterConfig, options ClusterTargetOptions) (string, error) {
	if !validClusterName(config.Name) {
		return "", fmt.Errorf("invalid managed cluster name %q", config.Name)
	}
	path := options.KubeconfigPath
	if path == "" {
		path = config.KubeconfigPath
	}
	if path == "" {
		path = filepath.Join("~", ".kube", "dot", config.Name+".yaml")
	}
	return ExpandPath(path), nil
}

func parseKubeconfigMetadata(content []byte) (kubeconfigMetadata, error) {
	var metadata kubeconfigMetadata
	if err := yaml.Unmarshal(content, &metadata); err != nil {
		return metadata, fmt.Errorf("invalid kubeconfig metadata: %w", err)
	}
	return metadata, nil
}

func (metadata kubeconfigMetadata) context(name string) (kubeconfigContextMetadata, bool) {
	for _, candidate := range metadata.Contexts {
		if candidate.Name != name {
			continue
		}
		for _, cluster := range metadata.Clusters {
			if cluster.Name == candidate.Context.Cluster && cluster.Cluster.Server != "" {
				return kubeconfigContextMetadata{Context: name, Namespace: candidate.Context.Namespace, Server: cluster.Cluster.Server}, true
			}
		}
		return kubeconfigContextMetadata{}, false
	}
	return kubeconfigContextMetadata{}, false
}

func clusterOwnershipFingerprint(name string) string {
	digest := sha256.Sum256([]byte("k3d\x00" + name))
	return hex.EncodeToString(digest[:])[:12]
}

func sanitizedServerAddress(server string) string {
	parsed, err := url.Parse(server)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "<invalid>"
	}
	return parsed.Scheme + "://" + parsed.Host
}

func writeOwnerOnlyFile(path string, content []byte, description string) error {
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing linked %s path %s", description, path)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to inspect %s path: %w", description, err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create %s directory: %w", description, err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("failed to secure %s directory: %w", description, err)
	}
	temp, err := os.CreateTemp(dir, ".kubeconfig-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary %s: %w", description, err)
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("failed to secure temporary %s: %w", description, err)
	}
	if _, err := temp.Write(content); err != nil {
		_ = temp.Close()
		return fmt.Errorf("failed to write temporary %s: %w", description, err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("failed to sync temporary %s: %w", description, err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("failed to close temporary %s: %w", description, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("failed to publish %s: %w", description, err)
	}
	return nil
}

func writeIsolatedKubeconfig(path string, content []byte) error {
	return writeOwnerOnlyFile(path, content, "isolated kubeconfig")
}

func fetchClusterKubeconfig(ctx context.Context, state *GlobalState, name string) ([]byte, kubeconfigMetadata, error) {
	output, err := state.Runner.Run(ctx, "", nil, "k3d", "kubeconfig", "get", name)
	if err != nil {
		// k3d reports an absent and a stopped cluster the same way, through a bare
		// non-zero exit. That is the ordinary state on a fresh machine, so name it
		// here: the raw k3d stderr alone reads as a tooling fault rather than as
		// "there is no cluster yet".
		return nil, kubeconfigMetadata{}, fmt.Errorf("failed to get kubeconfig for cluster %q (it does not exist or is stopped): %w", name, err)
	}
	if strings.TrimSpace(output) == "" {
		return nil, kubeconfigMetadata{}, fmt.Errorf("k3d returned an empty kubeconfig for cluster %q", name)
	}
	metadata, err := parseKubeconfigMetadata([]byte(output))
	if err != nil {
		return nil, metadata, err
	}
	if _, ok := metadata.context(expectedClusterContext(name)); !ok {
		return nil, metadata, fmt.Errorf("k3d kubeconfig for cluster %q omitted expected context %q", name, expectedClusterContext(name))
	}
	return []byte(output), metadata, nil
}

func ensureClusterKubeconfig(ctx context.Context, state *GlobalState, options ClusterTargetOptions) error {
	path, err := clusterKubeconfigPath(state.Config.Cluster, options)
	if err != nil {
		return err
	}
	if info, statErr := os.Lstat(path); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return fmt.Errorf("isolated kubeconfig is not a regular file: %s", path)
		}
		if info.Mode().Perm()&0o077 != 0 {
			return fmt.Errorf("isolated kubeconfig permissions are %o, expected owner-only", info.Mode().Perm())
		}
		return nil
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return fmt.Errorf("failed to inspect isolated kubeconfig: %w", statErr)
	}
	content, _, err := fetchClusterKubeconfig(ctx, state, state.Config.Cluster.Name)
	if err != nil {
		return err
	}
	if err := writeIsolatedKubeconfig(path, content); err != nil {
		return err
	}
	return nil
}

func refreshClusterKubeconfig(ctx context.Context, state *GlobalState, options ClusterTargetOptions) (string, error) {
	path, err := clusterKubeconfigPath(state.Config.Cluster, options)
	if err != nil {
		return "", err
	}
	content, _, err := fetchClusterKubeconfig(ctx, state, state.Config.Cluster.Name)
	if err != nil {
		return "", err
	}
	if err := writeIsolatedKubeconfig(path, content); err != nil {
		return "", err
	}
	return path, nil
}

func resolveClusterTarget(ctx context.Context, state *GlobalState, options ClusterTargetOptions, requireAPI bool) (clusterTarget, error) {
	name := state.Config.Cluster.Name
	path, err := clusterKubeconfigPath(state.Config.Cluster, options)
	if err != nil {
		return clusterTarget{}, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return clusterTarget{}, fmt.Errorf("failed to inspect isolated kubeconfig %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return clusterTarget{}, fmt.Errorf("isolated kubeconfig is not a regular file: %s", path)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return clusterTarget{}, fmt.Errorf("isolated kubeconfig permissions are %o, expected owner-only", info.Mode().Perm())
	}
	content, err := os.ReadFile(path) //nolint:gosec // owner-only managed kubeconfig; only non-secret metadata is retained
	if err != nil {
		return clusterTarget{}, fmt.Errorf("failed to read isolated kubeconfig metadata: %w", err)
	}
	actual, err := parseKubeconfigMetadata(content)
	if err != nil {
		return clusterTarget{}, err
	}
	expectedContext := expectedClusterContext(name)
	selectedContext := options.Context
	if selectedContext == "" {
		selectedContext = expectedContext
	}
	actualContext, ok := actual.context(selectedContext)
	if !ok {
		return clusterTarget{}, fmt.Errorf("cluster target mismatch: expected context %q, actual current context %q; selected context %q is absent", expectedContext, actual.CurrentContext, selectedContext)
	}
	_, authoritative, err := fetchClusterKubeconfig(ctx, state, name)
	if err != nil {
		return clusterTarget{}, err
	}
	expectedMetadata, ok := authoritative.context(expectedContext)
	if !ok {
		return clusterTarget{}, fmt.Errorf("cluster target mismatch: expected context %q is absent from k3d metadata", expectedContext)
	}
	if actualContext.Server != expectedMetadata.Server {
		return clusterTarget{}, fmt.Errorf("cluster target mismatch: expected context %q server %q, actual context %q server %q", expectedContext, sanitizedServerAddress(expectedMetadata.Server), selectedContext, sanitizedServerAddress(actualContext.Server))
	}
	namespace := actualContext.Namespace
	if namespace == "" {
		namespace = "default"
	}
	target := clusterTarget{Name: name, Kubeconfig: path, Context: selectedContext, Namespace: namespace, Server: actualContext.Server, Fingerprint: clusterOwnershipFingerprint(name)}
	if requireAPI {
		args := target.kubectlArgs("--request-timeout=5s", "get", "--raw=/version")
		if _, err := state.Runner.Run(ctx, "", nil, "kubectl", args...); err != nil {
			return clusterTarget{}, fmt.Errorf("cluster API verification failed: expected context %q, actual context %q: %w", expectedContext, selectedContext, err)
		}
	}
	return target, nil
}

func (target clusterTarget) kubectlArgs(args ...string) []string {
	return append(args, "--kubeconfig", target.Kubeconfig, "--context", target.Context)
}

func reportClusterTarget(state *GlobalState, target clusterTarget) {
	_, _ = fmt.Fprintf(state.Stdout, "Target: cluster=%s context=%s namespace=%s fingerprint=%s kubeconfig=%s\n", target.Name, target.Context, target.Namespace, target.Fingerprint, target.Kubeconfig)
}

func intendedClusterTarget(state *GlobalState, options ClusterTargetOptions) (clusterTarget, error) {
	path, err := clusterKubeconfigPath(state.Config.Cluster, options)
	if err != nil {
		return clusterTarget{}, err
	}
	contextName := options.Context
	if contextName == "" {
		contextName = expectedClusterContext(state.Config.Cluster.Name)
	}
	return clusterTarget{Name: state.Config.Cluster.Name, Kubeconfig: path, Context: contextName, Namespace: "default", Fingerprint: clusterOwnershipFingerprint(state.Config.Cluster.Name)}, nil
}

package dot

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Skipping path expansion test because home dir is not available")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"/abs/path", "/abs/path"},
		{"relative/path", "relative/path"},
		{"~", home},
		{"~/foo", filepath.Join(home, "foo")},
		{"~bar", "~bar"}, // "~username" form is left untouched, not resolved as $HOME/bar
	}

	for _, tc := range tests {
		got := ExpandPath(tc.input)
		// On windows vs Unix path separators might differ, clean them
		gotClean := filepath.ToSlash(got)
		expClean := filepath.ToSlash(tc.expected)
		if gotClean != expClean {
			t.Errorf("ExpandPath(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if len(cfg.Pull.Directories) == 0 {
		t.Error("Expected default pull directories to be non-empty")
	}

	if len(cfg.Commit.AllowedTypes) == 0 {
		t.Error("Expected default allowed commit types to be non-empty")
	}

	if len(cfg.Verify.Tools) == 0 {
		t.Error("Expected default tools to check to be non-empty")
	}

	if cfg.Cluster.Name != "local" {
		t.Errorf("Expected default cluster name to be 'local', got %q", cfg.Cluster.Name)
	}

	if cfg.PR.BaseBranch != "main" {
		t.Errorf("Expected default PR base branch to be 'main', got %q", cfg.PR.BaseBranch)
	}
	if cfg.Completions.Path != "~/.config/fish/completions" {
		t.Errorf("Expected default Completions path to be '~/.config/fish/completions', got %q", cfg.Completions.Path)
	}
	if cfg.Login.GithubHost != "github.com" {
		t.Errorf("Expected default Login GitHub host to be 'github.com', got %q", cfg.Login.GithubHost)
	}
	if len(cfg.Login.GithubScopes) == 0 {
		t.Error("Expected default GithubScopes to be non-empty")
	}
	if len(cfg.Login.WorkspaceScopes) == 0 {
		t.Error("Expected default WorkspaceScopes to be non-empty")
	}
}

func TestLoadConfig_NonExistent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dot-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// We pass a path to a non-existent file, it should return default config and an error.
	nonExistentPath := filepath.Join(tempDir, "missing-dot.yaml")
	cfg, err := LoadConfig(nonExistentPath)
	if err == nil {
		t.Error("Expected error when config file does not exist at explicit path, got nil")
	}
	if cfg == nil {
		t.Fatal("Expected config to be loaded but got nil")
	}

	if cfg.Cluster.Name != "local" {
		t.Errorf("Expected cluster name 'local', got %q", cfg.Cluster.Name)
	}
}

func TestLoadConfig_CustomYaml(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dot-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	yamlContent := `
cluster:
  name: "custom-cluster"
  config_path: "/path/to/k3d.yaml"
verify:
  tools:
    - kubectl
    - helm
pr:
  base_branch: "develop"
  prompt: "custom pr prompt"
commit:
  prompt: "custom commit prompt"
completions:
  path: "/custom/completions/path"
login:
  github_host: "github.enterprise.local"
  github_scopes:
    - repo
    - user
  workspace_scopes:
    - openid
    - email
`
	configPath := filepath.Join(tempDir, "dot.yaml")
	err = os.WriteFile(configPath, []byte(yamlContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to write dot.yaml: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Cluster.Name != "custom-cluster" {
		t.Errorf("Expected cluster name 'custom-cluster', got %q", cfg.Cluster.Name)
	}
	if len(cfg.Verify.Tools) != 2 || cfg.Verify.Tools[0] != "kubectl" || cfg.Verify.Tools[1] != "helm" {
		t.Errorf("Expected verify tools ['kubectl', 'helm'], got %v", cfg.Verify.Tools)
	}
	if cfg.PR.BaseBranch != "develop" {
		t.Errorf("Expected PR base branch 'develop', got %q", cfg.PR.BaseBranch)
	}
	if cfg.PR.Prompt != "custom pr prompt" {
		t.Errorf("Expected PR prompt 'custom pr prompt', got %q", cfg.PR.Prompt)
	}
	if cfg.Commit.Prompt != "custom commit prompt" {
		t.Errorf("Expected Commit prompt 'custom commit prompt', got %q", cfg.Commit.Prompt)
	}
	if cfg.Completions.Path != "/custom/completions/path" {
		t.Errorf("Expected Completions path '/custom/completions/path', got %q", cfg.Completions.Path)
	}
	if cfg.Login.GithubHost != "github.enterprise.local" {
		t.Errorf("Expected Login GitHub host 'github.enterprise.local', got %q", cfg.Login.GithubHost)
	}
	if len(cfg.Login.GithubScopes) != 2 || cfg.Login.GithubScopes[0] != "repo" || cfg.Login.GithubScopes[1] != "user" {
		t.Errorf("Expected GithubScopes ['repo', 'user'], got %v", cfg.Login.GithubScopes)
	}
	if len(cfg.Login.WorkspaceScopes) != 2 || cfg.Login.WorkspaceScopes[0] != "openid" || cfg.Login.WorkspaceScopes[1] != "email" {
		t.Errorf("Expected WorkspaceScopes ['openid', 'email'], got %v", cfg.Login.WorkspaceScopes)
	}

	// Verify defaults are still populated if not specified
	if len(cfg.Pull.Directories) == 0 {
		t.Error("Expected pull directories to be filled with defaults if omitted from YAML")
	}
}

func TestLoadConfig_InvalidYaml(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dot-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	configPath := filepath.Join(tempDir, "dot.yaml")
	_ = os.WriteFile(configPath, []byte("invalid:::yaml: {[[}"), 0o600)

	cfg, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error when loading invalid YAML config, got nil")
	}
	if cfg == nil {
		t.Fatal("Expected fallback config to be returned despite error, got nil")
	}
	// Fallback should still have defaults
	if cfg.Cluster.Name != "local" {
		t.Errorf("Expected default cluster name 'local', got %q", cfg.Cluster.Name)
	}
}

func TestAppConfigFlag(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dot-app-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	yamlContent := `
cluster:
  name: "app-test-cluster"
  config_path: "/path/to/k3d.yaml"
`
	configPath := filepath.Join(tempDir, "dot.yaml")
	err = os.WriteFile(configPath, []byte(yamlContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to write dot.yaml: %v", err)
	}

	app := NewApp()

	// Run app with custom config flag
	err = app.Run(context.Background(), []string{"dot", "--config", configPath, "help"})
	if err != nil {
		t.Fatalf("Expected no error running app with --config, got %v", err)
	}
}

func TestLoadConfig_DefaultPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dot-home-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	t.Setenv("HOME", tempDir)

	// Test non-existent config file at default path
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Expected no error loading default config, got %v", err)
	}
	if cfg.Cluster.Name != "local" {
		t.Errorf("Expected fallback default config cluster 'local', got %q", cfg.Cluster.Name)
	}

	// Create ~/.config directory and dot.yaml
	configDir := filepath.Join(tempDir, ".config")
	err = os.MkdirAll(configDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create .config dir: %v", err)
	}

	yamlContent := `
cluster:
  name: "default-path-cluster"
`
	err = os.WriteFile(filepath.Join(configDir, "dot.yaml"), []byte(yamlContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to write dot.yaml: %v", err)
	}

	cfg, err = LoadConfig("")
	if err != nil {
		t.Fatalf("Expected no error loading default config, got %v", err)
	}
	if cfg.Cluster.Name != "default-path-cluster" {
		t.Errorf("Expected cluster 'default-path-cluster', got %q", cfg.Cluster.Name)
	}
}

func TestLoadConfig_UnknownKey(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "dot.yaml")
	// 'naem' is a typo for 'name'; strict decoding must fail loudly instead of silently ignoring it.
	if err := os.WriteFile(configPath, []byte("cluster:\n  naem: oops\n"), 0o600); err != nil {
		t.Fatalf("Failed to write dot.yaml: %v", err)
	}

	if _, err := LoadConfig(configPath); err == nil {
		t.Error("Expected error for an unknown config key, got nil")
	}
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "dot.yaml")
	if err := os.WriteFile(configPath, []byte(""), 0o600); err != nil {
		t.Fatalf("Failed to write dot.yaml: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Expected no error for an empty config file, got %v", err)
	}
	if cfg.Cluster.Name != "local" {
		t.Errorf("Expected defaults for an empty config, got %q", cfg.Cluster.Name)
	}
}

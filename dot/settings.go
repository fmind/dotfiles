package dot

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the unified configuration structure for the dot CLI.
type Config struct {
	Cluster      ClusterConfig      `yaml:"cluster"`
	AI           AIConfig           `yaml:"ai"`
	Verify       VerifyConfig       `yaml:"verify"`
	Login        LoginConfig        `yaml:"login"`
	PR           PRConfig           `yaml:"pr"`
	Completions  CompletionConfig   `yaml:"completions"`
	ChezmoiClean ChezmoiCleanConfig `yaml:"chezmoi_clean"`
	Pull         PullConfig         `yaml:"pull"`
	Setup        SetupConfig        `yaml:"setup"`
	Commit       CommitConfig       `yaml:"commit"`
}

// ExpandPath replaces a leading "~" or "~/" (or "~\") with the user's home directory.
// A "~username" form is intentionally left untouched: resolving it as $HOME/username
// would silently fabricate an incorrect path, and dot does not support other users' homes.
func ExpandPath(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return path
	}
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// DefaultConfig returns the fallback configuration when dot.yaml is not found.
func DefaultConfig() *Config {
	return &Config{
		Completions:  defaultCompletionConfig(),
		Cluster:      defaultClusterConfig(),
		PR:           defaultPRConfig(),
		AI:           defaultAIConfig(),
		Verify:       defaultVerifyConfig(),
		Login:        defaultLoginConfig(),
		Pull:         defaultPullConfig(),
		Setup:        defaultSetupConfig(),
		Commit:       defaultCommitConfig(),
		ChezmoiClean: defaultChezmoiCleanConfig(),
	}
}

// ConfigFilePath resolves the path LoadConfig reads for the given flag value.
// It returns the resolved path and whether it is the implicit default location
// (an empty flag), which callers use to decide whether a missing file is an error.
// An empty path resolves to ~/.config/dot.yaml; any other value is ~-expanded.
func ConfigFilePath(path string) (resolved string, isDefault bool, err error) {
	if path == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return "", true, fmt.Errorf("failed to get home directory: %w", homeErr)
		}
		return filepath.Join(home, ".config", "dot.yaml"), true, nil
	}
	return ExpandPath(path), false, nil
}

// LoadConfig reads and parses the config from the specified path.
// If path is empty, it defaults to ~/.config/dot.yaml.
// If the config file does not exist at the default path, it returns the DefaultConfig without error.
// If the config file is explicitly specified but missing or invalid, it returns the DefaultConfig and the error.
func LoadConfig(path string) (*Config, error) {
	resolved, isDefault, err := ConfigFilePath(path)
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		if isDefault && os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), fmt.Errorf("failed to read config file at %s: %w", resolved, err)
	}

	// Strict decoding so a misspelled key in a hand-authored config fails loudly instead
	// of silently leaving the default in effect. An empty file (io.EOF) keeps the defaults.
	//
	// Precedence note: decoding onto a populated DefaultConfig() means yaml.v3 REPLACES
	// slice fields (e.g. Verify.Tools, Login.*Scopes) but MERGES into map fields
	// (Completions.CustomCommands) — so a user can override/add a completion command but
	// cannot remove a built-in one. This asymmetry is intentional; keep it in mind.
	cfg := DefaultConfig()
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(cfg); err != nil && !errors.Is(err, io.EOF) {
		return DefaultConfig(), fmt.Errorf("failed to parse config file at %s: %w", resolved, err)
	}
	return cfg, nil
}

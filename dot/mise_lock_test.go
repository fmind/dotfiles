package dot

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestMiseLockedInstallsResolveForSupportedPlatforms(t *testing.T) {
	mise, err := exec.LookPath("mise")
	if err != nil {
		t.Fatal("mise must be installed to verify its lock semantics")
	}
	repo := repositoryRoot(t)
	for _, test := range []struct {
		name   string
		config string
		lock   string
	}{
		{name: "repository", config: "mise.toml", lock: "mise.lock"},
		{name: "global", config: "dot_config/mise/config.toml.tmpl", lock: "dot_config/mise/mise.lock"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := t.TempDir()
			copyMiseTestFile(t, filepath.Join(repo, test.config), filepath.Join(fixture, "mise.toml"))
			copyMiseTestFile(t, filepath.Join(repo, test.lock), filepath.Join(fixture, "mise.lock"))
			emptyGlobal := filepath.Join(fixture, "empty-global.toml")
			if err := os.WriteFile(emptyGlobal, []byte("# Isolate the lock under test.\n"), 0o600); err != nil {
				t.Fatal(err)
			}

			for _, platform := range readDeclaredLockPlatforms(t, filepath.Join(fixture, "mise.toml")) {
				t.Run(platform, func(t *testing.T) {
					osName, arch, ok := strings.Cut(platform, "-")
					if !ok {
						t.Fatalf("invalid mise platform %q", platform)
					}
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					cmd := exec.CommandContext(ctx, mise, "-C", fixture, "--locked", "install", "--dry-run", "--quiet", "-y")
					cmd.Env = isolatedMiseEnvironment(fixture, emptyGlobal, osName, arch)
					output, err := cmd.CombinedOutput()
					if err != nil {
						t.Fatalf("locked install does not resolve for %s: %v\n%s", platform, err, output)
					}
				})
			}
		})
	}
}

func copyMiseTestFile(t *testing.T, source, target string) {
	t.Helper()
	content, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, content, 0o600); err != nil {
		t.Fatal(err)
	}
}

func isolatedMiseEnvironment(root, global, osName, arch string) []string {
	environment := make([]string, 0, len(os.Environ())+7)
	for _, value := range os.Environ() {
		name, _, _ := strings.Cut(value, "=")
		if !strings.HasPrefix(name, "MISE_") {
			environment = append(environment, value)
		}
	}
	return append(environment,
		"MISE_GLOBAL_CONFIG_FILE="+global,
		"MISE_DATA_DIR="+filepath.Join(root, "data-"+osName+"-"+arch),
		"MISE_CACHE_DIR="+filepath.Join(root, "cache"),
		"MISE_STATE_DIR="+filepath.Join(root, "state"),
		"MISE_OS="+osName,
		"MISE_ARCH="+arch,
		"MISE_YES=1",
	)
}

func readDeclaredLockPlatforms(t *testing.T, path string) []string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	setting := regexp.MustCompile(`(?m)^lockfile_platforms\s*=\s*\[([^]]+)\]$`).FindSubmatch(content)
	if len(setting) != 2 {
		t.Fatalf("%s must declare one lockfile_platforms setting", path)
	}
	matches := regexp.MustCompile(`"([a-z0-9-]+)"`).FindAllSubmatch(setting[1], -1)
	platforms := make([]string, 0, len(matches))
	for _, match := range matches {
		platforms = append(platforms, string(match[1]))
	}
	if len(platforms) == 0 {
		t.Fatalf("%s declares no lockfile platforms", path)
	}
	return platforms
}

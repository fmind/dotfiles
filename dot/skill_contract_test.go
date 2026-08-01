package dot

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

const maxSkillLines = 500

type skillManifest struct {
	Skills  map[string][]string `json:"skills"`
	Version int                 `json:"version"`
}

var (
	localLinkPattern = regexp.MustCompile(`\[[^]]+\]\(([^)]+)\)`)
	toolNamePattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9+.-]*$`)
)

func TestSkillContracts(t *testing.T) {
	repo := skillRepositoryRoot(t)
	manifest := readSkillManifest(t, repo)
	if manifest.Version != 1 {
		t.Fatalf("skills/contracts.json: unsupported version %d; replace with 1", manifest.Version)
	}

	skillFiles, err := filepath.Glob(filepath.Join(repo, "skills", "*", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]struct{}, len(skillFiles))
	for _, skillFile := range skillFiles {
		name := filepath.Base(filepath.Dir(skillFile))
		seen[name] = struct{}{}
		tools, ok := manifest.Skills[name]
		if !ok {
			t.Errorf("skills/contracts.json: missing required-tools declaration for skill %q; add an explicit array, including [] when no CLI is required", name)
			continue
		}
		checkSkillContract(t, repo, skillFile, tools)
	}
	for name := range manifest.Skills {
		if _, ok := seen[name]; !ok {
			t.Errorf("skills/contracts.json: unknown skill %q; remove it or add skills/%s/SKILL.md", name, name)
		}
	}
}

func skillRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve skill contract source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(source), ".."))
}

func readSkillManifest(t *testing.T, repo string) skillManifest {
	t.Helper()
	path := filepath.Join(repo, "skills", "contracts.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var manifest skillManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("skills/contracts.json: decode: %v", err)
	}
	return manifest
}

func checkSkillContract(t *testing.T, repo, skillFile string, tools []string) {
	t.Helper()
	data, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatal(err)
	}
	relativeSkill, err := filepath.Rel(repo, skillFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := bytes.Count(data, []byte("\n")) + 1; lines > maxSkillLines {
		t.Errorf("%s: %d lines exceeds the %d-line progressive-disclosure limit; move detail into a referenced resource", relativeSkill, lines, maxSkillLines)
	}

	lowerContent := strings.ToLower(string(data))
	for _, tool := range tools {
		if !toolNamePattern.MatchString(tool) {
			t.Errorf("skills/contracts.json: skill %q has invalid required tool %q; use its executable name", filepath.Base(filepath.Dir(skillFile)), tool)
		}
		if !strings.Contains(lowerContent, strings.ToLower(tool)) {
			t.Errorf("%s: required tool %q is undocumented; name the prerequisite or remove it from skills/contracts.json", relativeSkill, tool)
		}
	}

	checkSkillLinks(t, repo, skillFile, string(data))
	checkSkillResources(t, repo, skillFile, string(data))
}

func checkSkillLinks(t *testing.T, repo, skillFile, content string) {
	t.Helper()
	for _, match := range localLinkPattern.FindAllStringSubmatch(content, -1) {
		target := strings.Trim(match[1], "<>")
		if target == "" || strings.HasPrefix(target, "#") || strings.HasPrefix(target, "~") || strings.Contains(target, "://") {
			continue
		}
		target = strings.SplitN(target, "#", 2)[0]
		if _, err := os.Stat(filepath.Clean(filepath.Join(filepath.Dir(skillFile), target))); err != nil {
			relativeSkill, relErr := filepath.Rel(repo, skillFile)
			if relErr != nil {
				relativeSkill = skillFile
			}
			t.Errorf("%s: missing referenced resource %q; replace it with an existing relative path", relativeSkill, target)
		}
	}
}

func checkSkillResources(t *testing.T, repo, skillFile, content string) {
	t.Helper()
	allowedDirectories := []string{"agents", "assets", "references", "resources", "scripts", "templates", "tests"}
	skillDirectory := filepath.Dir(skillFile)
	err := filepath.WalkDir(skillDirectory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == skillDirectory || path == skillFile {
			return nil
		}
		relative, relErr := filepath.Rel(skillDirectory, path)
		if relErr != nil {
			return relErr
		}
		first, _, _ := strings.Cut(filepath.ToSlash(relative), "/")
		if !slices.Contains(allowedDirectories, first) {
			return &skillLayoutError{path: path}
		}
		if !entry.IsDir() && !strings.Contains(content, filepath.ToSlash(relative)) {
			relativeSkill, repoErr := filepath.Rel(repo, skillFile)
			if repoErr != nil {
				relativeSkill = skillFile
			}
			t.Errorf("%s: unreferenced resource %q; link it from SKILL.md or remove it", relativeSkill, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		var layoutErr *skillLayoutError
		if errors.As(err, &layoutErr) {
			relative, relErr := filepath.Rel(repo, layoutErr.path)
			if relErr != nil {
				relative = layoutErr.path
			}
			t.Errorf("%s: unsupported progressive-disclosure path; move it under agents, assets, references, resources, scripts, templates, or tests", relative)
			return
		}
		t.Fatal(err)
	}
}

type skillLayoutError struct {
	path string
}

func (err *skillLayoutError) Error() string {
	return err.path
}

func TestHighRiskSkillSmokeContracts(t *testing.T) {
	repo := skillRepositoryRoot(t)
	tests := []struct {
		name     string
		path     string
		required []string
		ordered  []string
		commands [][]string
	}{
		{
			name:     "release gates publication",
			path:     "skills/release/SKILL.md",
			ordered:  []string{"Clean working tree", "git tag -a", "git push --follow-tags", "gh release create"},
			commands: [][]string{{"release"}, {"r"}},
		},
		{
			name:     "kubernetes verifies and stops local state",
			path:     "skills/k8s-local/SKILL.md",
			required: []string{"OFF by Default", "docker info", "--kubeconfig ~/.kube/dot/local.yaml --context k3d-local", "dot cluster start", "dot cluster stop"},
			commands: [][]string{{"cluster", "start"}, {"k", "s"}, {"cluster", "stop"}},
		},
		{
			name:    "git publication validates before push",
			path:    "skills/git-add-commit-push/SKILL.md",
			ordered: []string{"git status --short", "git commit -m", "git push", "Do not use `--no-verify`"},
		},
		{
			name:     "repository review preserves proof boundaries",
			path:     "skills/repository-review/SKILL.md",
			required: []string{"staged, unstaged, and untracked", "isolated temporary worktree", "mise run all", "source-ready", "local-green", "exact-head-CI", "runtime-proven", "deployed", "release-published", "Key findings", "Actions"},
		},
	}
	app := NewApp()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(repo, test.path))
			if err != nil {
				t.Fatal(err)
			}
			content := string(data)
			for _, required := range test.required {
				if !strings.Contains(content, required) {
					t.Errorf("%s: missing offline smoke invariant %q; restore the documented safety step", test.path, required)
				}
			}
			previous := -1
			for _, required := range test.ordered {
				position := strings.Index(content, required)
				if position < 0 {
					t.Errorf("%s: missing offline smoke invariant %q; restore the documented safety step", test.path, required)
					continue
				}
				if position <= previous {
					t.Errorf("%s: safety step %q is out of order; keep validation before mutation and publication", test.path, required)
				}
				previous = position
			}
			for _, commandPath := range test.commands {
				if !hasCommandPath(app, commandPath) {
					t.Errorf("%s: documented dot command %q is unavailable; replace it with live command metadata", test.path, strings.Join(commandPath, " "))
				}
			}
		})
	}
}

func hasCommandPath(root *cli.Command, path []string) bool {
	current := root
	for _, part := range path {
		var found *cli.Command
		for _, command := range current.Commands {
			if command.Name == part || slices.Contains(command.Aliases, part) {
				found = command
				break
			}
		}
		if found == nil {
			return false
		}
		current = found
	}
	return true
}

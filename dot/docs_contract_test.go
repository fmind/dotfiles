package dot

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

type documentedTask struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}

var (
	miseRunPattern = regexp.MustCompile(`mise run ([a-zA-Z0-9_-]+(?::[a-zA-Z0-9_-]+)*)`)
	dotRunPattern  = regexp.MustCompile("`dot(?: ([^`]+))?`")
	linkPattern    = regexp.MustCompile(`\[[^]]+\]\(([^)]+)\)`)
	layoutPattern  = regexp.MustCompile("^- `([^`]+)`")
	aliasPattern   = regexp.MustCompile("`([a-z]+)` ([a-z][a-z:]*)")
)

var repositoryCommandFiles = map[string]struct{}{
	"AGENTS.md":                           {},
	".agents/skills/chezmoi/SKILL.md":     {},
	"skills/dot-cli/SKILL.md":             {},
	".agents/skills/dot-release/SKILL.md": {},
}

var workspaceLayoutPaths = map[string]struct{}{
	// Agent CLIs create these ignored state directories; a clean clone must not need them.
	".agents":         {},
	".antigravitycli": {},
	".claude":         {},
	".gemini":         {},
}

func TestDocumentationContract(t *testing.T) {
	repo := repositoryRoot(t)
	tasks, aliases := liveMiseTasks(t, repo)
	commands := commandPaths(NewApp())
	for _, path := range documentationFiles(t, repo) {
		_, checkTasks := repositoryCommandFiles[relative(repo, path)]
		checkDocumentationFile(t, repo, path, tasks, aliases, commands, checkTasks)
	}
	checkLayoutInventory(t, repo)
	checkCommandInventory(t, repo)
}

func TestCIUsesCanonicalGate(t *testing.T) {
	repo := repositoryRoot(t)
	for _, relativePath := range []string{
		".github/workflows/ci.yml",
		"skills/github-actions/references/ci.yml",
	} {
		content, err := os.ReadFile(filepath.Join(repo, relativePath))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "run: mise run all") {
			t.Errorf("%s: CI must delegate to the canonical `mise run all` gate", relativePath)
		}
	}
}

// checkCommandInventory asserts the reverse of checkDocumentationFile: that file
// only proves every documented command exists, which silently tolerates a shipped
// command nobody wrote down. Every top-level command must appear in the dot-cli
// skill, and every one must carry an alias, because that skill promises both.
func checkCommandInventory(t *testing.T, repo string) {
	t.Helper()
	path := filepath.Join(repo, "skills", "dot-cli", "SKILL.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	documented := string(content)
	for _, command := range NewApp().Commands {
		if command.Name == "help" {
			continue
		}
		if !strings.Contains(documented, "`dot "+command.Name+"`") {
			t.Errorf("command %q is not documented in skills/dot-cli/SKILL.md", command.Name)
		}
		if len(command.Aliases) == 0 {
			t.Errorf("command %q has no alias, but skills/dot-cli/SKILL.md promises one for every command", command.Name)
		}
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	return skillRepositoryRoot(t)
}

func liveMiseTasks(t *testing.T, repo string) (map[string]struct{}, map[string]string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "mise", "tasks", "--json")
	cmd.Dir = repo
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("mise tasks --json: %v", err)
	}
	var metadata []documentedTask
	if err := json.Unmarshal(output, &metadata); err != nil {
		t.Fatalf("decode mise tasks: %v", err)
	}
	tasks := make(map[string]struct{}, len(metadata))
	aliases := make(map[string]string)
	for _, task := range metadata {
		tasks[task.Name] = struct{}{}
		for _, alias := range task.Aliases {
			// mise resolves a duplicated alias to a single arbitrary task instead of
			// erroring, so the shadowed task becomes unreachable from its documented
			// shortcut. Fail here: every other assertion below reads this map and would
			// otherwise validate the winner while the loser silently never runs.
			if owner, taken := aliases[alias]; taken {
				t.Errorf("alias %q is claimed by both %q and %q; aliases must be unique", alias, owner, task.Name)
			}
			aliases[alias] = task.Name
		}
	}
	return tasks, aliases
}

func commandPaths(root *cli.Command) map[string]struct{} {
	paths := map[string]struct{}{"": {}}
	var visit func([]string, []*cli.Command)
	visit = func(prefixes []string, commands []*cli.Command) {
		for _, command := range commands {
			names := append([]string{command.Name}, command.Aliases...)
			nextPrefixes := make([]string, 0, len(prefixes)*len(names))
			for _, prefix := range prefixes {
				for _, name := range names {
					path := strings.TrimSpace(prefix + " " + name)
					paths[path] = struct{}{}
					nextPrefixes = append(nextPrefixes, path)
				}
			}
			visit(nextPrefixes, command.Commands)
		}
	}
	visit([]string{""}, root.Commands)
	return paths
}

func documentationFiles(t *testing.T, repo string) []string {
	t.Helper()
	skills, err := filepath.Glob(filepath.Join(repo, "skills", "*", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	agentSkills, err := filepath.Glob(filepath.Join(repo, ".agents", "skills", "*", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	files := make([]string, 0, 1+len(skills)+len(agentSkills))
	files = append(files, filepath.Join(repo, "AGENTS.md"))
	files = append(files, skills...)
	return append(files, agentSkills...)
}

func checkDocumentationFile(t *testing.T, repo, path string, tasks map[string]struct{}, aliases map[string]string, commands map[string]struct{}, checkTasks bool) {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Errorf("close %s: %v", relative(repo, path), err)
		}
	})

	scanner := bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		text := scanner.Text()
		for _, match := range miseRunPattern.FindAllStringSubmatch(text, -1) {
			if _, ok := tasks[match[1]]; checkTasks && !ok {
				t.Errorf("%s:%d: missing mise task %q; replace with %q", relative(repo, path), line, match[1], nearest(match[1], mapKeys(tasks)))
			}
		}
		for _, match := range dotRunPattern.FindAllStringSubmatch(text, -1) {
			parts := strings.Fields(match[1])
			if len(parts) == 0 || strings.HasPrefix(parts[0], "<") {
				continue
			}
			command := parts[0]
			if _, ok := commands[command]; !ok {
				t.Errorf("%s:%d: missing dot command %q; replace with %q", relative(repo, path), line, command, nearest(command, mapKeys(commands)))
				continue
			}
			for _, part := range parts[1:] {
				if strings.HasPrefix(part, "-") || strings.HasPrefix(part, "<") {
					break
				}
				candidate := command + " " + part
				if _, ok := commands[candidate]; !ok {
					break // The remaining words are positional arguments, not command metadata.
				}
				command = candidate
			}
		}
		for _, match := range linkPattern.FindAllStringSubmatch(text, -1) {
			target := strings.Trim(match[1], "<>")
			if target == "" || strings.HasPrefix(target, "#") || strings.Contains(target, "://") || strings.HasPrefix(target, "~") {
				continue
			}
			target = strings.SplitN(target, "#", 2)[0]
			if _, err := os.Stat(filepath.Clean(filepath.Join(filepath.Dir(path), target))); err != nil {
				t.Errorf("%s:%d: missing relative link %q; replace it with an existing repository path", relative(repo, path), line, target)
			}
		}
		if path == filepath.Join(repo, "AGENTS.md") && strings.Contains(text, "tasks take") {
			for _, match := range aliasPattern.FindAllStringSubmatch(text, -1) {
				if aliases[match[1]] != match[2] {
					t.Errorf("AGENTS.md:%d: alias %q does not resolve to %q; replace it with the live task mapping", line, match[1], match[2])
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}

func checkLayoutInventory(t *testing.T, repo string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repo, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	inLayout := false
	documented := make(map[string]struct{})
	for lineNumber, line := range strings.Split(string(data), "\n") {
		if line == "## Layout" {
			inLayout = true
			continue
		}
		if inLayout && strings.HasPrefix(line, "## ") {
			break
		}
		match := layoutPattern.FindStringSubmatch(line)
		if !inLayout || match == nil {
			continue
		}
		target := strings.TrimSuffix(match[1], "/")
		documented[target] = struct{}{}
		if _, workspaceState := workspaceLayoutPaths[target]; workspaceState {
			continue
		}
		if _, statErr := os.Stat(filepath.Join(repo, target)); statErr != nil {
			t.Errorf("AGENTS.md:%d: missing layout path %q; remove it or replace it with an existing top-level path", lineNumber+1, match[1])
		}
	}
	cmd := exec.CommandContext(context.Background(), "git", "ls-files")
	cmd.Dir = repo
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	for _, path := range strings.Fields(string(output)) {
		top := strings.SplitN(path, "/", 2)[0]
		if _, ok := documented[top]; !ok {
			t.Errorf("AGENTS.md: missing tracked top-level path %q from Layout; add it to the inventory", top)
		}
	}
}

func nearest(want string, candidates []string) string {
	slices.Sort(candidates)
	for _, candidate := range candidates {
		if strings.HasPrefix(candidate, strings.SplitN(want, ":", 2)[0]) {
			return candidate
		}
	}
	if len(candidates) == 0 {
		return "an available command"
	}
	return candidates[0]
}

func mapKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func relative(repo, path string) string {
	result, err := filepath.Rel(repo, path)
	if err != nil {
		return path
	}
	return result
}

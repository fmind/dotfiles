package dot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v3"
)

func TestBuildContextMarkdownIsDeterministicAndBounded(t *testing.T) {
	root := contextFixture(t)
	state := contextTestState(root)
	state.Config.Context.FailureFiles = []string{"failure.log"}
	options := ContextOptions{Format: "markdown", Bytes: 4000, GeneratedAt: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)}

	first, err := BuildContext(context.Background(), state, options)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildContext(context.Background(), state, options)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("fixed inputs must produce byte-identical context")
	}
	if len(first) > options.Bytes {
		t.Fatalf("context is %d bytes, budget is %d", len(first), options.Bytes)
	}
	for _, want := range []string{"Schema: 1.0", "Repository: project", "instructions |", "skills |", "git |", "tasks |", "dependencies |", "failures |", "fingerprint=sha256:", "observed=2026-08-01T12:00:00Z", "session-metadata (unavailable in project-only v1)"} {
		if !strings.Contains(first, want) {
			t.Errorf("context missing %q:\n%s", want, first)
		}
	}
	if strings.Contains(first, filepath.Dir(root)) {
		t.Fatalf("context leaked the host path %q", filepath.Dir(root))
	}
}

func TestBuildContextJSONUsesTokenBudgetAndMarksOmissions(t *testing.T) {
	root := contextFixture(t)
	state := contextTestState(root)
	state.Config.Context.Collectors = []string{"instructions", "tasks"}
	options := ContextOptions{Format: "json", Tokens: 400, GeneratedAt: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)}
	payload, err := BuildContext(context.Background(), state, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(payload) > options.Tokens*bytesPerToken {
		t.Fatalf("context is %d bytes, effective budget is %d", len(payload), options.Tokens*bytesPerToken)
	}
	var envelope contextEnvelope
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, payload)
	}
	if envelope.SchemaVersion != contextSchemaVersion || envelope.Budget.RequestedTokens != options.Tokens {
		t.Fatalf("unexpected envelope: %+v", envelope)
	}
	omitted := false
	for _, section := range envelope.Sections {
		if section.OmittedBytes > 0 && len(section.OmittedLines) > 0 {
			omitted = true
		}
	}
	if !omitted {
		t.Fatalf("small budget did not report omissions: %s", payload)
	}
}

func TestBuildContextReportsUnavailableAndUnsafeSources(t *testing.T) {
	root := contextFixture(t)
	externalDir := filepath.Join(filepath.Dir(root), "private")
	if err := os.MkdirAll(externalDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(externalDir, "secret.txt"), []byte("private"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(externalDir, filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	state := contextTestState(root)
	state.Config.Context.Collectors = []string{"tasks", "instructions", "unknown"}
	state.Config.Context.InstructionFiles = []string{"../outside", "missing.md", "linked/secret.txt"}
	runner, ok := state.Runner.(*FakeRunner)
	if !ok {
		t.Fatal("context test state does not use FakeRunner")
	}
	runner.LookPathFunc = func(name string) (string, error) {
		if name == "mise" {
			return "", errors.New("not installed")
		}
		return "/usr/bin/" + name, nil
	}
	payload, err := BuildContext(context.Background(), state, ContextOptions{Format: "json", Bytes: 5000, GeneratedAt: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"mise unavailable", "path escapes project root", "source does not exist", "source resolves outside project root", "collector is not allowlisted"} {
		if !strings.Contains(payload, want) {
			t.Errorf("context missing error %q: %s", want, payload)
		}
	}
	if strings.Contains(payload, filepath.Dir(root)) {
		t.Fatalf("error reporting leaked the host path %q", filepath.Dir(root))
	}
}

func TestRunContextScansExactFinalPayload(t *testing.T) {
	root := contextFixture(t)
	state := contextTestState(root)
	var scanned string
	baseRunner, ok := state.Runner.(*FakeRunner)
	if !ok {
		t.Fatal("context test state does not use FakeRunner")
	}
	originalRun := baseRunner.RunFunc
	baseRunner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
		if name == "/usr/bin/gitleaks" {
			content, readErr := io.ReadAll(stdin)
			if readErr != nil {
				t.Fatal(readErr)
			}
			scanned = string(content)
			return "", nil
		}
		return originalRun(ctx, dir, stdin, name, args...)
	}
	var output bytes.Buffer
	state.Stdout = &output
	options := ContextOptions{Format: "markdown", Bytes: 4000, GeneratedAt: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)}
	if err := RunContext(context.Background(), state, options); err != nil {
		t.Fatal(err)
	}
	if scanned == "" || scanned != output.String() {
		t.Fatal("printed context differs from the exact payload scanned by gitleaks")
	}
}

func TestScanContextPayloadRejectsConfiguredPatterns(t *testing.T) {
	state := newTestState(&FakeRunner{RunFunc: func(_ context.Context, _ string, _ io.Reader, name string, _ ...string) (string, error) {
		if name == "/usr/bin/gitleaks" {
			return "", nil
		}
		return "", nil
	}})
	state.Config.Context.SensitivePathPatterns = []string{"private/customer"}
	if err := ScanContextPayload(context.Background(), state, "private/customer/data"); err == nil {
		t.Fatal("configured sensitive path must fail closed")
	}
	t.Setenv("CONTEXT_TEST_SECRET", "very-private-value")
	state.Config.Context.SensitivePathPatterns = nil
	state.Config.Context.SensitiveEnvPatterns = []string{"CONTEXT_TEST_*"}
	if err := ScanContextPayload(context.Background(), state, "value=very-private-value"); err == nil {
		t.Fatal("configured environment value must fail closed")
	}
}

func TestContextBudgetValidation(t *testing.T) {
	config := defaultContextConfig()
	for _, options := range []ContextOptions{{Bytes: -1}, {Tokens: -1}, {Bytes: 1, Tokens: 1}} {
		if _, err := resolveContextBudget(config, options); err == nil {
			t.Fatalf("invalid budget accepted: %+v", options)
		}
	}
	if _, err := resolveContextBudget(config, ContextOptions{Tokens: int(^uint(0)>>1)/bytesPerToken + 1}); err == nil {
		t.Fatal("overflowing token budget must fail")
	}
	if _, err := BuildContext(context.Background(), contextTestState(contextFixture(t)), ContextOptions{Format: "yaml", Bytes: 1000, GeneratedAt: time.Now()}); err == nil {
		t.Fatal("unsupported format must fail")
	}
	if _, err := BuildContext(context.Background(), contextTestState(contextFixture(t)), ContextOptions{Format: "json", Bytes: 1, GeneratedAt: time.Now()}); err == nil || !strings.Contains(err.Error(), "omission manifest") {
		t.Fatalf("budget below the manifest must fail, got %v", err)
	}
}

func TestRedactProjectRootStripsAbsolutePathsFromErrors(t *testing.T) {
	// Collector errors are the one place a raw absolute path can reach the pack, and
	// the pack is meant to be shareable — so the project root must never survive.
	root := filepath.Join("/home", "someone", "work", "project")
	cases := []struct {
		name  string
		value string
		want  string
	}{
		{name: "leaves unrelated text", value: "open go.mod: no such file", want: "open go.mod: no such file"},
		{name: "redacts native separators", value: "open " + filepath.Join(root, "go.mod") + ": denied", want: "open ." + string(filepath.Separator) + "go.mod: denied"},
		{name: "redacts every occurrence", value: root + " and " + root, want: ". and ."},
		// Tool output often reports slash-separated paths regardless of platform, so
		// both spellings of the root have to be covered.
		{name: "redacts slash form", value: "open " + filepath.ToSlash(root) + "/go.mod", want: "open ./go.mod"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := redactProjectRoot(testCase.value, root)
			if got != testCase.want {
				t.Fatalf("redactProjectRoot(%q) = %q, want %q", testCase.value, got, testCase.want)
			}
			if strings.Contains(got, root) {
				t.Fatalf("redactProjectRoot leaked the project root: %q", got)
			}
		})
	}
}

func TestContextCommandFlags(t *testing.T) {
	root := contextFixture(t)
	state := contextTestState(root)
	var output bytes.Buffer
	state.Stdout = &output
	runner, ok := state.Runner.(*FakeRunner)
	if !ok {
		t.Fatal("context test state does not use FakeRunner")
	}
	originalRun := runner.RunFunc
	runner.RunFunc = func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
		if name == "/usr/bin/gitleaks" {
			return "", nil
		}
		return originalRun(ctx, dir, stdin, name, args...)
	}
	app := &cli.Command{Commands: []*cli.Command{NewContextCmd(state)}}
	if err := app.Run(context.Background(), []string{"dot", "context", "--bytes", "4000", "--format", "json"}); err != nil {
		t.Fatal(err)
	}
	var envelope contextEnvelope
	if err := json.Unmarshal(output.Bytes(), &envelope); err != nil {
		t.Fatalf("command emitted invalid JSON: %v", err)
	}
	if envelope.Budget.RequestedBytes != 4000 {
		t.Fatalf("command did not forward byte budget: %+v", envelope.Budget)
	}
}

func contextFixture(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "project")
	for _, dir := range []string{root, filepath.Join(root, "skills", "sample")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"AGENTS.md":              "# Instructions\n" + strings.Repeat("Keep evidence explicit.\n", 20),
		"skills/sample/SKILL.md": "---\nname: sample\ndescription: Sample project skill.\n---\n# Sample\n",
		"go.mod":                 "module example.com/project\n\ngo 1.25\n",
		"failure.log":            "test: expected true, got false\n",
	}
	for path, content := range files {
		fullPath := filepath.Join(root, path)
		if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func contextTestState(root string) *GlobalState {
	runner := &FakeRunner{RunFunc: func(_ context.Context, dir string, _ io.Reader, name string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		switch {
		case name == "git" && joined == "rev-parse --show-toplevel":
			return root + "\n", nil
		case name == "git" && joined == "rev-parse HEAD":
			return strings.Repeat("a", 40) + "\n", nil
		case name == "git" && joined == "status --short --branch":
			return "## main\n M AGENTS.md\n", nil
		case name == "git" && strings.HasPrefix(joined, "log -5"):
			return "abc1234 feat: example", nil
		case name == "git" && joined == "ls-files -z":
			return "AGENTS.md\x00go.mod\x00skills/sample/SKILL.md\x00", nil
		case name == "mise" && joined == "tasks --json":
			return `[{"name":"test","source":"` + root + `/mise.toml"},{"name":"check","aliases":["c"],"description":"Check source","dir":"` + root + `"}]`, nil
		default:
			return "", errors.New("unexpected command: " + name + " " + joined + " in " + dir)
		}
	}}
	return newTestState(runner)
}

package dot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestDiagnosticSanitizerRedactsKnownSecrets(t *testing.T) {
	sanitizer, err := newDiagnosticSanitizer([]string{`project-[0-9]+`})
	if err != nil {
		t.Fatal(err)
	}
	input := strings.Join([]string{
		"Authorization: Bearer abc.def.ghi",
		"MY_API_TOKEN=super-secret-value",
		"github=ghp_abcdefghijklmnopqrstuvwxyz123456",
		"jwt=eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjMifQ.signature_value",
		"tenant-id=private-customer",
		"project-4821",
		"-----BEGIN PRIVATE KEY-----",
		"multiline-private-material",
		"-----END PRIVATE KEY-----",
	}, "\n")
	got := sanitizer.sanitize(input)
	for _, secret := range []string{"abc.def.ghi", "super-secret-value", "ghp_", "eyJhbGci", "private-customer", "project-4821", "multiline-private-material"} {
		if strings.Contains(got, secret) {
			t.Fatalf("sanitized output retained %q: %s", secret, got)
		}
	}

	secretObject := "apiVersion: v1\nkind: Secret\ndata:\n  token: c2VjcmV0\n"
	if got := sanitizer.sanitize(secretObject); got != "[REDACTED KUBERNETES SECRET]" {
		t.Fatalf("Secret object was not replaced: %q", got)
	}
}

func TestDiagnosticSanitizerRejectsInvalidProjectPattern(t *testing.T) {
	if _, err := newDiagnosticSanitizer([]string{"["}); err == nil || !strings.Contains(err.Error(), "invalid diagnostic redaction pattern") {
		t.Fatalf("expected invalid pattern error, got %v", err)
	}
}

func TestLoadConfigClusterDiagnosticPatterns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dot.yaml")
	if err := os.WriteFile(path, []byte("cluster:\n  diagnostics:\n    redact_patterns:\n      - 'customer-[0-9]+'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(config.Cluster.Diagnostics.RedactPatterns, []string{`customer-[0-9]+`}) {
		t.Fatalf("redact patterns = %v", config.Cluster.Diagnostics.RedactPatterns)
	}
}

func TestBoundDiagnosticTextEnforcesLineByteAndUTF8Limits(t *testing.T) {
	got, truncated := boundDiagnosticText("one\ntwo\nthree\nfour", 3, 100)
	if !truncated || got != "one\ntwo\n[TRUNCATED]" || len(strings.Split(got, "\n")) > 3 {
		t.Fatalf("line bound = %q truncated=%t", got, truncated)
	}

	got, truncated = boundDiagnosticText(strings.Repeat("é", 20), 10, 17)
	if !truncated || len(got) > 17 || !strings.Contains(got, "[TRUNCATED]") {
		t.Fatalf("byte bound = %q bytes=%d truncated=%t", got, len(got), truncated)
	}
}

func TestClusterDiagnosticPlanIsAllowlistedAndBounded(t *testing.T) {
	limits := defaultClusterDiagnosticLimits()
	plan := ClusterDiagnosticPlan("team", limits)
	wantCategories := []string{"versions", "versions", "nodes", "workloads", "conditions", "events", "resources", "resources", "storage", "logs"}
	if len(plan) != len(wantCategories) {
		t.Fatalf("plan length = %d, want %d", len(plan), len(wantCategories))
	}
	for index, probe := range plan {
		if probe.Category != wantCategories[index] || probe.Timeout != limits.Timeout || probe.MaxLines != limits.MaxLines || probe.MaxBytes != limits.MaxBytes || probe.TimeWindow == "" {
			t.Fatalf("unbounded or reordered probe %d: %+v", index, probe)
		}
		command := diagnosticCommand(probe)
		if strings.Contains(command, "secret") || strings.Contains(command, "env") || strings.Contains(command, "kubeconfig") {
			t.Fatalf("unsafe probe command: %s", command)
		}
	}
}

func TestRunClusterDiagnosePreservesSanitizedPartialResults(t *testing.T) {
	const namespace = "team"
	limits := ClusterDiagnosticLimits{Timeout: time.Second, TimeWindow: 15 * time.Minute, TailLines: 2, MaxLines: 4, MaxBytes: 96, MaxLogPods: 2}
	authoritative := clusterTestKubeconfig("local", expectedClusterContext("local"), clusterTestServer)
	var logPods []string
	runner := &FakeRunner{
		LookPathFunc: func(command string) (string, error) { return "/bin/" + command, nil },
		RunFunc: func(_ context.Context, _ string, _ io.Reader, command string, args ...string) (string, error) {
			if command == "k3d" {
				if slices.Equal(args, []string{"kubeconfig", "get", "local"}) {
					return authoritative, nil
				}
				if slices.Equal(args, []string{"version"}) {
					return "k3d version v5", nil
				}
			}
			if command != "kubectl" {
				return "", errors.New("unexpected tool")
			}
			if !slices.Contains(args, "--kubeconfig") || !slices.Contains(args, "--context") {
				t.Fatalf("kubectl probe omitted explicit target: %v", args)
			}
			switch {
			case slices.Contains(args, "--raw=/version"):
				return `{"gitVersion":"v1"}`, nil
			case slices.Contains(args, "namespace") && slices.Contains(args, "--ignore-not-found"):
				return "namespace/" + namespace, nil
			case slices.Contains(args, "--containers"):
				return "", errors.New("metrics unavailable MY_TOKEN=do-not-store")
			case slices.Contains(args, "events"):
				return `{"items":[{"eventTime":"2026-08-01T11:50:00Z","metadata":{"namespace":"team"},"involvedObject":{"kind":"Pod","name":"pod-a"},"type":"Warning","reason":"BackOff","message":"MY_TOKEN=event-secret"},{"eventTime":"2026-08-01T10:00:00Z","metadata":{"namespace":"team"},"involvedObject":{"kind":"Pod","name":"old"},"type":"Normal","reason":"Old","message":"outside window"}]}`, nil
			case slices.Contains(args, `jsonpath={range .items[*]}{.metadata.namespace}{"\t"}{.metadata.name}{"\n"}{end}`):
				return "team\tpod-z\nteam\tpod-a\nteam\tpod-b\n", nil
			case slices.Contains(args, "logs"):
				for index, arg := range args {
					if index > 0 && args[index-1] == "logs" {
						logPods = append(logPods, arg)
					}
				}
				return "tenant-id=private\nMY_PASSWORD=hunter2\n" + strings.Repeat("x", 200), nil
			default:
				return "ok", nil
			}
		},
	}
	state := newTestState(runner)
	writeClusterTargetFixture(t, state, expectedClusterContext("local"), clusterTestServer)
	var stdout bytes.Buffer
	state.Stdout = &stdout
	outputPath := filepath.Join(t.TempDir(), "diagnostics.json")
	fixedNow := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)

	err := RunClusterDiagnose(context.Background(), state, clusterDiagnosticOptions{
		Target: ClusterTargetOptions{}, Namespace: namespace, OutputPath: outputPath, Limits: limits, Now: func() time.Time { return fixedNow }, RedactPatterns: []string{`pod-z`},
	})
	if err != nil {
		t.Fatalf("diagnose: %v", err)
	}
	if !slices.Equal(logPods, []string{"pod-a", "pod-b"}) {
		t.Fatalf("bounded deterministic log targets = %v", logPods)
	}
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("manifest permissions = %o, want 600", info.Mode().Perm())
	}
	payload, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"do-not-store", "event-secret", "outside window", "private", "hunter2", clusterTestServer, state.Config.Cluster.KubeconfigPath} {
		if strings.Contains(string(payload), secret) {
			t.Fatalf("manifest retained sensitive target or output %q: %s", secret, payload)
		}
	}
	var manifest ClusterDiagnosticManifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaVersion != ClusterDiagnosticSchemaVersion || !manifest.CollectedAt.Equal(fixedNow) || manifest.Target.Fingerprint == "" {
		t.Fatalf("unexpected manifest metadata: %+v", manifest)
	}
	if len(manifest.Results) != len(ClusterDiagnosticPlan(namespace, limits))+2 {
		t.Fatalf("results = %d", len(manifest.Results))
	}
	partial := 0
	truncated := 0
	for _, result := range manifest.Results {
		if len(result.Output) > limits.MaxBytes || len(result.Error) > limits.MaxBytes {
			t.Fatalf("result exceeded byte bound: %+v", result)
		}
		if result.Status == "error" {
			partial++
		}
		if result.Truncated {
			truncated++
		}
	}
	if partial != 1 || truncated != 2 {
		t.Fatalf("partial=%d truncated=%d", partial, truncated)
	}
	if !strings.Contains(stdout.String(), "1 partial errors") || !strings.Contains(stdout.String(), "Upload: disabled") {
		t.Fatalf("human summary missing evidence: %s", stdout.String())
	}
}

func TestRunClusterDiagnoseFailsBeforeCollectionForMissingNamespace(t *testing.T) {
	authoritative := clusterTestKubeconfig("local", expectedClusterContext("local"), clusterTestServer)
	var collectionCalls int
	runner := &FakeRunner{
		LookPathFunc: func(command string) (string, error) { return "/bin/" + command, nil },
		RunFunc: func(_ context.Context, _ string, _ io.Reader, command string, args ...string) (string, error) {
			switch {
			case command == "k3d":
				return authoritative, nil
			case slices.Contains(args, "--raw=/version"):
				return "version", nil
			case slices.Contains(args, "namespace"):
				return "", nil
			default:
				collectionCalls++
				return "", nil
			}
		},
	}
	state := newTestState(runner)
	writeClusterTargetFixture(t, state, expectedClusterContext("local"), clusterTestServer)
	outputPath := filepath.Join(t.TempDir(), "must-not-exist.json")
	err := RunClusterDiagnose(context.Background(), state, clusterDiagnosticOptions{Namespace: "missing", OutputPath: outputPath, Limits: defaultClusterDiagnosticLimits()})
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected namespace verification error, got %v", err)
	}
	if collectionCalls != 0 {
		t.Fatalf("ran %d collection probes before target verification", collectionCalls)
	}
	if _, statErr := os.Stat(outputPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("unexpected manifest after verification failure: %v", statErr)
	}
}

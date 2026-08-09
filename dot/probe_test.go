package dot

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCapabilityProbeRegistryContract(t *testing.T) {
	registry := CapabilityProbeRegistry()
	for name, probe := range registry {
		if probe.Name != name || probe.Command == "" || len(probe.Args) == 0 {
			t.Errorf("invalid probe %q: %+v", name, probe)
		}
		if probe.Timeout <= 0 || probe.OutputLimit <= 0 {
			t.Errorf("unbounded probe %q: %+v", name, probe)
		}
	}
	for _, name := range []string{"clasp", "gcloud", "gh", "gws", "jules"} {
		if !registry[name].RequiresAuth {
			t.Errorf("probe %q must declare its separate authentication boundary", name)
		}
	}

	delete(registry, "git")
	if _, ok := CapabilityProbeRegistry()["git"]; !ok {
		t.Error("registry callers must receive a defensive copy")
	}
}

func TestHelmCapabilityProbeSupportsCurrentClientOnlyCLI(t *testing.T) {
	probe := CapabilityProbeRegistry()["helm"]
	if !slices.Equal(probe.Args, []string{"version", "--short"}) {
		t.Fatalf("helm probe args = %v, want version --short", probe.Args)
	}
}

func TestStandardRunnerBoundsBrokenProbeOutput(t *testing.T) {
	bin := t.TempDir()
	path := filepath.Join(bin, "broken-probe")
	script := "#!/usr/bin/env bash\nprintf '%0200d' 0 >&2\nexit 1\n"
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	probe := CapabilityProbe{Name: "broken-probe", Command: "broken-probe", Args: []string{"--version"}, Timeout: time.Second, OutputLimit: 64}
	results, passed := runToolProbes(context.Background(), NewStandardRunner(nil, &bytes.Buffer{}, &bytes.Buffer{}), []string{probe.Name}, map[string]CapabilityProbe{probe.Name: probe}, 0)
	if passed || results[0].Condition != ProbeBroken {
		t.Fatalf("result = %+v, passed=%v", results, passed)
	}
	if len(results[0].Details) > probe.OutputLimit+len("…") {
		t.Errorf("failure details exceeded limit: %d > %d", len(results[0].Details), probe.OutputLimit)
	}
	if results[0].Path != path {
		t.Errorf("path = %q, want %q", results[0].Path, path)
	}
}

func TestStandardRunnerTimesOutCapabilityProbe(t *testing.T) {
	bin := t.TempDir()
	path := filepath.Join(bin, "slow-probe")
	if err := os.WriteFile(path, []byte("#!/usr/bin/env bash\nsleep 5\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	probe := CapabilityProbe{Name: "slow-probe", Command: "slow-probe", Args: []string{"--version"}, Timeout: 20 * time.Millisecond, OutputLimit: 64}
	results, passed := runToolProbes(context.Background(), NewStandardRunner(nil, &bytes.Buffer{}, &bytes.Buffer{}), []string{probe.Name}, map[string]CapabilityProbe{probe.Name: probe}, 0)
	if passed || results[0].Condition != ProbeBroken || !strings.Contains(results[0].Details, context.DeadlineExceeded.Error()) {
		t.Fatalf("timeout result = %+v, passed=%v", results, passed)
	}
}

// TestRunToolProbesBoundsConcurrency pins the bound that keeps each probe's timeout
// honest: unbounded fan-out let ~30 concurrent CLI cold starts contend for CPU until
// healthy tools reported "context deadline exceeded".
func TestRunToolProbesBoundsConcurrency(t *testing.T) {
	for _, tc := range []struct {
		name  string
		limit int
		want  int
	}{
		{name: "explicit limit", limit: 3, want: 3},
		{name: "non-positive falls back to default", limit: 0, want: defaultProbeConcurrency},
		{name: "negative falls back to default", limit: -1, want: defaultProbeConcurrency},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var mu sync.Mutex
			inFlight, peak := 0, 0

			runner := &FakeRunner{
				LookPathFunc: func(name string) (string, error) { return "/bin/" + name, nil },
				RunFunc: func(ctx context.Context, dir string, stdin io.Reader, name string, args ...string) (string, error) {
					mu.Lock()
					inFlight++
					peak = max(peak, inFlight)
					mu.Unlock()
					// Hold the slot long enough that every probe allowed to run
					// concurrently is observed in flight at the same time.
					time.Sleep(20 * time.Millisecond)
					mu.Lock()
					inFlight--
					mu.Unlock()
					return "ok", nil
				},
			}

			names := make([]string, 4*defaultProbeConcurrency)
			registry := make(map[string]CapabilityProbe, len(names))
			for i := range names {
				names[i] = fmt.Sprintf("tool-%d", i)
				registry[names[i]] = CapabilityProbe{
					Name: names[i], Command: names[i], Args: []string{"--version"},
					Timeout: time.Minute, OutputLimit: defaultOutputLimit,
				}
			}

			results, passed := runToolProbes(context.Background(), runner, names, registry, tc.limit)
			if !passed || len(results) != len(names) {
				t.Fatalf("passed=%v, results=%d, want true and %d", passed, len(results), len(names))
			}
			for i, result := range results {
				if result.Status != statusPass || result.Name != names[i] {
					t.Fatalf("results[%d] = %+v, want a pass for %q", i, result, names[i])
				}
			}
			if peak > tc.want {
				t.Errorf("peak concurrency = %d, want at most %d", peak, tc.want)
			}
			if peak < 2 {
				t.Errorf("peak concurrency = %d, probes did not run concurrently at all", peak)
			}
		})
	}
}

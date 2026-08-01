package dot

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestStandardRunnerBoundsBrokenProbeOutput(t *testing.T) {
	bin := t.TempDir()
	path := filepath.Join(bin, "broken-probe")
	script := "#!/usr/bin/env bash\nprintf '%0200d' 0 >&2\nexit 1\n"
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	probe := CapabilityProbe{Name: "broken-probe", Command: "broken-probe", Args: []string{"--version"}, Timeout: time.Second, OutputLimit: 64}
	results, passed := runToolProbes(context.Background(), NewStandardRunner(nil, &bytes.Buffer{}, &bytes.Buffer{}), []string{probe.Name}, map[string]CapabilityProbe{probe.Name: probe})
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
	results, passed := runToolProbes(context.Background(), NewStandardRunner(nil, &bytes.Buffer{}, &bytes.Buffer{}), []string{probe.Name}, map[string]CapabilityProbe{probe.Name: probe})
	if passed || results[0].Condition != ProbeBroken || !strings.Contains(results[0].Details, context.DeadlineExceeded.Error()) {
		t.Fatalf("timeout result = %+v, passed=%v", results, passed)
	}
}

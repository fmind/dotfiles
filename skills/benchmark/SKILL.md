---
name: benchmark
description: Measure command latency with hyperfine and HTTP throughput with oha, compare before and after, and report honest numbers. Use for performance comparisons, load tests, or a suspected slowdown.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/benchmark
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Benchmark

Two tools, two questions. `hyperfine` answers "how long does this command take" with warmup, repeated runs, and a comparison; `oha` answers "how does this endpoint behave under load" with latency percentiles and a live TUI. Diagnosing why something is slow belongs to [systematic-debugging](../systematic-debugging/SKILL.md); this skill produces the numbers.

## Commands

```bash
hyperfine --warmup 3 --runs 10 'old-cmd' 'new-cmd'                       # A/B with mean ± σ and a relative speed line
hyperfine --warmup 3 --prepare 'go build ./...' 'go test ./...'          # reset state before each run
hyperfine --parameter-list n 10,100,1000 'tool --items {n}'               # scaling curve
hyperfine --export-markdown bench.md --export-json bench.json 'cmd'       # tables for the PR, raw data for later
oha -z 30s -c 50 --latency-correction http://localhost:8080/health        # 30 s, 50 connections, coordinated-omission safe
oha -n 2000 -c 20 -m POST -H 'Content-Type: application/json' -d '{"q":1}' http://localhost:8080/api
oha --no-tui -z 10s -c 10 --output-format json -o oha.json http://localhost:8080/   # scriptable output for CI or a report
```

## Workflow

1. **Fix the question**: one command or endpoint, one metric (mean latency, p99, requests per second), one hypothesis.
1. **Control the machine**: close heavy processes, run on AC power, and pin versions; record CPU, OS, and tool versions in the report.
1. **Warm up and repeat**: at least 3 warmup runs and 10 measured runs for commands; at least 30 seconds for endpoints. Compare against a baseline measured the same way in the same session.
1. **Read the variance**: a difference smaller than the standard deviation is noise; rerun with more iterations before claiming a win.
1. **Report**: the command lines, the exported table, the relative change, and the conditions. Keep `bench.json` if the number will be tracked over time.

## Gotchas

- **Never load-test a remote service you do not own** or a production system without explicit approval; `oha` at 50 connections is a denial-of-service from the target's point of view.
- **Localhost numbers exclude the network**: `oha` against `localhost` measures the server, not the user experience.
- **Shell startup pollutes short commands**: use `--shell=none` in hyperfine for sub-10 ms commands, or `-N`.
- **Caches lie**: a second run of a build or query hits caches; use `--prepare` to clear them when the cold path is what matters.
- **Cloud Run cold starts**: benchmark with `--min-instances` known, and separate first-request latency from steady state per the [cloud-run skill](../cloud-run/SKILL.md).

## Documentation

- [hyperfine](https://github.com/sharkdp/hyperfine) · [oha](https://github.com/hatoo/oha)
- Companion skills: [quality-assurance](../quality-assurance/SKILL.md) (performance as part of a test campaign), [production-readiness](../production-readiness/SKILL.md) (capacity evidence before promotion).

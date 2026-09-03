---
name: systematic-debugging
description: Diagnose unknown-cause bugs, test/build or auth failures, flakes, and runtime performance regressions. Investigate, reduce, localize, falsify hypotheses, and explain root cause before implementation.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/systematic-debugging
  created: "2026-08-08"
  updated: "2026-09-03"
---

# Systematic Debugging

Replace guess-and-check with an evidence loop that localizes where and why behavior diverges; [test-driven-development](../test-driven-development/SKILL.md) implements the fix and [incident-response](../incident-response/SKILL.md) owns live outages.

## Workflow

1. **Preserve evidence**: Capture the exact error, stack trace, command, inputs, versions, environment differences, timing, and recent changes before touching anything.
1. **Reproduce**: Find the shortest reliable command or sequence; for intermittent failures record the frequency and vary one dimension at a time.
1. **Reduce**: Minimize input, fixture, process count, and component path while keeping the same failure, preferably as a focused test or disposable harness.
1. **Localize**: Trace bad state backward across calls, processes, network boundaries, configuration, and generated artifacts; at each boundary compare what entered with what left.
1. **Find a working comparator**: Locate the nearest known-good test, code path, version, environment, or commit and list every relevant difference before choosing one.
1. **Form one hypothesis**: State `X causes the failure because Y evidence predicts Z observation` and define a minimal probe that could falsify it.
1. **Run the probe**: Change one variable in a reversible fixture or add narrow instrumentation; record whether the prediction held and discard failed hypotheses instead of layering fixes.
1. **Name the root cause**: Explain the triggering condition, the faulty assumption or invariant, the propagation path, and why existing controls missed it; never blame timing, the environment, or a third party until that path and the missing resilience are understood.
1. **Fix only when authorized**: Write a failing regression test, implement the smallest root-cause fix, and verify the symptom plus the wider gate.
1. **Report**: Return symptom and impact, minimal reproduction, evidence and ruled-out hypotheses, root cause and propagation path, the authorized fix or recommended correction, regression proof, and residual uncertainty with the next probe.

## Gotchas

- **Authority**: A request to diagnose authorizes investigation, not implementation; observe read-only, reproduce in an isolated temporary directory, and change product code only when the user also asks for a fix.
- **Thrashing**: After three failed fix attempts or hypotheses that expose different shared-state failures, stop and question the architecture, reproduction, or problem statement with the user.
- **Multi-component pipelines**: Instrument every boundary once with presence, shape, identity, status, timestamps, and correlation ids, never secret values; remove the instrumentation unless it has durable value.
- **Resolver failures**: Record the exact resolver, runtime or toolchain, platform, package index, manifest, lockfile, and installed source before changing any constraint.
- **Resolver reproduction**: Reproduce with the same resolver and distinguish direct constraints, transitive conflicts, platform markers, yanked releases, build-backend or wheel failures, authentication, network reachability, and stale locks.
- **Resolver routing**: Inspect the lock and installed source without executing it and apply the smallest constraint fix only when authorized; route upgrades to [upgrade-tools](../upgrade-tools/SKILL.md), registry facts to [technical-research](../technical-research/SKILL.md), and CVE or license triage to [secure](../secure/SKILL.md).

## Documentation

- Adapted from [Superpowers systematic-debugging](https://github.com/obra/superpowers/blob/44c9b2d6e889982ac18c27d05a19fefe335194e1/skills/systematic-debugging/SKILL.md), [gstack investigate](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/investigate/SKILL.md), [diagnosing-bugs](https://github.com/mattpocock/skills/blob/84fdeffd12f2ee307994d1eb6feb48173b6e0502/skills/engineering/diagnosing-bugs/SKILL.md).
- Companion skills: [test-driven-development](../test-driven-development/SKILL.md) (implement the fix), [repository-history](../repository-history/SKILL.md) (why the code exists), [incident-response](../incident-response/SKILL.md) (live outage), [upgrade-tools](../upgrade-tools/SKILL.md) (deliberate upgrades), [technical-research](../technical-research/SKILL.md) (registry and API facts), [secure](../secure/SKILL.md) (CVE triage).

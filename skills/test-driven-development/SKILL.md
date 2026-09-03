---
name: test-driven-development
description: Implement an isolated bug fix or behavior change with an honest red-green-refactor cycle. Use when a regression test should fail before the fix, or when refactors or seams must prove correctness.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/test-driven-development
  created: "2026-08-08"
  updated: "2026-09-03"
---

# Test-Driven Development

Prove a change with an honest red-green-refactor cycle: a failing test that detects the missing or broken behavior, then the smallest trustworthy change; [quality-assurance](../quality-assurance/SKILL.md) owns the broader campaign and [systematic-debugging](../systematic-debugging/SKILL.md) owns failures not yet understood.

## Workflow

1. **Discover the harness**: Read repository instructions, existing tests, task definitions, and nearby patterns; find the smallest command that exercises the target behavior.
1. **State the contract**: Name the production change that would make the test pass and a plausible regression that would make it fail.
1. **RED**: Write one minimal test for one observable behavior; run it and confirm it fails for the missing behavior, not for a syntax, fixture, environment, or setup error.
1. **GREEN**: Implement only enough production code to satisfy that test; run the focused test and read the full output.
1. **Protect the neighborhood**: Run the package or subsystem tests; fix production code when the new behavior breaks a valid existing contract and revisit the spec when contracts conflict.
1. **REFACTOR**: Improve names, structure, duplication, and types only while everything stays green; add no behavior.
1. **Repeat**: Take the next smallest behavior, edge case, or failure path through a new red cycle.
1. **Prove the regression test**: For a bug fix, temporarily reverse the fix when safe, confirm the test fails, then restore it and confirm green.
1. **Gate the candidate**: Run the full gate (`mise run all`); if the tree carries unrelated changes and the gate write-formats, run it in a temporary `git worktree` or fall back to `mise run check` and `mise run test` (see [mise](../mise/SKILL.md)).
1. **Report evidence**: Give the red command and expected failure, the focused green command, the wider suite, the full gate, and any boundary still covered only manually or not at all.

## Gotchas

- **Never weaken an existing test**: do not loosen a type, add a skip, or mock away the defect to manufacture green.
- **Implementation came first**: Never delete or overwrite user work because it preceded the test; preserve it, add a red-capable test, and disclose the sequence.
- **Test-first does not fit**: For a spike, generated code, or configuration-only change, say why and define another failing validation signal; do not call tests written afterward TDD.
- **Test level**: Fast unit or contract tests first; integration, property, concurrency, or browser tests only where the boundary demands them; characterize legacy behavior before changing it.
- **Real collaborators**: Prefer real parsers, databases, filesystems, and HTTP handlers at lightweight boundaries over mocks of the unit under test; fake only paid, destructive, slow, or unreliable systems behind a narrow owned interface.
- **Readable failures**: Keep fixtures readable and assertions on outcomes; a little duplicated setup beats hidden intent, and a failure should explain the broken contract without a debugger.

## Documentation

- Adapted from [Superpowers test-driven-development](https://github.com/obra/superpowers/blob/44c9b2d6e889982ac18c27d05a19fefe335194e1/skills/test-driven-development/SKILL.md), [agent-skills test-driven-development](https://github.com/addyosmani/agent-skills/blob/d2478bf0c73a6357df39a3ed6aff16acaa218843/skills/test-driven-development/SKILL.md).
- Companion skills: [quality-assurance](../quality-assurance/SKILL.md) (risk-based campaign), [systematic-debugging](../systematic-debugging/SKILL.md) (unexplained failure), [plan-execution](../plan-execution/SKILL.md) (planned slices), [mise](../mise/SKILL.md) (task vocabulary).

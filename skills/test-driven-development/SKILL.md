---
name: test-driven-development
description: Implement an isolated bug fix or behavior change with an honest red-green-refactor cycle. Use when a regression test should fail before the correction, or when tested logic, refactors, or seams must prove correctness.
license: MIT
---

# Test-Driven Development

Use a failing test to prove the test can detect the missing or broken behavior, then write the smallest trustworthy change.

## Pragmatic Contract

- Test observable behavior and contracts, not private implementation or mock choreography.
- Scale the test level to risk. Prefer fast unit or contract tests; add integration, property, concurrency, or browser tests where the boundary demands them.
- For legacy code, first write a characterization or regression test around the behavior being changed.
- For a throwaway spike, generated code, or configuration-only change, state why test-first is inappropriate and define another failing validation signal. Do not silently call tests written afterward TDD.
- Never delete or overwrite user work merely because implementation preceded the test. Preserve it, establish a red-capable test, and disclose the sequence honestly.
- Never weaken an existing test, loosen a type, add a skip, or mock away the defect to manufacture green.

## Cycle

1. **Discover the harness:** Read repository instructions, existing tests, task definitions, and nearby patterns. Identify the smallest command that exercises the target behavior.
1. **State the contract:** Name the production change that would make the test pass and a plausible regression that would make it fail.
1. **RED:** Write one minimal test for one behavior. Run it and confirm it fails for the expected missing behavior, not a syntax, fixture, environment, or setup error.
1. **GREEN:** Implement only enough production code to satisfy that test. Run the focused test and read the full output.
1. **Protect the neighborhood:** Run the relevant package or subsystem tests. Fix production code when the new behavior breaks an existing valid contract; revisit the spec if contracts conflict.
1. **REFACTOR:** Improve names, structure, duplication, and types only while all tests remain green. Do not add new behavior during refactoring.
1. **Repeat:** Add the next smallest behavior, edge case, or failure path through a new red cycle.
1. **Prove the regression test:** For a bug fix, temporarily reverse or bypass the fix when safe and confirm the regression test fails, then restore the fix and confirm green.
1. **Protect unrelated work:** Before any full gate, inspect the full gate's task definition and working-tree state. If it runs whole-tree write-formatters and unrelated or user changes are present, validate the exact candidate in an isolated temporary worktree or run equivalent non-mutating checks; never reformat unrelated work.
1. **Run the full gate:** Execute the repository-owned format, check, test, and build contract, normally `mise run all`, warning-free.

## Test Quality

- Use deterministic inputs, explicit clocks, bounded retries, and local fakes.
- Prefer real parsers, databases, filesystems, and HTTP handlers at lightweight boundaries over mocks of the unit under test.
- Fake paid, destructive, slow, or unreliable external systems through a narrow owned interface.
- Test success, invalid input, boundary values, authorization, cancellation, errors, and state transitions in proportion to risk.
- Keep fixtures readable and assertions focused on outcomes. Duplicating a little setup in tests is cheaper than hiding intent behind an abstraction.
- Ensure failures explain the broken contract without requiring a debugger.

## Evidence

Report the red command and expected failure, focused green command, wider suite, full gate, and any boundary still covered only manually or not at all.

Use [quality-assurance](../quality-assurance/SKILL.md) for a broader risk-based test campaign and [systematic-debugging](../systematic-debugging/SKILL.md) when the observed failure is not yet understood.

## Sources

Adapted independently from [Superpowers test-driven-development at `44c9b2d`](https://github.com/obra/superpowers/blob/44c9b2d6e889982ac18c27d05a19fefe335194e1/skills/test-driven-development/SKILL.md) and [agent-skills test-driven-development at `d2478bf`](https://github.com/addyosmani/agent-skills/blob/d2478bf0c73a6357df39a3ed6aff16acaa218843/skills/test-driven-development/SKILL.md).

---
name: implementation-plan
description: Turn accepted requirements into ordered dependency-aware slices and a repository-grounded implementation plan. Use before editing when work spans systems, has migration or rollout risk, needs architecture choices, or should be split.
license: MIT
---

# Implementation Plan

Design the smallest sequence of independently verifiable vertical slices that satisfies the accepted intent.

## Authority Boundary

Planning is read-only by default. Do not edit source, create issues, install dependencies, commit, push, deploy, or publish while producing a plan unless the user explicitly requested that action.

## Workflow

1. **Read the contract:** Re-read the accepted spec or request. Extract requirements, non-goals, constraints, success criteria, authority limits, and unresolved decisions.
1. **Inspect reality:** Read repository instructions, status, relevant source, tests, manifests, task definitions, recent history, and existing patterns. Preserve staged, unstaged, and untracked user work.
1. **Verify APIs:** Confirm installed dependency versions and read their exact local source plus current authoritative documentation before planning against an unfamiliar API.
1. **Map the change:** Describe the current flow, desired flow, affected components, interfaces, data, trust boundaries, and operational consumers. Record existing code to reuse or delete.
1. **Compare real options:** For a meaningful architectural or tooling choice, present two or three numbered options with complexity, maintenance, security, reversibility, and migration trade-offs. Recommend the simplest adequate option.
1. **Lock boundaries:** Assign each file or package one responsibility. Prefer deep modules that hide decisions over pass-through layers, put tests and callers across the same real seam, and delay a generalized adapter until more than one concrete variation proves it. Record interface signatures, ownership, invariants, domain vocabulary, and compatibility behavior shared across tasks.
1. **Slice vertically:** Order work by dependencies so each slice delivers a coherent behavior through its test boundary. Fold scaffolding, configuration, documentation, and telemetry into the slice that needs them.
1. **Design proof first:** For each slice, name the failing or characterization test, focused command, full project gate, runtime/manual evidence, and expected failure or success signal.
1. **Plan safe change:** Cover data or schema migration, backward compatibility, feature flags, observability, rollout, rollback, support, and cleanup only when applicable.
1. **Mark concurrency:** Parallelize only independent tasks with disjoint ownership. Name shared files, generated outputs, runtime leases, and merge order explicitly.
1. **Self-review:** Map every requirement to a task and proof, remove placeholders and speculative flexibility, verify paths and symbols, and identify anything that still requires a user decision. Apply the deletion test: an abstraction earns its place only when removing it would spread meaningful complexity or violate a real seam.

## Task Contract

Each task must include:

- **Outcome:** One observable behavior or enabling invariant.
- **Dependencies:** Earlier tasks, external decisions, and shared resources.
- **Files:** Exact create/modify/test paths and the responsibility of each change.
- **Interfaces:** Inputs, outputs, signatures, schemas, or commands neighboring tasks rely on.
- **Steps:** Small actions ordered test first, implementation second, focused verification third.
- **Proof:** Expected red signal, focused green command, full gate, and any runtime/manual check.
- **Safety:** Migration, compatibility, security, privacy, rollback, or authority notes.
- **Done:** Acceptance criteria that can be checked without subjective judgment.

For two or three low-risk linear slices, use a compact task table instead: retain outcome, exact files or interfaces, red-green proof, applicable safety, and objective done criteria, followed by a one-line dependency chain and explicit execution handoff. Omit empty parallel-lane, migration, rollback, and critical-path prose; never omit a real risk or proof boundary. Larger or cross-cutting plans use the full task contract and end with a dependency graph, critical path, parallel lanes, proof-boundary checklist, and explicit execution handoff. Use [plan-review](../plan-review/SKILL.md) for high-risk or strategic work and [plan-execution](../plan-execution/SKILL.md) only after the plan is accepted.

## Sources

Adapted independently from [Superpowers writing-plans at `44c9b2d`](https://github.com/obra/superpowers/blob/44c9b2d6e889982ac18c27d05a19fefe335194e1/skills/writing-plans/SKILL.md), [GitHub Spec Kit's plan template at `684b3d8`](https://github.com/github/spec-kit/blob/684b3d8e05263a7c1948d3d0699ab1cb4f77c3d5/templates/plan-template.md), [Matt Pocock's codebase design at `84fdeff`](https://github.com/mattpocock/skills/blob/84fdeffd12f2ee307994d1eb6feb48173b6e0502/skills/engineering/codebase-design/SKILL.md), and [domain modeling at `84fdeff`](https://github.com/mattpocock/skills/blob/84fdeffd12f2ee307994d1eb6feb48173b6e0502/skills/engineering/domain-modeling/SKILL.md).

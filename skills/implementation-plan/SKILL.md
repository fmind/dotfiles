---
name: implementation-plan
description: Turn accepted requirements into ordered dependency-aware slices and a repository-grounded implementation plan. Use before editing when work spans systems, has migration or rollout risk, needs architecture choices, or should be split.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/implementation-plan
  created: "2026-08-08"
  updated: "2026-09-03"
---

# Implementation Plan

Design the smallest sequence of independently verifiable vertical slices that satisfies accepted requirements; [plan-review](../plan-review/SKILL.md) challenges the plan and [plan-execution](../plan-execution/SKILL.md) carries it out.

## Workflow

1. **Read the contract**: Re-read the accepted spec or request; extract requirements, non-goals, constraints, success criteria, authority limits, and unresolved decisions.
1. **Inspect reality**: Read repository instructions, status, relevant source, tests, manifests, tasks, recent history, and existing patterns. Preserve staged, unstaged, and untracked user work.
1. **Verify unfamiliar APIs**: Use [technical-research](../technical-research/SKILL.md) before planning against a dependency whose local source you have not read.
1. **Map the change**: Describe current flow, desired flow, affected components, interfaces, data, trust boundaries, and operational consumers; record code to reuse or delete and name the real architectural choices.
1. **Lock boundaries**: Give each file or package one responsibility; record interface signatures, ownership, invariants, domain vocabulary, and compatibility behavior shared across tasks.
1. **Slice vertically**: Order work by dependencies so each slice delivers coherent behavior through its test boundary; fold scaffolding, configuration, documentation, and telemetry into the slice that needs them.
1. **Design proof first**: For each slice name the failing or characterization test, focused command, full project gate, runtime or manual evidence, and the expected failure or success signal.
1. **Plan safe change**: Cover migration, backward compatibility, feature flags, observability, rollout, rollback, support, and cleanup only when applicable.
1. **Mark concurrency**: Parallelize only independent tasks with disjoint ownership; name shared files, generated outputs, runtime leases, and merge order.
1. **Self-review**: Map every requirement to a task and proof, remove placeholders and speculative flexibility, verify paths and symbols, and list anything that still needs a user decision.
1. **Write each task**: In a full plan every task carries:
   - **Outcome**: one observable behavior or enabling invariant.
   - **Dependencies**: earlier tasks, external decisions, and shared resources.
   - **Files**: exact create, modify, and test paths with the responsibility of each change.
   - **Interfaces**: inputs, outputs, signatures, schemas, or commands neighboring tasks rely on.
   - **Steps**: test first, implementation second, focused verification third.
   - **Proof**: expected red signal, focused green command, full gate, and any runtime or manual check.
   - **Safety**: migration, compatibility, security, privacy, rollback, or authority notes.
   - **Done**: acceptance criteria checkable without subjective judgment.
1. **Scale the shape**: Match the plan's weight to its risk and hand off explicitly.
   - For two or three low-risk linear slices, use a compact task table with outcome, exact files or interfaces, red-green proof, applicable safety, and objective done criteria, followed by a one-line dependency chain and an execution handoff.
   - Omit empty parallel-lane, migration, rollback, and critical-path prose; never omit a real risk or proof boundary.
   - Larger or cross-cutting plans use the full task contract and end with a dependency graph, critical path, parallel lanes, proof checklist, and execution handoff.

## Gotchas

- **Planning is read-only by default**: Do not edit source, create issues, install dependencies, or deploy while planning unless the user explicitly asked.
- **Deletion test**: An abstraction earns its place only when removing it would spread meaningful complexity or violate a real seam; delay a generalized adapter until a second concrete variation exists.
- **Deep modules**: Prefer modules that hide decisions over pass-through layers, and put tests and callers across the same real seam.

## Documentation

- Adapted from [Superpowers writing-plans](https://github.com/obra/superpowers/blob/44c9b2d6e889982ac18c27d05a19fefe335194e1/skills/writing-plans/SKILL.md), [Spec Kit plan template](https://github.com/github/spec-kit/blob/684b3d8e05263a7c1948d3d0699ab1cb4f77c3d5/templates/plan-template.md), [codebase design](https://github.com/mattpocock/skills/blob/84fdeffd12f2ee307994d1eb6feb48173b6e0502/skills/engineering/codebase-design/SKILL.md), [domain modeling](https://github.com/mattpocock/skills/blob/84fdeffd12f2ee307994d1eb6feb48173b6e0502/skills/engineering/domain-modeling/SKILL.md).
- Companion skills: [plan-review](../plan-review/SKILL.md) (challenge the plan), [plan-execution](../plan-execution/SKILL.md) (carry it out), [technical-research](../technical-research/SKILL.md) (verify APIs), [test-driven-development](../test-driven-development/SKILL.md) (red-green proof), [threat-model](../threat-model/SKILL.md) (security design risk).

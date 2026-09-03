---
name: plan-execution
description: Execute an accepted implementation plan in bounded, verified slices. Use to coordinate agents and shared-file ownership, resume planned work, or finish scoped tasks without crossing commit or deploy boundaries.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/plan-execution
  created: "2026-08-08"
  updated: "2026-09-03"
---

# Plan Execution

Turn an accepted plan into reviewable evidence slice by slice; [implementation-plan](../implementation-plan/SKILL.md) writes the plan, and commit, push, deploy, and publication stay separate authorities.

## Workflow

1. **Confirm authority**: A plan alone is not authorization to mutate the repository; confirm the request includes implementation, then re-read the request, plan, repository instructions, and status before the first edit.
1. **Capture the baseline**: Record branch, `HEAD`, status, relevant test state, and known external boundaries. Preserve staged, unstaged, and untracked user work and note which edits belong to the user.
1. **Reconcile the plan**: Check that paths, symbols, dependency versions, commands, and assumptions still match the checkout; note deviations in the execution notes and ask for direction when the outcome would materially change.
1. **Order the work**: Follow the dependency graph and critical path; take the smallest vertical slice that yields independently useful behavior and proof.
1. **Establish red evidence**: Use [test-driven-development](../test-driven-development/SKILL.md) for behavior changes, a characterization test for legacy behavior, or a failing validation signal for non-code changes.
1. **Implement minimally**: Use the applicable stack skill and touch only files traceable to the current slice.
1. **Verify the slice**: Run the focused test, formatter, static checks, and any bounded runtime or manual check the plan promised; read the full output.
1. **Review the delta**: Compare the diff with the intended slice, remove only artifacts this work introduced, and use [diff-review](../diff-review/SKILL.md) for risky changes.
1. **Checkpoint honestly**: Mark a slice complete only with fresh evidence; record partial results, deviations, and residual gaps before moving on.
1. **Gate the candidate**: Run the full gate (`mise run all`); if the tree carries unrelated changes and the gate write-formats, run it in a temporary `git worktree` or fall back to `mise run check` and `mise run test` (see [mise](../mise/SKILL.md)).
1. **Hand off**: Summarize changes, verification, untested boundaries, and remaining tasks; report the highest proven rung of the [proof ladder](../production-readiness/SKILL.md) and leave the worktree for review.

## Gotchas

- **Parallel agents**: Delegate only independent tasks with disjoint file or runtime ownership; give each worker the raw task contract and minimum context, not the intended answer.
- **Shared state**: Assign one owner to shared files, generated state, migrations, and runtime leases; serialize integration and gate the combined candidate.
- **Agent reports are not evidence**: Inspect every returned diff and rerun its proof.
- **Stop rather than weaken an assertion**: never skip a gate, broaden an exclusion, or hide a warning; use [systematic-debugging](../systematic-debugging/SKILL.md) for unexpected failures and revisit the plan when repeated fixes expose new coupling.
- **Blocked proof**: When a service, credential, decision, or runtime blocks proof, continue safe local work and report the exact gap; never claim completion because the budget or context is exhausted.

## Documentation

- Adapted from [Superpowers executing-plans](https://github.com/obra/superpowers/blob/44c9b2d6e889982ac18c27d05a19fefe335194e1/skills/executing-plans/SKILL.md), [Superpowers subagent-driven development](https://github.com/obra/superpowers/blob/44c9b2d6e889982ac18c27d05a19fefe335194e1/skills/subagent-driven-development/SKILL.md).
- Companion skills: [implementation-plan](../implementation-plan/SKILL.md) (the plan), [test-driven-development](../test-driven-development/SKILL.md) (red-green slices), [diff-review](../diff-review/SKILL.md) (risky deltas), [systematic-debugging](../systematic-debugging/SKILL.md) (unexpected failures), [production-readiness](../production-readiness/SKILL.md) (proof ladder).

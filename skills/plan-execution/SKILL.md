---
name: plan-execution
description: Execute an accepted implementation plan in bounded, verified slices. Use to coordinate multiple agents and shared-file ownership, resume planned work, or finish scoped tasks without crossing commit, push, deploy, or publication boundaries.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/plan-execution
  created: 2026-08-08
  updated: 2026-08-09
---

# Plan Execution

Turn an accepted plan into reviewable evidence while preserving the user's work and authority boundaries.

## Preconditions

- Confirm the requested action includes implementation. A plan alone is not authorization to mutate the repository.
- Re-read the original request, accepted plan, repository instructions, and current status before the first edit.
- Preserve staged, unstaged, and untracked user work. Never reset, clean, stash, overwrite, or reformat unrelated files.
- Treat commit, push, pull request, issue mutation, deployment, infrastructure apply, external messaging, and publication as separate authorities.
- Stop rather than weaken an assertion, skip a gate, broaden an exclusion, or hide a warning.

## Workflow

1. **Reconcile the plan:** Verify that paths, symbols, dependency versions, commands, and assumptions still match the checkout. Amend the execution notes when reality differs; request direction if the required outcome would materially change.
1. **Capture the baseline:** Record branch, HEAD, status, relevant test state, and known external boundaries. Identify which existing edits belong to the user.
1. **Order the work:** Follow the dependency graph and critical path. Choose the smallest vertical slice that produces independently useful behavior and proof.
1. **Establish red evidence:** Use [test-driven-development](../test-driven-development/SKILL.md) for behavior changes or create a characterization test when changing legacy behavior. For non-code changes, define the failing contract or validation signal first.
1. **Implement minimally:** Use the applicable language, infrastructure, site, or document skill. Touch only files traceable to the current slice and explain non-obvious trade-offs at the decision point.
1. **Verify the slice:** Run the focused test, formatter, static checks, and any bounded runtime/manual check promised by the plan. Read the full output and fix root causes.
1. **Review the delta:** Compare the actual diff with the intended slice. Remove only artifacts introduced by this work, confirm no requirement drift, and use [diff-review](../diff-review/SKILL.md) for risky changes.
1. **Checkpoint honestly:** Mark the slice complete only with fresh evidence. Record partial results, deviations, and residual gaps before moving on.
1. **Protect unrelated work:** Before any full gate, inspect the full gate's task definition and working-tree state. If it runs whole-tree write-formatters and unrelated or user changes are present, validate the exact candidate in an isolated temporary worktree or run equivalent non-mutating checks; never reformat unrelated work.
1. **Run the whole contract:** Test changed behavior first, then run the repository-owned full gate, normally `mise run all`. Keep local, exact-head CI, runtime, deployed, and published evidence separate.
1. **Hand off:** Summarize changes, verification, untested boundaries, and remaining tasks. Leave the worktree for review unless further mutation was explicitly authorized.

## Parallel Work

- Delegate only independent tasks with disjoint file or runtime ownership.
- Give each worker the raw task contract and minimum context, not the intended answer.
- Assign one owner to shared files, generated state, migrations, and runtime leases.
- Inspect every returned diff and rerun its proof; an agent's success report is not evidence.
- Serialize integration, then run the full gate on the combined candidate.

## Failure Protocol

- Use [systematic-debugging](../systematic-debugging/SKILL.md) for unexpected failures instead of stacking speculative edits.
- After repeated failed fixes that expose new coupling, stop and revisit the architecture or plan.
- If an external service, credential, human decision, or authorized runtime blocks proof, continue safe local work and report the exact ceiling.
- Never claim completion because the budget, context, or patience is exhausted.

## Sources

Adapted independently from [Superpowers executing-plans at `44c9b2d`](https://github.com/obra/superpowers/blob/44c9b2d6e889982ac18c27d05a19fefe335194e1/skills/executing-plans/SKILL.md) and [Superpowers subagent-driven development at `44c9b2d`](https://github.com/obra/superpowers/blob/44c9b2d6e889982ac18c27d05a19fefe335194e1/skills/subagent-driven-development/SKILL.md).

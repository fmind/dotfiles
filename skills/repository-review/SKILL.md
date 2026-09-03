---
name: repository-review
description: Audit a whole repository across architecture, source, tests, tooling, security, CI/CD, docs, and releases with ranked findings. Use for cross-cutting audits, not one diff.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/repository-review
  created: "2026-08-01"
  updated: "2026-09-03"
---

# Repository Review

Audit the whole delivery system read-only and report ranked, evidence-backed findings; [diff-review](../diff-review/SKILL.md) owns one change, [project-backlog](../project-backlog/SKILL.md) turns findings into issues, and [project-health](../project-health/SKILL.md) applies the fixes.

## Workflow

1. **Scope the review**: Read the repository instructions and the stack, CI, security, or documentation skills the requested dimensions need; confirm which live systems are authorized and treat every other credential, cluster, project, deployment, or release as unavailable.
1. **Record the candidate**: Capture branch, `HEAD`, upstream, and `git status --short`; distinguish staged, unstaged, and untracked paths, treat pre-existing changes as user-owned, and state whether evidence concerns committed `HEAD`, the dirty tree, or both.
1. **Map the system**: Inspect manifests, entry points, package boundaries, tasks, hooks, workflows, deployment and release automation, docs, generated files, and runtime configuration against the [review matrix](references/review-matrix.md).
1. **Run native checks**: Iterate with `mise run check`, targeted `check:*` subtasks, and `mise run test`; invoke [secure](../secure/SKILL.md) only when the request needs the full security boundary.
1. **Gate the candidate**: Run the full gate (`mise run all`); if the tree carries unrelated changes and the gate write-formats, run it in a temporary `git worktree` or fall back to `mise run check` and `mise run test` (see [mise](../mise/SKILL.md)).
1. **Inspect live evidence read-only**: Compare the exact `HEAD` SHA with CI checks, deployments, tags, releases, or runtime observations only where authorized and reachable; a green result for another head is stale.
1. **Challenge each conclusion**: Reproduce high-impact claims when safe, prefer repository or dependency source over assumptions, and separate an observed defect from a speculative risk.
1. **Report findings**: Lead with **Key findings** ranked `P0`–`P3` (scale in [diff-review](../diff-review/SKILL.md)), each with the defect, impact, and direct file, command, CI, or runtime evidence; if none are material, say so and list residual gaps.
1. **State proof and actions**: Report the highest proven rung of the [proof ladder](../production-readiness/SKILL.md) with each check marked pass, fail, blocked, or not run, then list **Actions** by priority, each starting with a verb and naming the owner or required authority.

## Gotchas

- **Review only**: A review request authorizes inspection and bounded validation, not fixes, issues, comments, deployment, or publication; never write a report file or backlog item unless the user asks.
- **Dirty tree**: Preserve it and state which candidate each check covered; a full gate on a dirty tree is coherent only against a materialized candidate.
- **Partial scan**: Keep the exact timeout, unavailable database, skipped target, or truncated scope; a partial scan is neither green nor finding-free.
- **Stale CI**: Report the checked SHA and the reviewed SHA; never transfer a result across commits.
- **Documentation drift**: Prove mismatches against live tasks, CLI metadata, files, workflows, or generated output, not prose alone.
- **Unavailable runtime**: Name the missing tool, credential, service, cluster, or authorization and stop at the local boundary; readiness is not runtime acceptance.

## Documentation

- Companion skills: [diff-review](../diff-review/SKILL.md) (one change and the `P0`–`P3` scale), [production-readiness](../production-readiness/SKILL.md) (proof ladder and go/no-go), [secure](../secure/SKILL.md) (security boundary), [project-backlog](../project-backlog/SKILL.md) (findings to issues), [project-health](../project-health/SKILL.md) (apply the fixes).

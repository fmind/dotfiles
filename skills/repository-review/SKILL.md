---
name: repository-review
description: Perform an evidence-first, cross-cutting repository review across architecture, source, tests, tooling, security, CI/CD, documentation, generated state, releases, and authorized runtime behavior. Use for repository audits, full reviews, readiness assessments, technical-debt reviews, or any request that needs ranked findings and honest proof boundaries without implementing fixes or creating backlog items.
license: MIT
---

# Repository Review

Review the whole delivery system, preserve user work, and report only what the available evidence proves. A review-only request authorizes inspection and bounded validation, not fixes, issue creation, comments, deployment, publication, or runtime mutation.

Agent presentation metadata is kept in [openai.yaml](agents/openai.yaml).

## Authority and proof ceiling

1. Read the repository instructions and the stack, CI, security, or documentation skills needed by the requested dimensions. The full security-scan skill is not mandatory for a bounded review that excludes the security boundary.
1. Confirm the requested scope and any explicitly authorized live systems. Treat credentials, clusters, cloud projects, deployments, and releases outside that scope as unavailable.
1. Record the starting branch, `HEAD`, upstream, and `git status --short`. Distinguish staged, unstaged, and untracked paths; every pre-existing change is user-owned.
1. Never stash, reset, clean, reformat, stage, commit, push, comment, open issues, or otherwise mutate user or external state during a review.

## Review workflow

1. **Establish the candidate**: Decide whether the evidence concerns committed `HEAD`, the dirty working candidate, or both. Name that boundary in the report.
1. **Preserve the worktree**: Use read-only diffs first. If a coherent clean-tree gate can modify files, materialize the intended candidate in an isolated temporary worktree only when temporary local file and Git-metadata changes are authorized, validate the exact paths and Git state before use, and run the mutating gate only there. If the caller forbids even temporary state, skip that gate and report it as not checked. Verify the original index and worktree snapshot after cleanup.
1. **Map the system**: Inspect manifests, entry points, package boundaries, task definitions, hooks, workflows, deployment/release automation, user and agent documentation, generated files, and relevant runtime configuration. Read [the review matrix](references/review-matrix.md) for coverage and evidence expectations.
1. **Route native checks**: Use the repository's pinned `mise run` tasks and the applicable language stack skills. Use `mise run fast` only for iteration; use `mise run all` for the complete local candidate when safe and proportionate. Invoke the security-scan skill only when the request calls for a full security boundary beyond native checks.
1. **Inspect live evidence read-only**: Compare the exact `HEAD` SHA with CI checks, deployments, tags, releases, or runtime observations only when those systems are authorized and reachable. A historical or different-head green result is stale evidence.
1. **Challenge each conclusion**: Reproduce high-impact claims when safe, prefer repository or dependency source over assumptions, and separate an observed defect from a speculative risk.
1. **Report, do not repair**: Return ranked findings and ordered actions in the response. Never create an unsolicited report file or backlog item; compose with a backlog skill only when the user explicitly requests that mutation.

## Required failure-state handling

- **Dirty tree**: Preserve it and state which candidate each check covered. Never call a dirty-tree full gate coherent unless it ran against an isolated, materialized candidate.
- **Partial scan**: Retain the exact timeout, unavailable database, skipped target, or truncated scope. A partial scan is neither green nor a finding-free scan.
- **Stale CI**: Report the checked SHA and the reviewed SHA. Do not transfer a result across commits.
- **Documentation drift**: Prove mismatches against live tasks, CLI metadata, files, workflows, or generated output; do not judge prose in isolation.
- **Unavailable runtime**: Report the missing tool, credential, service, cluster, or authorization and stop at the source/local boundary. Readiness is not runtime acceptance.

## Output contract

Lead with an executive summary using these sections:

1. **Key findings**: Rank by impact (`critical`, `high`, `medium`, `low`). State the defect or risk, impact, and direct file, command, CI, public, or runtime evidence. If no material findings exist, say so and list residual gaps.
1. **Proof boundaries**: State source-ready, local-green, exact-head-CI, runtime-proven, deployed, and release-published separately. Use `not checked`, `blocked`, or `failed` rather than implying success from absence of evidence.
1. **Actions**: Order the smallest root-cause actions by priority. Start each item with a verb, name the owner or required authority when relevant, and do not claim an action was taken during a review.

Keep commands and logs concise, redact secrets, and include enough evidence for another reviewer to reproduce the conclusion.

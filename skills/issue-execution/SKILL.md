---
name: issue-execution
description: Execute one eligible GitHub issue end to end through cooperative claiming, bounded context, acceptance-driven implementation, repository validation, and an auditable handoff. Use when asked to pick up, implement, finish, or deliver an already-approved issue; reject blocked or human-gated work and never infer authorization to publish, merge, deploy, release, or mutate runtime state.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/issue-execution
  created: 2026-08-01
  updated: 2026-08-01
---

# Issue Execution

Consume one approved backlog item without expanding its authority. Compose [github-triage](../github-triage/SKILL.md) for queue and lease mechanics, [repository-review](../repository-review/SKILL.md) for evidence and proof boundaries, [feature-branch](../feature-branch/SKILL.md) for branch creation, [git-add-commit-push](../git-add-commit-push/SKILL.md) for authorized publication, and [github-pull-request](../github-pull-request/SKILL.md) for the handoff PR. Load the applicable language, CI, security, documentation, deployment, or release skill only when the issue scope requires it.

Use the [handoff contract](references/handoff-contract.md) before reporting completion. The offline failure cases are versioned in [cases.json](tests/fixtures/cases.json). Agent metadata is in [openai.yaml](agents/openai.yaml).

## 1. Resolve before claiming

1. Resolve the local root and confirmed `owner/repo`; read repository instructions and record the current branch, `HEAD`, upstream, staged, unstaged, and untracked state.
1. Read the issue's complete body, comments, labels, state, assignees, native `blockedBy` and `blocking` relationships, linked pull requests, and recent lease comments. Fetch relevant source, tests, documentation, workflows, and authorized live evidence before deciding that the issue is runnable.
1. Confirm the issue appears in the `github-triage` runnable queue and has the required problem, evidence, acceptance criteria, boundaries, validation, area, priority, and effort contract.
1. Reject with an explicit reason when the issue is blocked, actively claimed, `needs-human`, an epic, ambiguous, closed, already satisfied, missing actionable acceptance criteria, or inconsistent with current source. Do not manufacture a change to close already-satisfied work.
1. Treat a stale `status/in-progress` label as an audit trigger, never an automatic takeover. Use the triage stale query, inspect current comments plus branch or worktree activity, and post a takeover claim only when the lease is demonstrably stale.

## 2. Claim and isolate

1. Choose one coherent issue. Add `status/in-progress` and post the cooperative claim with agent identity, UTC RFC3339 timestamp, branch or worktree, and bounded scope.
1. Preserve all user-owned staged, unstaged, and untracked files. If the current tree is dirty, use read-only diffs and a separate worktree from the confirmed base; never stash, reset, clean, overwrite, or silently reformat user work.
1. Create a conventional issue-specific branch through `feature-branch`, unless the user explicitly authorizes another repository-supported flow. Post a lease update before the two-hour deadline while work continues.
1. Generate bounded project-only context with `dot context --format json --tokens <budget>` when available. Treat omissions and failures as visible evidence gaps; session history, private stores, credentials, and unrelated repositories are not ambient context.

## 3. Verify and implement

1. Convert every acceptance criterion into an evidence matrix before editing: current status, implementation surface, focused check, and required proof boundary.
1. Reproduce the reported gap against current source when safe. Challenge scope that is stale, duplicative, speculative, or already solved; comment and release the lease instead of coding blindly.
1. Implement the smallest complete root-cause solution. Preserve type safety and failure visibility; never weaken assertions, suppress warnings, skip checks, or introduce placeholder debt to obtain green output.
1. During iteration, run the repository's declared fast gate when it exists. Otherwise run the smallest deterministic focused formatter, checker, and test for the changed surface. A focused or fast result is iteration evidence only.
1. Re-read the issue after implementation and verify every criterion against current code, tests, documentation, workflows, generated state, and explicitly authorized live behavior. Record unmet criteria rather than inferring them from a green test suite.

## 4. Validate and classify proof

1. Run the full canonical repository gate, normally `mise run all`, against the coherent candidate before calling the implementation locally complete. Preserve exact failures and partial scans.
1. Recheck the original user-owned worktree and index after any isolated validation. Confirm no unrelated file entered the candidate.
1. Classify evidence separately as source-ready, focused-green, full-local-green, exact-head-CI, runtime-proven, deployed, and release-published. Never transfer a proof result across commits or promote local green into an external boundary.
1. If validation is partial or failing, stop the completion path, keep the lease current, and report the root failure. Do not publish, merge, close, or release to hide a red boundary.

## 5. Apply explicit publication authority

1. Map the user's instruction to exact allowed mutations. Implementing does not authorize commit; commit does not authorize push; push does not authorize a PR; a PR does not authorize merge or issue closure; merge does not authorize deployment or release.
1. Only when explicitly authorized, stage the issue-only diff, use `git-add-commit-push`, and open or update a structured PR with `github-pull-request`. Keep the issue open while review or required exact-head checks remain pending.
1. Only when explicitly authorized to merge, verify the PR head SHA, required reviews, and exact-head CI immediately before merge. Verify the resulting exact `main` SHA and its CI before reporting the merged boundary or closing the issue.
1. Invoke deployment or release workflows only under separate explicit authority and the applicable skill. Never mutate credentials, paid services, production, clusters, tags, releases, or irreversible state merely because the terminal objective says finish or do not stop.

## 6. Handoff and release the lease

1. Post concise issue evidence mapping each acceptance criterion to source, tests, PR, CI, or authorized live proof. Name every unverified layer and exact remaining gate.
1. On completion, remove `status/in-progress` after the final authorized state is verified. On abandonment, post the exact reason and remove the lease.
1. Add `needs-human` only for a concrete decision, credential, approval, or spend that cannot be resolved safely; name the unresolved choice and stop. Do not use it for ordinary test failures, hard work, or missing implementation effort.
1. Return the sections in the handoff contract. Never claim completion from a commit, PR, or locally green gate when the requested terminal state is stronger.

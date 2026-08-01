# Issue Execution Handoff

## Acceptance evidence

List every acceptance criterion separately with one status:

- `verified`: direct current evidence satisfies the criterion.
- `partial`: some required evidence exists and the missing boundary is explicit.
- `failed`: current evidence contradicts the criterion.
- `not checked`: the required system or authority was unavailable.

For each criterion, record the source path or external object, exact commit SHA where applicable, command or observation, and result. A suite-wide green result supplements but never replaces criterion-specific evidence.

## Validation

Report focused checks, full canonical local gate, exact-head CI, and clean-tree confirmation independently. Include exact commands, SHAs, and URLs for external checks. Retain warning, timeout, skipped-target, unavailable-service, and partial-scan output as a non-green result.

## Proof boundaries

Use these exact states:

1. `source-ready`: implementation and criterion evidence exist in the candidate.
1. `focused-green`: changed-surface checks passed.
1. `full-local-green`: the complete repository gate passed on the coherent candidate.
1. `exact-head-CI`: required CI passed on the named candidate or merged SHA.
1. `runtime-proven`: authorized representative runtime acceptance passed.
1. `deployed`: the named environment accepted the exact artifact or SHA.
1. `release-published`: the immutable public tag, release, and required artifacts were verified.

Mark each boundary `verified`, `failed`, `blocked`, or `not checked`; never infer a stronger boundary from a weaker one.

## Repository and user-work integrity

Record the confirmed repository, issue, branch or worktree, candidate SHA, base SHA, and final `git status --short`. Confirm the original staged, unstaged, and untracked user paths remain preserved and no unrelated change was committed.

## External mutations and lease

List every authorized mutation actually performed: claim, comments, labels, commit, push, PR, merge, issue closure, deployment, or release. Include URLs or identifiers. State whether `status/in-progress` was removed, retained with a current heartbeat, or replaced by `needs-human` for a named decision.

## Remaining actions

Name only concrete unresolved gates, their owner or required authority, and the exact safe resume point. Do not create an unsolicited report file outside this skill contract.

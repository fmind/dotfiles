---
name: project-backlog
description: Turn audit findings into deduplicated, prioritized issue drafts with dependencies and explicit authorization before any GitHub mutation. Use when review findings must become tracked issues.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/project-backlog
  created: "2026-08-01"
  updated: "2026-09-03"
---

# Project Backlog

Turn [repository-review](../repository-review/SKILL.md) findings into deduplicated, ordered issue drafts; discovery and drafting are read-only, and [github-issues](../github-issues/SKILL.md) owns the authorized creation of each issue.

## Workflow

1. **Confirm the target**: Read the local repository and its instructions with `git`, then resolve the candidate with `gh repo view --json nameWithOwner,visibility`; show it to the user without treating discovery as mutation authorization.
1. **Review**: Run the cross-cutting repository review; keep partial scans and unavailable services as evidence gaps.
1. **Read existing issues**: Fetch the open and closed issues needed for duplicate analysis with complete bodies, comments, labels, state, and native `blockedBy` and `blocking` relationships; search by the underlying problem and evidence, not title similarity.
1. **Research sparingly**: Consult primary documentation or upstream source only where it materially confirms a retained finding; record the version context and separate sourced fact from inference.
1. **Classify**: Mark each candidate a **verified finding** or a **trend opportunity**; a trend becomes an issue only when current project evidence proves fit and value.
1. **Reject**: Drop candidates that duplicate an existing issue, lack reproducible evidence, exceed the project's likely value, restore rejected scope, or add unjustified complexity.
1. **Draft**: Write every retained item with the [draft contract](references/draft-contract.md); explain why it is distinct from each close match and model dependencies as draft-to-draft or draft-to-issue edges.
1. **Stop at the gate**: Present deduplication decisions, ordered drafts, and the dependency graph; proceed only when the user explicitly authorizes creation in the confirmed repository, which is the only mutation target.
1. **Create issues**: Refresh visibility, labels, matching issues, and native dependencies immediately before writing and stop if drafts went stale; create each issue in draft order through [github-issues](../github-issues/SKILL.md) and record draft ID, issue number, node ID, and URL after each success.
1. **Add edges**: Only after every node exists, read current relationships and add the missing ones with the `addBlockedBy` mutation from the draft contract; never encode a dependency only in prose.
1. **Verify**: Read back every body, label set, and `blockedBy` and `blocking` relationship from GitHub before claiming completion.
1. **Report**: Return these sections in order.
   - **Review evidence**: highest proven rung of the [proof ladder](../production-readiness/SKILL.md) and material gaps.
   - **Deduplication**: candidate-to-existing-issue decisions and distinctness rationale.
   - **Draft backlog**: ordered drafts with routing and dependency edges.
   - **Authorization gate**: exact repository and mutations awaiting approval, or the authorization already received.
   - **Mutation receipt**: only after authorized writes; created issues, labels, verified native edges, and any partial state.

## Gotchas

- **Unauthorized writes** and **partial issue creation** are failed mutation boundaries: authorization to review, draft, implement, or open a pull request never authorizes issue creation, and a partial run is reported, not improvised around.
- **Partial creation**: Keep successful issues, create no edges, and list created, failed, and unattempted drafts; on a partial edge run keep successful edges, list the rest, and let a retry re-read and apply only missing edges.
- **Labels**: Prefer the repository's existing `area/*`, `priority/*`, and `effort/*` labels; when they are missing, propose the label set in the draft and let the user decide instead of stopping.
- **Public repositories**: Never copy private paths, credentials, customer data, private issue text, or non-public runtime details into drafts; sanitize the evidence or mark the draft `needs-human`.
- **Unavailable services**: When GitHub, research sources, or credentials are unavailable, keep local drafts, name the missing verification, and perform no write.
- **Dirty worktree**: It is review evidence, not permission to stash, clean, stage, format, or commit.

## Documentation

- [GitHub GraphQL mutations](https://docs.github.com/en/graphql/reference/mutations#addblockedby) · [Draft contract](references/draft-contract.md)
- Companion skills: [repository-review](../repository-review/SKILL.md) (the findings), [github-issues](../github-issues/SKILL.md) (issue creation and edits), [production-readiness](../production-readiness/SKILL.md) (proof ladder).

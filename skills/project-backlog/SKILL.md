---
name: project-backlog
description: Turn an evidence-first repository review into a deduplicated, prioritized, dependency-aware project backlog. Use when auditing a project for actionable improvements, preparing issue drafts, grooming technical debt, or creating GitHub issues from verified findings; default to reviewable drafts and require explicit authorization plus a confirmed repository before any GitHub mutation.
license: MIT
---

# Project Backlog

Compose the [repository-review skill](../repository-review/SKILL.md) with the repository's live issue taxonomy. Discovery and drafting are read-only by default; issue creation and native dependency mutation are a separate, explicitly authorized phase.

Agent presentation metadata is kept in [openai.yaml](agents/openai.yaml). Use the complete [draft contract](references/draft-contract.md) and exercise the failure boundaries in the [behavioral evaluations](tests/behavioral-evaluations.md).

## Phase 1: Read-only discovery

1. Confirm the local repository and read its instructions. Preserve staged, unstaged, and untracked work exactly as required by the repository-review skill.
1. Resolve the candidate GitHub repository with `gh repo view --json nameWithOwner,url,visibility`; show it to the user, but do not treat discovery as mutation authorization.
1. Run the cross-cutting repository review across source, tests, tooling, CI/CD, security, documentation, generated state, release state, and relevant authorized live behavior. Retain partial scans and unavailable services as evidence gaps.
1. Fetch open and closed issues needed for duplicate analysis, including complete bodies, comments, labels, state, and native `blockedBy` and `blocking` relationships. Search by the underlying problem and evidence, not title similarity alone.
1. Research current primary documentation or upstream source only where it can materially confirm a retained finding or implementation boundary. Record the publication or version context and distinguish sourced fact from inference.
1. Classify each candidate as a **verified finding** or **trend opportunity**. A trend becomes an issue only when current project evidence proves fit and value; popularity, novelty, or a generic best practice is insufficient.
1. Reject candidates that duplicate an existing issue, lack reproducible evidence, exceed the project's likely value, restore intentionally rejected scope, or add unjustified complexity.
1. Draft every retained item with [draft-contract.md](references/draft-contract.md). Explain why it is distinct from every close match and model dependencies as draft-to-draft or draft-to-existing-issue edges.
1. Present the deduplication decisions, ordered drafts, and dependency graph for review. Stop here unless the user explicitly authorizes creation in the displayed repository.

## Phase 2: Authorized creation

1. Require an explicit instruction to create the reviewed draft set and reconfirm the exact `owner/repo`. Authorization to review, draft, implement code, or create a pull request is not authorization to create backlog issues.
1. Refresh repository visibility, labels, matching open and closed issues, and native dependencies immediately before mutation. Stop if the drafts became stale or the required `area/*`, single `priority/p*`, and single `effort/*` labels do not exist.
1. Write each issue body to a temporary file and create issues in deterministic draft order with `gh issue create --repo <owner/repo> --title ... --body-file ... --label ...`. Record draft ID, issue number, node ID, and URL after each success.
1. Create no dependency edges until every issue node exists. If issue creation is partial, stop without deleting successful issues, report created, failed, and unattempted drafts, and leave the graph unapplied.
1. Read current native relationships, then add only missing edges with GitHub's `addBlockedBy` GraphQL mutation from the draft contract. Never encode dependencies only in prose.
1. If dependency creation is partial, stop and report successful, failed, and unattempted edges plus every created issue. Do not hide, delete, or replace the partial graph; a retry must re-read and apply only missing edges.
1. Verify every created issue body, label set, and `blockedBy` and `blocking` relationship from GitHub. Return a mutation receipt; do not claim completion from successful issue creation alone.

## Sensitive and unavailable evidence

- For a public repository, never copy private paths, credentials, customer data, private issue text, or non-public runtime details into drafts. Sanitize the evidence or mark the draft `needs-human` when the finding cannot be justified publicly.
- When GitHub, research sources, credentials, or authorized services are unavailable, keep local drafts, name the missing verification, and perform no write.
- A dirty worktree is review evidence, not permission to stash, clean, stage, format, or commit it.
- Unauthorized writes and partial issue creation are failed mutation boundaries, not reasons to improvise around the authorization or native dependency contract.

## Output contract

Return these sections:

1. **Review evidence**: Strongest proven review boundary and material gaps.
1. **Deduplication**: Candidate-to-existing-issue decisions and distinctness rationale.
1. **Draft backlog**: Ordered drafts with routing and dependency edges.
1. **Authorization gate**: Exact repository and mutations awaiting approval, or the explicit authorization already received.
1. **Mutation receipt**: Only after authorized writes; list created issues, labels, verified native edges, and any partial state.

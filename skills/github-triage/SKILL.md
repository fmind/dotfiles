---
name: github-triage
description: Triage, review, and label GitHub pull requests and issues from the terminal with gh-dash, octo.nvim, and gh. Use when reviewing agent-authored PRs, grooming an issue backlog, or deciding which labels to set.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/github-triage
  created: 2026-08-01
  updated: 2026-08-01
---

# Triage and Review GitHub Work

Agents open most of the pull requests and issues here, so throughput is limited by review, not by authoring. The whole loop stays in the terminal.

## Tools

| Tool        | Role                                                                                |
| ----------- | ----------------------------------------------------------------------------------- |
| `gh-dash`   | Triage queue — cross-repository sections, approve/merge/comment/label inline (`gd`) |
| `octo.nvim` | Review surface — line comments, review threads, submit/approve inside Neovim        |
| `gh`        | Scripting and bulk operations                                                       |
| `jq`        | Structured queue filtering used by the bundled shell helper                         |
| `delta`     | Diff rendering, already wired as the `gh` pager                                     |

Configuration lives in `~/.config/gh-dash/config.yml`. Sections are scoped by identity (`involves:@me`, `review-requested:@me`) rather than by owner, so new repositories and orgs appear without editing the file.

## Triage Loop

1. Run `gd` (`gh-dash`). Work the PR sections left to right — they are ordered by how much they need a human: `Needs review` → `CI red` → `Changes requested` → `Ready to merge` → `Requested from me` → `Drafts`.
1. `CI red` first: a failing agent PR is bounced back with a comment (`c`), never reviewed line by line.
1. On a candidate, press `E` to open it in octo, or `D` for the full diff in `delta`.
1. Label with `L`, approve with `v`, merge with `m`, close with `x`. The approval comment is intentionally not prefilled, so write a real one.
1. Switch views with `s` (PRs / issues / notifications), `?` for the full keymap.

Custom keys added on top of the built-ins, chosen to leave every built-in reachable — in particular `L` (label), `d` (diff), `e` (expand description), and `C`/`Space` (checkout):

| Key | View       | Action                         |
| --- | ---------- | ------------------------------ |
| `E` | PRs/issues | Open in octo.nvim over the API |
| `D` | PRs        | Full diff through `delta`      |
| `O` | PRs        | `lazygit` in the local clone   |

## Review in octo.nvim

`E` passes the PR URL to octo, which resolves owner, repository, and number itself — no local clone is needed, which matters because most agent PRs live in repositories that were never cloned.

1. `Octo review start` — enter review mode and open the changed-file panel.
1. Navigate files, add line comments in the diff, and stage them as review comments.
1. `Octo review submit` — choose approve, comment, or request changes.
1. `Octo review discard` — abandon a pending review.
1. Resolve or reply to existing threads directly in the PR buffer.

Use `Octo pr checks` when CI status needs interpretation before approving.

## Label Taxonomy

Labels carry the routing signal. Assignees do not — agents open work under a human account, so `assignee` is nearly always empty and must not be used as a triage input.

| Namespace            | Rule                                                                             |
| -------------------- | -------------------------------------------------------------------------------- |
| `area/*`             | One or more; names the subsystem (`area/infra`, `area/docs`, `area/security`, …) |
| `priority/p0`        | Now — the active frontier and anything blocking it                               |
| `priority/p1`        | Next — near-term once the p0 frontier clears                                     |
| `priority/p2`        | Later — lower urgency within its track                                           |
| `kind/epic`          | Milestone tracker issue that groups other issues                                 |
| `status/in-progress` | Cooperative lease: an agent or human is actively working this issue              |
| `needs-human`        | Blocked on a decision, account, approval, or spend that an agent cannot make     |
| `needs-cluster`      | Requires a live cluster; bring one up, capture evidence, tear it down            |

Rules:

1. Every issue carries at least one `area/*` and exactly one `priority/p*`. Anything missing a priority lands in the `Untriaged` section — that section should trend to empty.
1. Priority is urgency, not phase. A p2 in an active area still outranks nothing; it just waits.
1. Set `status/in-progress` when starting work and remove it when the issue closes or is abandoned, so the lease never goes stale.
1. Set `needs-human` the moment an agent hits a decision, credential, or spend it cannot make, and stop rather than guessing.
1. Reserve `kind/epic` for trackers; epics are not worked directly.

Not every repository defines the full taxonomy — the `area/*` sets are per-repository, and small repositories may only use the GitHub defaults. Check with `gh label list --repo <owner>/<name>` before applying labels, and do not invent labels that do not exist there.

## Agent-Runnable Issue Queue

The queue is a cooperative protocol, not an assignment system. Native GitHub dependencies are authoritative: a blocked issue is never runnable even when its labels otherwise match.

Provision the four shared routing labels idempotently, scoped to the confirmed repository:

```bash
~/.agents/skills/github-triage/scripts/queue.sh setup <owner>/<repo>
```

Every implementation issue must contain these explicit fields:

- **Problem** — the current, verified gap.
- **Evidence** — source, tests, commands, or authorized live observations proving the gap.
- **Acceptance criteria** — machine-checkable or directly verifiable outcomes.
- **Boundaries** — actions and proof levels the issue does not authorize.
- **Validation** — focused checks plus the repository's complete local gate.
- **Routing** — one or more `area/*`, exactly one `priority/p*`, and exactly one `effort/*` label.
- **Dependencies** — native GitHub `blocked-by` and `blocking` relationships, not prose-only links.

The queue helper's offline contract is exercised by [queue_test.sh](tests/queue_test.sh) against [issues.json](tests/fixtures/issues.json).

The underlying live query excludes blocked, claimed, human-gated, and epic issues:

```bash
gh issue list --repo <owner>/<repo> --state open --limit 100 \
  --search '-label:"status/in-progress" -label:"needs-human" -label:"kind/epic" -is:blocked' \
  --json number,title,state,url,labels,blockedBy,comments
```

Use the wrapper for contract validation and deterministic ordering by priority (`p0`, `p1`, `p2`), effort (`s`, `m`, `l`), then issue number:

```bash
~/.agents/skills/github-triage/scripts/queue.sh runnable <owner>/<repo>
```

### Claim, heartbeat, and release

Before claiming, read the complete body, comments, labels, and native dependencies; confirm the issue is returned by the runnable command; and preserve the current worktree. Add `status/in-progress`, then post:

```markdown
**Claim**

- Agent: `<agent identity>`
- Claimed at: `<UTC RFC3339 timestamp>`
- Branch/worktree: `<branch and worktree>`
- Scope: `<bounded implementation scope>`
```

A lease lasts two hours from the latest `**Claim**` or `**Lease update**` comment. Post a `**Lease update**` with current UTC time and concise evidence before that deadline while work continues. A stale label never grants an automatic takeover: first run `queue.sh stale`, inspect the current issue comments, branch/worktree activity, and dependencies, then post a new claim recording the takeover.

On completion, post the acceptance evidence and remove `status/in-progress`. On abandonment, post the exact reason and remove the label. Add `needs-human` only for a concrete human decision, credential, approval, or spend; name that unresolved gate in the release comment.

## Commands

```bash
gh label list --repo <owner>/<name> --limit 100        # discover the repo's taxonomy
gh issue edit <n> --repo <owner>/<name> \
  --add-label "area/infra,priority/p1"                 # label an issue
gh issue list --repo <owner>/<name> \
  --label needs-human --state open                     # what is blocked on a human
~/.agents/skills/github-triage/scripts/queue.sh stale \
  <owner>/<repo>                                       # auditable stale-lease candidates
gh pr review <n> --repo <owner>/<name> \
  --request-changes --body-file <tmpfile>              # bounce a PR from a script
gh pr diff <n> --repo <owner>/<name>                   # already piped through delta
```

## Gotchas

- Pull request and issue filters use GitHub search qualifiers (`is:`, `review:`, `status:`); notification filters are client-side and use `reason:` instead.
- `gh search prs '<qualifiers>'` treats a positional query as keyword text — pass qualifiers as flags, or use `gh api -X GET search/issues -f q='…'` to test a `gh-dash` filter verbatim.
- Local clones are grouped by owner (`~/:owner/:repo` e.g. `~/fmind`, `~/fmind-ai`), allowing `repoPaths` to use a single `:owner/:repo` mapping.
- Never approve on CI red, and never rubber-stamp: the approval comment is deliberately not prefilled.

## Documentation

- [github-pull-request](../github-pull-request/SKILL.md) — authoring the PR body being reviewed.
- [github-repository](../github-repository/SKILL.md) — repository metadata and settings.
- [conventional-commit](../conventional-commit/SKILL.md) — commit taxonomy referenced by PR titles.

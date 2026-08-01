# Backlog Draft Contract

## Draft fields

Give every draft a stable local ID and this complete Markdown body:

```markdown
## Problem

State the current verified gap and its impact.

## Proposal

Describe the smallest complete root-cause solution.

## Acceptance criteria

- [ ] Define directly verifiable outcomes.

## Evidence and references

- Cite repository paths, commands, exact-head CI, authorized runtime observations, or public primary sources.

## Boundaries

Name excluded mutations, proof levels, complexity, spend, and runtime scope.

## Validation

- List focused checks and the complete repository-owned gate.
```

Add routing metadata outside the body:

- One or more existing `area/*` labels.
- Exactly one existing `priority/p0`, `priority/p1`, or `priority/p2` label.
- Exactly one existing `effort/s`, `effort/m`, or `effort/l` label.
- Native dependency edges: `blocked-by` and `blocking`, each targeting a draft ID or existing issue number.
- Deduplication record: matching open and closed issue numbers, the material distinction, and retain or reject decision.
- Evidence class: `verified-finding` or `trend-opportunity`; a retained opportunity must include current project-fit evidence.

## Priority and dependency rules

- Prioritize impact and urgency, not implementation order. A blocker inherits urgency only when it blocks that frontier.
- Keep proposals minimal; reject frameworks, services, automation, or abstractions whose cost exceeds the evidenced problem.
- Make every dependency directional and acyclic. Use native relationships after creation; prose may explain an edge but never substitute for it.
- Create all issue nodes before edges so a partial node-creation run cannot leave a misleading half-graph.

## Native dependency mutation

Resolve each created issue's node ID with `gh issue view <number> --repo <owner/repo> --json id`. For an issue that is blocked by another issue, pass the blocked issue as `issueId` and the prerequisite as `blockingIssueId`:

```bash
gh api graphql \
  -f query='mutation($issueId: ID!, $blockingIssueId: ID!) { addBlockedBy(input: {issueId: $issueId, blockingIssueId: $blockingIssueId}) { issue { number } blockingIssue { number } } }' \
  -f issueId="$blocked_issue_id" \
  -f blockingIssueId="$prerequisite_issue_id"
```

Before every mutation or retry, query `blockedBy` and `blocking`, skip existing edges, and retain a deterministic receipt. After mutation, read both directions back from GitHub and compare them with the reviewed draft graph.

## Partial mutation receipt

Never erase partial state. Report four lists separately:

1. Created and verified issues.
1. Failed and unattempted drafts.
1. Successfully verified dependency edges.
1. Failed and unattempted edges.

Name the exact retry boundary and reconfirm authorization before resuming after material draft or repository changes.

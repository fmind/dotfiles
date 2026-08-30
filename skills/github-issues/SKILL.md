---
name: github-issues
description: "Manage GitHub issues with gh: confirm repo and state, deduplicate, preserve evidence, and verify authorized creation, edits, labels, comments, or closure."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/github-issues
  created: 2026-08-30
  updated: 2026-08-30
---

# Manage GitHub Issues

Use `gh` for evidence-backed issue operations. [project-backlog](../project-backlog/SKILL.md) owns local issue drafts; this skill owns confirmed remote issue state and authorized mutation.

## Workflow

1. **Confirm the target:** Resolve the repository from the explicit URL or `git remote get-url origin`, then state `OWNER/REPO`. Never infer a different repository from a similarly named checkout.
1. **Refresh current state:** Read the issue and relevant neighbors before proposing a change:

   ```bash
   gh issue view <number> -R <owner>/<repo> --json number,title,body,state,stateReason,labels,assignees,milestone,comments,url
   gh issue list -R <owner>/<repo> --state all --search '<distinct terms>' --json number,title,state,url
   ```

1. **Deduplicate and preserve evidence:** Prefer updating an existing issue when it represents the same outcome. Keep reproduction, acceptance criteria, dependencies, decisions, and proof; remove stale activity logs and duplicate checklists.
1. **Check authority:** Reading and drafting are safe defaults. Creation, edits, labels, assignments, comments, closure, reopening, locking, transfer, project changes, and deletion are separate remote mutations requiring user authority.
1. **Apply one bounded mutation:** Prepare substantial bodies in a temporary file and pass it with `--body-file`; avoid shell interpolation and interactive prompts:

   ```bash
   gh issue create -R <owner>/<repo> --title '<title>' --body-file <body-file>
   gh issue edit <number> -R <owner>/<repo> --body-file <body-file>
   ```

1. **Verify from GitHub:** Re-read the issue with `gh issue view --json ...`, compare the intended fields, and report its URL. A successful command alone is not proof of final state.

## Boundaries

- Do not close an issue merely because local code is green; verify the issue's acceptance criteria and requested delivery boundary.
- Do not assign people, notify them in comments, or alter projects and milestones without explicit coordination authority.
- Do not create issues from raw review findings until [project-backlog](../project-backlog/SKILL.md) has deduplicated and prioritized them.
- Use [plan-execution](../plan-execution/SKILL.md) for implementing an accepted issue and [github-pull-request](../github-pull-request/SKILL.md) for the resulting PR.

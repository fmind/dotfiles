---
name: github-issues
description: "Manage GitHub issues with gh: confirm the target, deduplicate, preserve evidence, and verify authorized creation, edits, labels, comments, or closure. Use for GitHub issues."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/github-issues
  created: "2026-08-30"
  updated: "2026-09-03"
---

# GitHub Issues

Read and mutate GitHub issues with `gh` from verified remote state; [project-backlog](../project-backlog/SKILL.md) owns local issue drafts and prioritization.

## Workflow

1. **Confirm the target**: resolve the repository from the explicit URL or `git remote get-url origin` and state `OWNER/REPO`; never infer another repository from a similarly named checkout.
1. **Refresh current state** before proposing a change:

   ```bash
   gh issue view <number> -R <owner>/<repo> --json number,title,body,state,stateReason,labels,assignees,milestone,comments,url
   gh issue list -R <owner>/<repo> --state all --search '<distinct terms>' --json number,title,state,url
   ```

1. **Deduplicate**: update the existing issue that represents the same outcome; keep reproduction, acceptance criteria, dependencies, decisions, and proof; drop stale logs and duplicate checklists.
1. **Apply one bounded mutation** the user authorized. Write substantial bodies to a temporary file and pass `--body-file`; avoid shell interpolation and interactive prompts:

   ```bash
   gh issue create -R <owner>/<repo> --title '<title>' --body-file <body-file>
   gh issue edit <number> -R <owner>/<repo> --body-file <body-file>
   ```

1. **Verify from GitHub**: re-read with `gh issue view --json ...`, compare the intended fields, and report the URL; a zero exit code alone is not proof of final state.

## Gotchas

- **Green is not closed**: verify the issue's acceptance criteria and requested delivery boundary before `gh issue close`; local passing code is not delivery.
- **People and planning fields**: assignments, comment notifications, milestones, and project changes are coordination acts; make them only when the request names them.
- **Raw findings**: route review findings through [project-backlog](../project-backlog/SKILL.md) before creating issues from them.

## Official Skills

Upstream: `cli/cli`. List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
gh skill preview cli/cli gh
gh skill install cli/cli gh
```

## Documentation

- [gh issue manual](https://cli.github.com/manual/gh_issue)
- Companion skills: [project-backlog](../project-backlog/SKILL.md) (drafts and priorities), [plan-execution](../plan-execution/SKILL.md) (implement an issue), [github-pull-request](../github-pull-request/SKILL.md) (the PR).

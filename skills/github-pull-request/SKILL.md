---
name: github-pull-request
description: Create or update a GitHub pull request with a structured What, Why, How, and Test-plan body. Use when opening or updating a PR for the current branch.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/github-pull-request
  created: "2026-06-23"
  updated: "2026-09-03"
---

# GitHub Pull Request

Create or update the pull request for the current branch against `main` with a What, Why, How, and Test plan body; [feature-branch](../feature-branch/SKILL.md) owns branch creation and [conventional-commit](../conventional-commit/SKILL.md) the commits on it.

## Workflow

1. **Stop on `main`**: a PR must come from a feature branch.
1. **Gather context** so the PR reflects the actual work, not only the commit subjects:

   ```bash
   git fetch origin main                                   # refresh origin/main for the ranges below
   git branch --show-current                               # current branch
   git status --short                                      # working tree state
   gh pr view --json number,state,url                      # existing PR for this branch (non-zero exit when none)
   git log --reverse --oneline origin/main..HEAD           # commits since main
   git diff --stat --find-renames origin/main...HEAD       # diff stats
   git diff --name-only --find-renames origin/main...HEAD  # changed files
   ```

1. **Write the title** in imperative mood, under 72 characters.
1. **Write the body** into a temporary file (never inline shell quoting) with these sections:

   ```markdown
   ## What

   ## Why

   ## How

   ## Test plan
   ```

1. **Push the branch** with an upstream when it has none: `git push -u origin "$(git branch --show-current)"`.
1. **Create or update** depending on step 2:

   ```bash
   gh pr edit --base main --title "<title>" --body-file <tmpfile>     # a PR exists
   gh pr create --base main --title "<title>" --body-file <tmpfile>   # no PR yet
   ```

1. **Report** the PR URL, the final title, and the final body; if `origin/main` or GitHub auth is unavailable, explain the blocker and stop.

## Official Skills

Upstream: `cli/cli`. List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
gh skill preview cli/cli gh
gh skill install cli/cli gh
```

## Documentation

- [gh pr manual](https://cli.github.com/manual/gh_pr)
- Companion skills: [feature-branch](../feature-branch/SKILL.md) (branch first), [conventional-commit](../conventional-commit/SKILL.md) (commit cadence), [github-issues](../github-issues/SKILL.md) (the issue the PR closes).

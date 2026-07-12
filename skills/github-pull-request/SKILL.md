---
name: github-pull-request
description: Create or update a GitHub pull request with a structured What / Why / How / Test-plan body. Use when opening or updating a PR for the current branch.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/github-pull-request
  created: 2026-06-23
  updated: 2026-07-09
---

# Create GitHub Pull Request

Create or update a GitHub pull request for the current branch against `main`.

## Context Gathering

Gather context before drafting:

- `git fetch origin main` — refresh `origin/main` so the ranges below are accurate.
- `git config --get remote.origin.url` — origin URL.
- `git branch --show-current` — current branch.
- `git status --short` — working tree state.
- `gh pr view --json number,state,url` — detect an existing PR for this branch (exits non-zero / empty when none, keying the create-vs-edit choice).
- `git log --reverse --oneline origin/main..HEAD` — commits since `main`.
- `git diff --stat --find-renames origin/main...HEAD` — diff stats.
- `git diff --name-only --find-renames origin/main...HEAD` — changed files.

## Workflow

1. If the current branch is `main`, stop and say a PR must come from a feature branch.
1. Inspect changed files as needed so the PR matches the actual work, not just the commit subjects.
1. Write a clear PR title in imperative mood, **under 72 characters**.
1. Write a Markdown body with these sections:

   ```markdown
   ## What

   ## Why

   ## How

   ## Test plan
   ```

1. Push the current branch to `origin` with upstream if needed (`git push -u origin <branch>`).
1. If a PR already exists for this branch, update it:

   ```bash
   gh pr edit --base main --title "<title>" --body-file <tmpfile>
   ```

1. Otherwise create it:

   ```bash
   gh pr create --base main --title "<title>" --body-file <tmpfile>
   ```

1. Use a temporary file for the PR body instead of inline shell quoting.
1. After success, print the PR URL, the final title, and the final body.
1. If `origin/main` or GitHub auth is unavailable, explain the blocker briefly and stop.

## Gotchas

- Keep the final response compact.
- Never force-push to `main`.

## Documentation

- [feature-branch](../feature-branch/SKILL.md) — branch creation upstream of the PR.
- [conventional-commit](../conventional-commit/SKILL.md) — commit cadence on the branch.
